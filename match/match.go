package match

type ISources interface {
	Read(name string)(string,bool)
}
//type IMatch interface {
//	Match(sources ISources)(interface{},bool)
//}

type IReader interface {
	Read(sources interface{})(string,bool)
}


type ReaderCreateFunc func(name string,key string) IReader

