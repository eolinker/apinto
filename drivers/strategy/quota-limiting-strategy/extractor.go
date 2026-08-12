package quota_limiting_strategy

import (
	"github.com/eolinker/apinto/utils/response"
	"sort"
	"sync"
	
	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

// sortStrategies 对同级策略按 period 升序、再按 threshold 升序进行原地排序
func sortStrategies(ss []IStrategy) {
	sort.SliceStable(ss, func(i, j int) bool {
		if ss[i].Period() != ss[j].Period() {
			return ss[i].Period() < ss[j].Period()
		}
		return ss[i].Threshold() < ss[j].Threshold()
	})
}

type Period int

func (p Period) String() string {
	switch p {
	case PeriodSecond:
		return "second"
	case PeriodMinute:
		return "minute"
	case PeriodHour:
		return "hour"
	case PeriodDay:
		return "day"
	case PeriodMonth:
		return "month"
	case PeriodTotal:
		return "total"
	default:
		return "second"
	}
}

const (
	PeriodSecond Period = iota
	PeriodMinute
	PeriodHour
	PeriodDay
	PeriodMonth
	PeriodTotal
)

const (
	quotaTypeRequest = iota
	quotaTypeTotalToken
	quotaTypeAmount
)

// QuotaRule 单维度的配额规则
type QuotaRule struct {
	Second int64 `json:"second"` // 每秒配额
	Minute int64 `json:"minute"` // 每分钟配额
	Hour   int64 `json:"hour"`   // 每小时配额
	Day    int64 `json:"day"`    // 每天配额
	Month  int64 `json:"month"`  // 每月配额
	Total  int64 `json:"total"`  // 总配额
}

type IStrategy interface {
	ID() string
	TargetType() string
	Period() Period
	Threshold() int64
	Response() response.IResponse
}

var _ IStrategy = (*Strategy)(nil)

type Strategy struct {
	id         string
	targetType string
	period     Period
	threshold  int64
	response   response.IResponse
}

func (s *Strategy) Response() response.IResponse {
	return s.response
}

func (s *Strategy) TargetType() string {
	return s.targetType
}

func NewStrategies(id string, targetType string, rule QuotaRule, resp response.IResponse) []IStrategy {
	periods := []struct {
		period Period
		val    int64
	}{
		{PeriodSecond, rule.Second},
		{PeriodMinute, rule.Minute},
		{PeriodHour, rule.Hour},
		{PeriodDay, rule.Day},
		{PeriodMonth, rule.Month},
		{PeriodTotal, rule.Total},
	}
	
	result := make([]IStrategy, 0, len(periods))
	for _, p := range periods {
		if p.val > 0 {
			result = append(result, &Strategy{
				id:         id,
				targetType: targetType,
				period:     p.period,
				threshold:  p.val,
				response:   resp,
			})
		}
	}
	return result
}

func (s *Strategy) ID() string {
	return s.id
}

func (s *Strategy) Period() Period {
	return s.period
}

func (s *Strategy) Threshold() int64 {
	return s.threshold
}

type IExtractor interface {
	GetStrategies(ctx eocontext.EoContext, quotaType ...int) ([]ITenantStrategy, bool)
	GetParentStrategies(ctx eocontext.EoContext, quotaType ...int) ([]ITenantStrategy, bool)
}

type ITenantStrategy interface {
	Tenant() string
	Strategies() []IStrategy
}

type tenantStrategy struct {
	tenant     string
	strategies []IStrategy
}

func (t *tenantStrategy) Tenant() string {
	return t.tenant
}

func (t *tenantStrategy) Strategies() []IStrategy {
	return t.strategies
}

// Extractor 支持根据策略ID进行增删改查以及多维树全量DFS匹配的接口
type Extractor interface {
	IExtractor
	AddStrategy(id string, config *Config)
	RemoveStrategy(id string)
	GetStrategy(id string) (*Config, bool)
}

// GenericDimensionTree 通用多维索引树，支持深层 N 维决策索引
type GenericDimensionTree[T any] struct {
	root *GenericTreeNode[T]
}

type GenericTreeNode[T any] struct {
	children map[string]*GenericTreeNode[T]
	items    []T
}

func NewGenericDimensionTree[T any]() *GenericDimensionTree[T] {
	return &GenericDimensionTree[T]{
		root: &GenericTreeNode[T]{
			children: make(map[string]*GenericTreeNode[T]),
		},
	}
}

func (t *GenericDimensionTree[T]) Insert(keys []string, item T) {
	curr := t.root
	for _, key := range keys {
		if curr.children == nil {
			curr.children = make(map[string]*GenericTreeNode[T])
		}
		child, ok := curr.children[key]
		if !ok {
			child = &GenericTreeNode[T]{children: make(map[string]*GenericTreeNode[T])}
			curr.children[key] = child
		}
		curr = child
	}
	curr.items = append(curr.items, item)
}

func (t *GenericDimensionTree[T]) Remove(keys []string, matchFunc func(item T) bool) {
	curr := t.root
	for _, key := range keys {
		if curr.children == nil {
			return
		}
		child, ok := curr.children[key]
		if !ok {
			return
		}
		curr = child
	}
	
	newItems := make([]T, 0, len(curr.items))
	for _, item := range curr.items {
		if !matchFunc(item) {
			newItems = append(newItems, item)
		}
	}
	curr.items = newItems
}

func (t *GenericDimensionTree[T]) Search(queryKeys [][]string) []T {
	var results []T
	var dfs func(node *GenericTreeNode[T], depth int)
	dfs = func(node *GenericTreeNode[T], depth int) {
		if node == nil {
			return
		}
		if depth == len(queryKeys) {
			results = append(results, node.items...)
			return
		}
		for _, key := range queryKeys[depth] {
			if child, exists := node.children[key]; exists {
				dfs(child, depth+1)
			}
		}
	}
	dfs(t.root, 0)
	return results
}

// IndexedStrategy 包含策略主体、维度路径及转换后的算法 Strategy 列表
type IndexedStrategy struct {
	ID             string
	Config         *Config
	Strategies     []IStrategy
	DimensionPaths [][]string // 策略映射到多维树的所有索引路径（用于 Remove 精确清理）
}

type strategyExtractor struct {
	mu         sync.RWMutex
	quotaType  int                                     // 额度类型: quotaTypeRequest / quotaTypeTotalToken / quotaTypeAmount
	strategies map[string]*IndexedStrategy             // 策略ID -> IndexedStrategy 主表
	tree       *GenericDimensionTree[*IndexedStrategy] // 多维树索引
}

func newSingleExtractor(qType int) Extractor {
	return &strategyExtractor{
		quotaType:  qType,
		strategies: make(map[string]*IndexedStrategy),
		tree:       NewGenericDimensionTree[*IndexedStrategy](),
	}
}

// multiQuotaExtractor 组合三个配额类型的 strategyExtractor
type multiQuotaExtractor struct {
	requestExtractor    Extractor
	totalTokenExtractor Extractor
	amountExtractor     Extractor
}

// NewExtractor 创建按 Quota 类型管理的综合提取器实例
func NewExtractor() Extractor {
	
	return &multiQuotaExtractor{
		requestExtractor:    newSingleExtractor(quotaTypeRequest),
		totalTokenExtractor: newSingleExtractor(quotaTypeTotalToken),
		amountExtractor:     newSingleExtractor(quotaTypeAmount),
	}
}

func (m *multiQuotaExtractor) GetExtractor(quotaType int) Extractor {
	switch quotaType {
	case quotaTypeRequest:
		return m.requestExtractor
	case quotaTypeTotalToken:
		return m.totalTokenExtractor
	case quotaTypeAmount:
		return m.amountExtractor
	default:
		return m.requestExtractor
	}
}

func (m *multiQuotaExtractor) AddStrategy(id string, config *Config) {
	m.requestExtractor.AddStrategy(id, config)
	m.totalTokenExtractor.AddStrategy(id, config)
	m.amountExtractor.AddStrategy(id, config)
}

func (m *multiQuotaExtractor) RemoveStrategy(id string) {
	m.requestExtractor.RemoveStrategy(id)
	m.totalTokenExtractor.RemoveStrategy(id)
	m.amountExtractor.RemoveStrategy(id)
}

func (m *multiQuotaExtractor) GetStrategy(id string) (*Config, bool) {
	if cfg, ok := m.requestExtractor.GetStrategy(id); ok {
		return cfg, true
	}
	if cfg, ok := m.totalTokenExtractor.GetStrategy(id); ok {
		return cfg, true
	}
	return m.amountExtractor.GetStrategy(id)
}

func (m *multiQuotaExtractor) GetStrategies(ctx eocontext.EoContext, quotaType ...int) ([]ITenantStrategy, bool) {
	if len(quotaType) == 0 {
		return m.GetExtractor(quotaTypeRequest).GetStrategies(ctx)
	}
	
	allStrategies := make([]ITenantStrategy, 0, 10*len(quotaType))
	for _, t := range quotaType {
		ss, ok := m.GetExtractor(t).GetStrategies(ctx)
		if ok {
			allStrategies = append(allStrategies, ss...)
		}
	}
	
	return allStrategies, len(allStrategies) > 0
}

func (m *multiQuotaExtractor) GetParentStrategies(ctx eocontext.EoContext, quotaType ...int) ([]ITenantStrategy, bool) {
	if len(quotaType) > 0 {
		return m.GetExtractor(quotaType[0]).GetParentStrategies(ctx)
	}
	
	var allStrategies []ITenantStrategy
	reqStr, ok1 := m.requestExtractor.GetParentStrategies(ctx)
	if ok1 {
		allStrategies = append(allStrategies, reqStr...)
	}
	tokStr, ok2 := m.totalTokenExtractor.GetParentStrategies(ctx)
	if ok2 {
		allStrategies = append(allStrategies, tokStr...)
	}
	amtStr, ok3 := m.amountExtractor.GetParentStrategies(ctx)
	if ok3 {
		allStrategies = append(allStrategies, amtStr...)
	}
	return allStrategies, len(allStrategies) > 0
}

// GetTenantChain 根据给定租户 ID 从 ICustomerVar 中向上查找层级，返回从当前租户至一级根租户的切片
func GetTenantChain(tenant string, cv eosc.ICustomerVar) []string {
	if tenant == "" {
		return nil
	}
	
	chain := make([]string, 0, 10)
	if cv == nil {
		return chain
	}
	
	visited := map[string]bool{tenant: true}
	curr := tenant
	
	for i := 0; i < 50; i++ {
		parentMap, has := cv.GetAll("parent:" + curr)
		if !has || len(parentMap) == 0 {
			break
		}
		
		var parentID string
		for p := range parentMap {
			if p != "" {
				parentID = p
				break
			}
		}
		
		if parentID == "" || visited[parentID] {
			break
		}
		
		visited[parentID] = true
		chain = append(chain, parentID)
		curr = parentID
	}
	
	return chain
}

func normalizeKeys(keys []string) []string {
	seen := make(map[string]bool)
	res := make([]string, 0, len(keys))
	for _, k := range keys {
		if !seen[k] {
			seen[k] = true
			res = append(res, k)
		}
	}
	return res
}

// generateDimensionPaths 计算策略包含的所有维度匹配路径
// 维度顺序: Tenant -> TargetType -> TargetItem -> ResType -> ResParent -> ResItem
func generateDimensionPaths(filter FiltersConfig) [][]string {
	// Depth 0: Tenant (不允许 all，必须要有指定的 tenant)
	tenant := filter.Tenant
	if tenant == "" || tenant == "all" || tenant == "*" {
		return nil
	}
	tKeys := []string{tenant}
	if filter.Target.Type == "channel" {
		if len(filter.Target.Items) == 0 {
			return nil
		}
		tKeys = filter.Target.Items
	}
	
	// Depth 1: TargetType
	var targetTypeKeys []string
	if filter.Target.Type == "channel" {
		targetTypeKeys = []string{"all"}
	} else {
		targetTypeKeys = normalizeKeys([]string{filter.Target.Type})
	}
	
	// Depth 2: TargetItem
	var targetItemKeys []string
	if filter.Target.All {
		targetItemKeys = []string{"all"}
	} else {
		targetItemKeys = normalizeKeys(filter.Target.Items)
	}
	
	// Depth 3: ResourceType
	var resTypeKeys []string
	
	resTypeKeys = normalizeKeys([]string{filter.Resource.Type})
	
	// Depth 4: ResourceParent
	var resParentKeys []string
	if filter.Resource.All {
		resParentKeys = []string{"all"}
	} else {
		resParentKeys = normalizeKeys(filter.Resource.Parents)
	}
	
	// Depth 5: ResourceItem
	var resItemKeys []string
	if filter.Resource.All || len(filter.Resource.Parents) > 0 {
		resItemKeys = []string{"all"}
	} else {
		resItemKeys = normalizeKeys(filter.Resource.Items)
	}
	
	// 组合 6 维笛卡尔积路径 (Tenant -> TargetType -> TargetItem -> ResType -> ResParent -> ResItem)
	paths := make([][]string, 0, len(tKeys)*len(targetTypeKeys)*len(targetItemKeys)*len(resTypeKeys)*len(resParentKeys)*len(resItemKeys))
	for _, t := range tKeys {
		for _, tt := range targetTypeKeys {
			for _, ti := range targetItemKeys {
				for _, rt := range resTypeKeys {
					for _, rp := range resParentKeys {
						for _, ri := range resItemKeys {
							paths = append(paths, []string{t, tt, ti, rt, rp, ri})
						}
					}
				}
			}
		}
	}
	return paths
}

func (e *strategyExtractor) AddStrategy(id string, config *Config) {
	if id == "" || config == nil {
		return
	}
	
	e.mu.Lock()
	defer e.mu.Unlock()
	
	// 1. 若存在旧策略，从多维树中彻底清理旧路径节点
	if old, exists := e.strategies[id]; exists {
		for _, path := range old.DimensionPaths {
			e.tree.Remove(path, func(item *IndexedStrategy) bool {
				return item.ID == id
			})
		}
	}
	
	// 2. 计算策略映射到 6 维多维树的所有索引路径（要求必须有指定的 Tenant）
	dimPaths := generateDimensionPaths(config.Filters)
	if len(dimPaths) == 0 {
		return
	}
	
	// 3. 根据当前 extractor 的 quotaType 提取对应的 QuotaRule
	var rule QuotaRule
	switch e.quotaType {
	case quotaTypeRequest:
		rule = config.Quota.Request
	case quotaTypeTotalToken:
		rule = config.Quota.TotalToken
	case quotaTypeAmount:
		rule = config.Quota.Amount
	}
	
	strategies := NewStrategies(id, config.Filters.Target.Type, rule, response.Parse(config.Response))
	if len(strategies) == 0 {
		delete(e.strategies, id)
		return
	}
	
	indexed := &IndexedStrategy{
		ID:             id,
		Config:         config,
		Strategies:     strategies,
		DimensionPaths: dimPaths,
	}
	
	// 4. 存入 ID 主表并将所有路径节点插入多维树
	e.strategies[id] = indexed
	for _, path := range dimPaths {
		e.tree.Insert(path, indexed)
	}
}

func (e *strategyExtractor) RemoveStrategy(id string) {
	if id == "" {
		return
	}
	
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if indexed, exists := e.strategies[id]; exists {
		for _, path := range indexed.DimensionPaths {
			e.tree.Remove(path, func(item *IndexedStrategy) bool {
				return item.ID == id
			})
		}
		delete(e.strategies, id)
	}
}

func (e *strategyExtractor) GetStrategy(id string) (*Config, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	if indexed, exists := e.strategies[id]; exists {
		return indexed.Config, true
	}
	return nil, false
}

func (e *strategyExtractor) GetStrategies(ctx eocontext.EoContext, quotaType ...int) ([]ITenantStrategy, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// 1. 从 Context 获取 6 个维度的特征
	tenant := ctx.GetLabel("tenant")
	
	// Depth 0: Tenant (必须要有指定的 tenant，不允许 all)
	if tenant == "" {
		return nil, false
	}
	
	resourceType := ctx.GetLabel("resource_type")
	resource := ctx.GetLabel("resource")
	provider := ctx.GetLabel("provider")
	consumer := ctx.GetLabel("consumer")
	consumerType := context_label.GetConsumerType(ctx)
	
	// 2. 构造 6 维 Depth 查询组合 Candidate Keys
	// 维度顺序: Tenant -> TargetType -> TargetItem -> ResType -> ResParent -> ResItem
	
	// Depth 0: Tenant (仅精确指定 tenant)
	dim0Tenants := []string{tenant}
	
	// Depth 1: TargetType
	dim1TargetTypes := []string{"all"}
	if consumerType != "" {
		cType := context_label.GetConsumerType(ctx)
		if cType != "" {
			dim1TargetTypes = append(dim1TargetTypes, string(cType))
		}
		if context_label.IsUserConsumer(ctx) {
			dim1TargetTypes = append(dim1TargetTypes, "user_of_resource_group")
		}
	}
	
	// Depth 2: TargetItem
	dim2TargetItems := []string{"all"}
	if consumer != "" {
		dim2TargetItems = append(dim2TargetItems, consumer)
	}
	
	// Depth 3: ResourceType
	dim3ResTypes := []string{"all"}
	if resourceType != "" {
		dim3ResTypes = append(dim3ResTypes, resourceType)
	}
	
	// Depth 4: ResourceParent
	dim4ResParents := []string{"all"}
	if provider != "" {
		dim4ResParents = append(dim4ResParents, provider)
	}
	
	// Depth 5: ResourceItem
	dim5ResItems := []string{"all"}
	if resource != "" {
		dim5ResItems = append(dim5ResItems, resource)
	}
	
	queryKeys := [][]string{dim0Tenants, dim1TargetTypes, dim2TargetItems, dim3ResTypes, dim4ResParents, dim5ResItems}
	
	// 3. 一步通过多维树全量 DFS 搜索直接精准定位目标策略
	matchedList := e.tree.Search(queryKeys)
	if len(matchedList) == 0 {
		return nil, false
	}
	
	// 4. 去重收集符合条件的 Strategy
	visited := make(map[string]bool, len(matchedList))
	result := make([]IStrategy, 0, 10)
	for _, cand := range matchedList {
		if !visited[cand.ID] {
			visited[cand.ID] = true
			result = append(result, cand.Strategies...)
		}
	}
	
	// 5. 同级策略按 period 升序，再按 threshold 升序排序
	sortStrategies(result)
	
	return []ITenantStrategy{
		&tenantStrategy{
			tenant:     tenant,
			strategies: result,
		},
	}, len(result) > 0
}

func (e *strategyExtractor) GetParentStrategies(ctx eocontext.EoContext, quotaType ...int) ([]ITenantStrategy, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// 1. 从 Context 获取初始父租户 ID
	tenant := ctx.GetLabel("tenant")
	
	if tenant == "" {
		return nil, false
	}
	
	// 2. 递归向上获取所有父租户链 (从直接父租户按顺序递归至一级根父租户)
	parentChain := GetTenantChain(tenant, customerVar)
	if len(parentChain) == 0 {
		return nil, false
	}
	
	resourceType := ctx.GetLabel("resource_type")
	resource := ctx.GetLabel("resource")
	provider := ctx.GetLabel("provider")
	
	// 3. 构造 6 维 Depth 查询组合 Candidate Keys
	// 维度顺序: Tenant -> TargetType -> TargetItem -> ResType -> ResParent -> ResItem
	
	// Depth 1: TargetType (父租户策略只有 TargetType 为 all 的情况)
	dim1TargetTypes := []string{"all"}
	
	// Depth 2: TargetItem
	dim2TargetItems := []string{"all"}
	
	// Depth 3: ResourceType
	dim3ResTypes := []string{"all"}
	if resourceType != "" {
		dim3ResTypes = append(dim3ResTypes, resourceType)
	}
	
	// Depth 4: ResourceParent
	dim4ResParents := []string{"all"}
	if provider != "" {
		dim4ResParents = append(dim4ResParents, provider)
	}
	
	// Depth 5: ResourceItem
	dim5ResItems := []string{"all"}
	if resource != "" {
		dim5ResItems = append(dim5ResItems, resource)
	}
	
	visited := make(map[string]bool)
	result := make([]IStrategy, 0, 10)
	
	// 4. 沿父租户链从上往下（根租户到直接父租户）逐层检索匹配策略
	//    GetTenantChain 返回顺序为从直接父租户到根租户，此处需反转为从根到直接父
	dim0Tenants := make([]string, 0, len(parentChain))
	for i := len(parentChain) - 1; i >= 0; i-- {
		dim0Tenants = append(dim0Tenants, parentChain[i])
	}
	
	// 按层级（父租户）逐层查询，保证父级从上往下的顺序，
	// 同层级内按 period 升序、再按 threshold 升序排序
	for _, parentTenant := range dim0Tenants {
		queryKeys := [][]string{{parentTenant}, dim1TargetTypes, dim2TargetItems, dim3ResTypes, dim4ResParents, dim5ResItems}
		matchedList := e.tree.Search(queryKeys)
		if len(matchedList) == 0 {
			continue
		}
		levelResult := make([]IStrategy, 0, len(matchedList))
		for _, cand := range matchedList {
			if !visited[cand.ID] {
				visited[cand.ID] = true
				levelResult = append(levelResult, cand.Strategies...)
			}
		}
		sortStrategies(levelResult)
		result = append(result, levelResult...)
	}
	
	return []ITenantStrategy{
		&tenantStrategy{
			tenant:     tenant,
			strategies: result,
		},
	}, len(result) > 0
}

var (
	extractorManager = NewExtractor()
)

func GetRequestStrategies(ctx eocontext.EoContext) ([]ITenantStrategy, bool) {
	return GetStrategies(ctx, quotaTypeRequest)
}

func GetTotalTokenStrategies(ctx eocontext.EoContext) ([]ITenantStrategy, bool) {
	return GetStrategies(ctx, quotaTypeTotalToken)
}

func GetAmountStrategies(ctx eocontext.EoContext) ([]ITenantStrategy, bool) {
	return GetStrategies(ctx, quotaTypeAmount)
}

func GetStrategies(ctx eocontext.EoContext, quotaType ...int) ([]ITenantStrategy, bool) {
	result := make([]ITenantStrategy, 0, 20)
	// 先返回父级策略（从上往下，根租户在前）
	ss, has := extractorManager.GetParentStrategies(ctx, quotaType...)
	if has {
		result = append(result, ss...)
	}
	// 再返回当前级别策略（同级按 period 升序、再按 threshold 升序）
	ss, has = extractorManager.GetStrategies(ctx, quotaType...)
	if has {
		result = append(result, ss...)
	}
	return result, len(result) > 0
}

func addStrategy(id string, config *Config) {
	extractorManager.AddStrategy(id, config)
}

func removeStrategy(id string) {
	extractorManager.RemoveStrategy(id)
}
