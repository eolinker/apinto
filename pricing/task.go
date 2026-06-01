package pricing

import (
	"encoding/json"
	"sync"
	"time"
)

// TaskStatus 异步计费任务状态。
type TaskStatus string

const (
	TaskStatusPending TaskStatus = "pending"
	TaskStatusBilled  TaskStatus = "billed"
	TaskStatusFailed  TaskStatus = "failed"
)

// DefaultTaskTTL 异步任务缓存默认过期时间（24 小时）。
const DefaultTaskTTL = 24 * time.Hour

// TaskEntry 异步计费任务条目，用于关联 submit/query 两次独立请求并防重复计费。
type TaskEntry struct {
	TaskID    string                 `json:"task_id"`
	Provider  string                 `json:"provider"`
	Resource  string                 `json:"resource"`
	Phase     PricingPhase           `json:"phase"`
	PreResult *CalculateResult       `json:"pre_result,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Status    TaskStatus             `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	BilledAt  time.Time              `json:"billed_at,omitempty"`

	// 余额扣减相关
	FreezeID      string  `json:"freeze_id,omitempty"`
	EstimatedCost float64 `json:"estimated_cost,omitempty"`
	ActualCost    float64 `json:"actual_cost,omitempty"`
	AccountID     string  `json:"account_id,omitempty"`
	SettleStatus  string  `json:"settle_status,omitempty"`
}

// ITaskExecutor 异步计费任务执行器接口。
type ITaskExecutor interface {
	Submit(taskID, provider, resource string, phase PricingPhase, preResult *CalculateResult, fields map[string]interface{}) error
	Get(taskID string) (*TaskEntry, bool)
	IsBilled(taskID string) bool
	MarkBilled(taskID string) bool
	MarkFailed(taskID string) bool
	Remove(taskID string)
	Stop()
}

// LocalTaskExecutor 进程内本地内存实现，作为分布式实现的回退层。
// 后台每 5 分钟清理过期任务，停止时通过 stopCh 退出。
type LocalTaskExecutor struct {
	mu     sync.RWMutex
	tasks  map[string]*TaskEntry
	ttl    time.Duration
	stopCh chan struct{}
}

func NewLocalTaskExecutor(ttl time.Duration) *LocalTaskExecutor {
	if ttl <= 0 {
		ttl = DefaultTaskTTL
	}
	m := &LocalTaskExecutor{
		tasks:  make(map[string]*TaskEntry),
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}
	go m.cleanupLoop()
	return m
}

func (m *LocalTaskExecutor) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.cleanExpired()
		case <-m.stopCh:
			return
		}
	}
}

func (m *LocalTaskExecutor) cleanExpired() {
	m.mu.Lock()
	now := time.Now()
	for id, entry := range m.tasks {
		if now.Sub(entry.CreatedAt) > m.ttl {
			delete(m.tasks, id)
		}
	}
	m.mu.Unlock()
}

// Submit 提交任务，taskID 已存在时幂等忽略。
func (m *LocalTaskExecutor) Submit(taskID, provider, resource string, phase PricingPhase, preResult *CalculateResult, fields map[string]interface{}) error {
	m.mu.Lock()
	if _, exists := m.tasks[taskID]; exists {
		m.mu.Unlock()
		return nil
	}
	m.tasks[taskID] = &TaskEntry{
		TaskID:    taskID,
		Provider:  provider,
		Resource:  resource,
		Phase:     phase,
		PreResult: preResult,
		Fields:    fields,
		Status:    TaskStatusPending,
		CreatedAt: time.Now(),
	}
	m.mu.Unlock()
	return nil
}

func (m *LocalTaskExecutor) Get(taskID string) (*TaskEntry, bool) {
	m.mu.RLock()
	entry, ok := m.tasks[taskID]
	m.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return entry, true
}

func (m *LocalTaskExecutor) IsBilled(taskID string) bool {
	m.mu.RLock()
	entry, ok := m.tasks[taskID]
	m.mu.RUnlock()
	if !ok {
		return false
	}
	return entry.Status == TaskStatusBilled
}

func (m *LocalTaskExecutor) MarkBilled(taskID string) bool {
	m.mu.Lock()
	entry, ok := m.tasks[taskID]
	if !ok {
		m.mu.Unlock()
		return false
	}
	if entry.Status == TaskStatusBilled {
		m.mu.Unlock()
		return false
	}
	entry.Status = TaskStatusBilled
	entry.BilledAt = time.Now()
	m.mu.Unlock()
	return true
}

func (m *LocalTaskExecutor) MarkFailed(taskID string) bool {
	m.mu.Lock()
	entry, ok := m.tasks[taskID]
	if !ok {
		m.mu.Unlock()
		return false
	}
	entry.Status = TaskStatusFailed
	m.mu.Unlock()
	return true
}

func (m *LocalTaskExecutor) Remove(taskID string) {
	m.mu.Lock()
	delete(m.tasks, taskID)
	m.mu.Unlock()
}

// Stop 通过非阻塞写入触发后台清理协程退出。
// 多次调用安全：通道阻塞时直接返回，避免 panic。
func (m *LocalTaskExecutor) Stop() {
	select {
	case m.stopCh <- struct{}{}:
	default:
	}
}

// ExtractTaskID 从字段集合中按优先级提取异步任务 ID。
func ExtractTaskID(fields map[string]interface{}) string {
	for _, key := range []string{"task_id", "id", "request_id"} {
		if v, ok := fields[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

// MarshalTaskEntry / UnmarshalTaskEntry 用于将任务条目在 Redis 等远端存储中序列化。
func MarshalTaskEntry(entry *TaskEntry) ([]byte, error) {
	return json.Marshal(entry)
}

func UnmarshalTaskEntry(data []byte) (*TaskEntry, error) {
	entry := &TaskEntry{}
	if err := json.Unmarshal(data, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// TaskCacheKey 生成 Redis 中存储任务条目的键名（与旧版隔离，使用 billing 前缀）。
func TaskCacheKey(taskID string) string {
	return "apinto:billing:task:" + taskID
}

// TaskBilledKey 生成 Redis 中存储已计费标记的键名（与旧版隔离，使用 billing 前缀）。
func TaskBilledKey(taskID string) string {
	return "apinto:billing:task-billed:" + taskID
}
