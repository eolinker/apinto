package store_memory

import (
	"context"
	"github.com/eolinker/eosc"
	"sync"
)

type Store struct {
	data       eosc.IUntyped
	dispatcher *eosc.StoreEventDispatcher
	locker     sync.RWMutex
}

func NewStore() (eosc.IStore ,error){

	s:=&Store{
		data:       eosc.NewUntyped(),
		dispatcher: eosc.NewStoreDispatcher(),
	}

	return s,nil
}
func (s *Store) Initialization() error {
	return nil
}

func (s *Store) All() []eosc.StoreValue {
	list:=s.data.List()
	res:=make([]eosc.StoreValue,len(list))
	for i,v:=range list{
		res[i] = *(v.(*eosc.StoreValue))
	}
	return res
}

func (s *Store) Get(id string) (eosc.StoreValue, bool) {
	if o, has := s.data.Get(id);has{
		return *o.(*eosc.StoreValue),true
	}
	return eosc.StoreValue{},false
}

func (s *Store) Set(v eosc.StoreValue) error {

	s.locker.Lock()
	defer s.locker.Unlock()
	 err:= s.dispatcher.DispatchChange(v)
	 if err!= nil{
	 	return err
	 }
	s.data.Set(v.Id,&v)
	 return nil
}


func (s *Store) Del(id string) error {
	v,has:=s.data.Del(id)
	if has{
		return 	s.dispatcher.DispatchDel(*v.(*eosc.StoreValue))
	}
	return nil
}

func (s *Store) ReadOnly() bool {
	return false
}

func (s *Store) ReadLock(ctx context.Context) (bool, error) {
	s.locker.RLock()
	return true,nil
}

func (s *Store) ReadUnLock() error {
	s.locker.RUnlock()
	return nil
}

func (s *Store) TryLock(ctx context.Context, expire int) (bool, error) {
	s.locker.Lock()
	return true,nil
}

func (s *Store) UnLock() error {
	s.locker.Unlock()
	return nil
}

func (s *Store) GetListener() eosc.IStoreListener {
	return s
}

func (s *Store) AddListen(h eosc.IStoreEventHandler) error {
	if s.dispatcher.AddListen(h){
		list:= s.All()
		return h.OnInit(list)
	}

	return nil
}
