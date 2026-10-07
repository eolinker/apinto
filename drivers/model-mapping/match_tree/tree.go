package match_tree

// GenericTreeNode 泛型多维前缀树节点
type GenericTreeNode[T any] struct {
	children map[string]*GenericTreeNode[T]
	items    []T
}

// GenericDimensionTree 通用多维前缀树，支持深层 N 维决策索引
type GenericDimensionTree[T any] struct {
	root *GenericTreeNode[T]
}

// NewGenericDimensionTree 创建多维前缀树
func NewGenericDimensionTree[T any]() *GenericDimensionTree[T] {
	return &GenericDimensionTree[T]{
		root: &GenericTreeNode[T]{
			children: make(map[string]*GenericTreeNode[T]),
		},
	}
}

// Insert 向指定的多维路径叶子节点插入数据项
func (t *GenericDimensionTree[T]) Insert(keys []string, item T) {
	curr := t.root
	for _, key := range keys {
		if curr.children == nil {
			curr.children = make(map[string]*GenericTreeNode[T])
		}
		child, ok := curr.children[key]
		if !ok {
			child = &GenericTreeNode[T]{
				children: make(map[string]*GenericTreeNode[T]),
			}
			curr.children[key] = child
		}
		curr = child
	}
	curr.items = append(curr.items, item)
}

// Remove 精确移除指定路径下满足 matchFunc 的数据项
func (t *GenericDimensionTree[T]) Remove(keys []string, matchFunc func(item T) bool) int {
	curr := t.root
	for _, key := range keys {
		if curr.children == nil {
			return -1
		}
		child, ok := curr.children[key]
		if !ok {
			return -1
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
	return len(curr.items)
}

// SearchResult 搜索结果：包含命中的路径与对应叶子节点挂载的数据
type SearchResult[T any] struct {
	Path  []string
	Items []T
}

// Search 根据各维度的候选 key 进行全量 DFS 搜索（无兜底模式下 queryKeys[depth] 仅含单一精确值）
func (t *GenericDimensionTree[T]) Search(queryKeys [][]string) []SearchResult[T] {
	var results []SearchResult[T]

	var dfs func(node *GenericTreeNode[T], depth int, path []string)
	dfs = func(node *GenericTreeNode[T], depth int, path []string) {
		if node == nil {
			return
		}

		// 到达最深层，记录当前组合路径和 items
		if depth == len(queryKeys) {
			if len(node.items) > 0 {
				p := make([]string, len(path))
				copy(p, path)
				it := make([]T, len(node.items))
				copy(it, node.items)
				results = append(results, SearchResult[T]{
					Path:  p,
					Items: it,
				})
			}
			return
		}

		// 遍历当前深度的候选 key
		for _, key := range queryKeys[depth] {
			if child, exists := node.children[key]; exists {
				dfs(child, depth+1, append(path, key))
			}
		}
	}

	dfs(t.root, 0, nil)
	return results
}
