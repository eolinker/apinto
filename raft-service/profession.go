package raft_service

type Profession struct {
}

func (p *Profession) ProcessHandler(propose []byte) (string, []byte, error) {
	panic("implement me")
}

func (p *Profession) CommitHandler(data []byte) error {
	panic("implement me")
}
