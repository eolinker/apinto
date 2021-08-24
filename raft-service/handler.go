package raft_service

type ICommitHandler interface {
	// CommitHandler 节点commit信息前的处理
	CommitHandler(data []byte) error
}

type IProcessHandler interface {
	// ProcessHandler 节点propose信息前的处理
	ProcessHandler(propose []byte) (string, []byte, error)
}
