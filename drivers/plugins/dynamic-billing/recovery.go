package dynamic_billing

import (
	"context"
	"sync"
	"time"

	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/eosc/log"
)

// RecoveryTask 描述了冲正失败后需要重试的任务元数据
type RecoveryTask struct {
	Cache       resources.ICache
	CreditKey   string
	BalanceKey  string
	LockKey     string
	DeductType  int
	CreditID    string
	EstCost     float64
	UserID      string
	Retries     int
	MaxRetries  int
	NextRetryAt time.Time
}

// recoveryManager 单例，用于维护本地冲正重试队列（指数退避机制）
type recoveryManager struct {
	mu     sync.Mutex
	tasks  []*RecoveryTask
	wakeCh chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	defaultRecoveryManager = newRecoveryManager()
)

func newRecoveryManager() *recoveryManager {
	ctx, cancel := context.WithCancel(context.Background())
	rm := &recoveryManager{
		wakeCh: make(chan struct{}, 1),
		ctx:    ctx,
		cancel: cancel,
	}
	go rm.loop()
	return rm
}

func (rm *recoveryManager) Enqueue(task *RecoveryTask) {
	if task == nil || task.Cache == nil || task.EstCost <= 0 {
		return
	}
	if task.MaxRetries <= 0 {
		task.MaxRetries = 5
	}
	task.NextRetryAt = time.Now().Add(time.Second) // 初始延迟 1 秒

	rm.mu.Lock()
	rm.tasks = append(rm.tasks, task)
	rm.mu.Unlock()

	select {
	case rm.wakeCh <- struct{}{}:
	default:
	}
}

func (rm *recoveryManager) loop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
		case <-rm.wakeCh:
		}

		rm.process()
	}
}

func (rm *recoveryManager) process() {
	now := time.Now()
	rm.mu.Lock()
	if len(rm.tasks) == 0 {
		rm.mu.Unlock()
		return
	}

	var ready []*RecoveryTask
	var pending []*RecoveryTask

	for _, t := range rm.tasks {
		if !now.Before(t.NextRetryAt) {
			ready = append(ready, t)
		} else {
			pending = append(pending, t)
		}
	}
	rm.tasks = pending
	rm.mu.Unlock()

	for _, t := range ready {
		ok := executeRecovery(context.Background(), t.Cache, t.CreditKey, t.BalanceKey, t.LockKey, t.DeductType, t.CreditID, t.EstCost, t.UserID)
		if ok {
			log.Infof("[dynamic-billing] recovery task succeeded on retry #%d for user %s, credit_id=%s, amount=%f",
				t.Retries+1, t.UserID, t.CreditID, t.EstCost)
			continue
		}

		t.Retries++
		if t.Retries >= t.MaxRetries {
			log.Errorf("[dynamic-billing] CRITICAL: recovery retry exhausted (%d times) for user %s, credit_id=%s, amount=%f! Manual reconciliation required.",
				t.Retries, t.UserID, t.CreditID, t.EstCost)
			continue
		}

		// 指数退避：1s, 2s, 4s, 8s, 16s
		delay := time.Second * time.Duration(1<<t.Retries)
		t.NextRetryAt = time.Now().Add(delay)

		rm.mu.Lock()
		rm.tasks = append(rm.tasks, t)
		rm.mu.Unlock()

		log.Warnf("[dynamic-billing] recovery failed for user %s, scheduled retry #%d after %v",
			t.UserID, t.Retries+1, delay)
	}
}
