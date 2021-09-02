package store_memory

import (
	"context"
	"sync"

	"github.com/eolinker/eosc"
)

//Store 存储器结构
type Store struct {
	data       eosc.IUntyped
	dispatcher *eosc.StoreEventDispatcher
	locker     sync.RWMutex
}

//Reset 重置存储器
func (s *Store) Reset(values []eosc.StoreValue) error {
	data := eosc.NewUntyped()
	for _, v := range values {
		data.Set(v.Id, v)
	}
	s.locker.Lock()
	defer s.locker.Unlock()
	s.data = data
	return s.dispatcher.DispatchInit(values)
}

//NewStore 创建存储器
func NewStore() (eosc.IStore, error) {

	s := &Store{
		data:       eosc.NewUntyped(),
		dispatcher: eosc.NewStoreDispatcher(),
	}

	return s, nil
}

//Initialization 初始化存储器
func (s *Store) Initialization() error {
	return nil
}

//All 返回StoreValue列表
func (s *Store) All() []eosc.StoreValue {
	list := s.data.List()
	res := make([]eosc.StoreValue, len(list))
	for i, v := range list {
		res[i] = *(v.(*eosc.StoreValue))
	}
	return res
}

//Get 根据ID获取StoreValue实例
func (s *Store) Get(id string) (eosc.StoreValue, bool) {
	if o, has := s.data.Get(id); has {
		return *o.(*eosc.StoreValue), true
	}
	return eosc.StoreValue{}, false
}

//Set 设置StoreValue实例到存储器中
func (s *Store) Set(v eosc.StoreValue) error {

	s.locker.Lock()
	defer s.locker.Unlock()

	err := s.dispatcher.DispatchChange(v)
	if err != nil {
		return err
	}

	s.data.Set(v.Id, &v)
	return nil
}

//Del 根据ID删除存储器内的StoreValue实例
func (s *Store) Del(id string) error {
	v, has := s.data.Del(id)
	if has {
		return s.dispatcher.DispatchDel(*v.(*eosc.StoreValue))
	}
	return nil
}

//ReadOnly 返回储存器是否只读状态值
func (s *Store) ReadOnly() bool {
	return false
}

//ReadLock 存储器开启读锁
func (s *Store) ReadLock(ctx context.Context) (bool, error) {
	s.locker.RLock()
	return true, nil
}

//ReadUnLock 存储器解除读锁
func (s *Store) ReadUnLock() error {
	s.locker.RUnlock()
	return nil
}

//TryLock 存储器开启写锁
func (s *Store) TryLock(ctx context.Context, expire int) (bool, error) {
	s.locker.Lock()
	return true, nil
}

//UnLock 存储器解除写锁
func (s *Store) UnLock() error {
	s.locker.Unlock()
	return nil
}

//GetListener 获取存储监听器
func (s *Store) GetListener() eosc.IStoreListener {
	return s
}

//AddListen 增加监听
func (s *Store) AddListen(h eosc.IStoreEventHandler) error {
	if s.dispatcher.AddListen(h) {
		return h.OnInit(s.All())
	}

	return nil
}
