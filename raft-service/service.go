package raft_service

import (
	"encoding/json"
	"errors"

	"github.com/eolinker/eosc"
)

var (
	ErrInvalidNamespace     = errors.New("invalid namespace")
	ErrInvalidCommitHandler = errors.New("invalid commit handler")
)

type Service struct {
	snapshots       eosc.IUntyped
	commitHandlers  eosc.IUntyped
	processHandlers eosc.IUntyped
}

func (s *Service) CommitHandler(namespace string, data []byte) error {
	v, has := s.commitHandlers.Get(namespace)
	if !has {
		return ErrInvalidNamespace
	}
	f, ok := v.(ICommitHandler)
	if !ok {
		return ErrInvalidCommitHandler
	}
	return f.CommitHandler(data)
}

func (s *Service) ProcessHandler(namespace string, propose []byte) (cmd string, data []byte, err error) {
	panic("implement me")
}

func (s *Service) GetInit() (cmd string, data []byte, err error) {
	panic("implement me")
}

func (s *Service) ResetSnap(data []byte) error {
	var snaps map[string]interface{}

	err := json.Unmarshal(data, &snaps)
	if err != nil {
		return err
	}
	snapshots := eosc.NewUntyped()
	for key, value := range snaps {
		snapshots.Set(key, value)
	}
	s.snapshots = snapshots
	return nil
}

func (s *Service) GetSnapshot() ([]byte, error) {
	return json.Marshal(s.snapshots.All())
}
