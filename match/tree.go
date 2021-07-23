package match

type IMatcherNode interface {
	Match(interface{})(interface{},bool)
}
type IMatcher interface {
	Match(interface{})(target interface{},has bool)
}
type Item struct {
	values map[string]IMatcherNode

	endpoint string
}

type Node struct {
	Key string

	endpoint interface{}
}

type Pattern struct {
	CMD string
	Value string
}
type Rule struct {
	pattern []Pattern
	target interface{}
}

type ReaderFactory interface {
	Reader(v string)(IReader,error)
}

func Parse(rules []Rule,factory ReaderFactory) (IMatcher,error){
	return nil,nil
}

