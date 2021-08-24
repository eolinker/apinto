package raft_service

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"time"

	"github.com/eolinker/eosc"
)

var (
	errInvalidKey = errors.New("invalid key")
	commandSet    = "set"
	commandDel    = "delete"
)

type baseConfig struct {
	Id         string `json:"id" yaml:"id"`
	Name       string `json:"name" yaml:"name"`
	Profession string `json:"profession" yaml:"profession"`
	Driver     string `json:"driver" yaml:"driver"`
	CreateTime string `json:"create_time" yaml:"create_time"`
	UpdateTime string `json:"update_time" yaml:"update_time"`
}

type Worker struct {
	store eosc.IStore
}

func (w *Worker) ProcessHandler(propose []byte) (string, []byte, error) {
	panic("implement me")
}

func (w *Worker) CommitHandler(data []byte) error {
	kv := &WorkerCmd{}
	err := kv.Decode(data)
	if err != nil {
		return err
	}
	switch kv.Key {
	case commandSet:
		{
			if kv.Config.CreateTime == "" {
				kv.Config.CreateTime = time.Now().Format("2006-01-02 15:04:05")
			}
			if kv.Config.UpdateTime == "" {
				kv.Config.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
			}
			b, err := json.Marshal(kv.Config)
			if err != nil {
				return err
			}
			return w.store.Set(eosc.StoreValue{
				Id:         kv.Config.Id,
				Profession: kv.Config.Profession,
				Name:       kv.Config.Name,
				Driver:     kv.Config.Driver,
				CreateTime: kv.Config.CreateTime,
				UpdateTime: kv.Config.UpdateTime,
				IData:      eosc.JsonData(b),
				Sing:       "",
			})
		}
	case commandDel:
		{
			return w.store.Del(kv.Config.Id)
		}
	default:
		return errInvalidKey
	}
}

// WorkerCmd 用于传输的结构
type WorkerCmd struct {
	Key    string
	Config *baseConfig
}

func (kv *WorkerCmd) Encode() ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(kv); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (kv *WorkerCmd) Decode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	if err := dec.Decode(kv); err != nil {
		return err
	}
	return nil
}
