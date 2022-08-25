package application

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
)

type UserGet func(map[string]string) (string, bool)

type User struct {
	Expire         int64             `json:"expire"`
	Labels         map[string]string `json:"labels"`
	Pattern        map[string]string `json:"pattern"`
	HideCredential bool              `json:"hide_credential"`
}

type UserInfo struct {
	AppID          string
	Name           string
	Value          string
	Expire         int64
	Labels         map[string]string
	HideCredential bool
	AppLabels      map[string]string
	Disable        bool
	TokenName      string
	Position       string
}

var _ IUserManager = (*UserManager)(nil)

type IUserManager interface {
	Get(name string) (*UserInfo, bool)
	Set(appID string, user []*UserInfo)
	Del(name string)
	Check(appID string, driver string, users []*User) error
	DelByAppID(appID string)
	List() []*UserInfo
	Count() int
}

type UserManager struct {
	users       eosc.IUntyped
	connApp     eosc.IUntyped
	getUserFunc UserGet
}

func (u *UserManager) Check(appID string, driver string, users []*User) error {
	us := make(map[string]*User)
	for _, user := range users {
		name, has := u.getUserFunc(user.Pattern)
		if !has {
			return fmt.Errorf("[%s] invalid user", driver)
		}
		t, ok := u.get(name)
		if ok {
			log.Debug(name, " appid is ", t.AppID, " ", appID)
			if t.AppID != appID {
				return fmt.Errorf("[%s] user(%s) is existed", driver, name)
			}
		} else {
			log.Debug("no has")
		}
		if _, ok = us[name]; ok {
			return fmt.Errorf("[%s] user(%s) is repeated", driver, name)
		}
		us[name] = user
	}
	return nil
}

func (u *UserManager) Count() int {
	return u.users.Count()
}

func (u *UserManager) List() []*UserInfo {
	users := u.users.List()
	us := make([]*UserInfo, 0, len(users))
	for _, user := range users {
		us = append(us, user.(*UserInfo))
	}
	return us
}

func NewUserManager(getUserFunc UserGet) *UserManager {
	return &UserManager{users: eosc.NewUntyped(), connApp: eosc.NewUntyped(), getUserFunc: getUserFunc}
}

func (u *UserManager) Get(name string) (*UserInfo, bool) {
	return u.get(name)
}

func (u *UserManager) get(name string) (*UserInfo, bool) {
	log.Debug("get user name:", name)
	user, has := u.users.Get(name)
	if !has {
		return nil, false
	}

	return user.(*UserInfo), true
}

func (u *UserManager) Set(appID string, users []*UserInfo) {

	userMap := make(map[string]bool)
	names, has := u.getByAppID(appID)
	if has {
		for _, name := range names {
			userMap[name] = true
		}
	}

	newUsers := make([]string, 0, len(users))
	for _, user := range users {

		u.users.Set(user.Name, user)
		newUsers = append(newUsers, user.Name)
		delete(userMap, user.Name)
	}
	for name := range userMap {
		u.users.Del(name)
	}
	u.connApp.Set(appID, newUsers)
}

func (u *UserManager) Del(name string) {
	u.users.Del(name)
}

func (u *UserManager) DelByAppID(appID string) {
	names, has := u.delByAppID(appID)
	if !has {
		return
	}
	for _, name := range names {
		u.users.Del(name)
	}
}

func (u *UserManager) delByAppID(appID string) ([]string, bool) {
	names, has := u.connApp.Del(appID)
	if !has {
		return nil, false
	}
	return names.([]string), true
}

func (u *UserManager) getByAppID(appID string) ([]string, bool) {
	names, has := u.connApp.Get(appID)
	if !has {
		return nil, false
	}
	return names.([]string), true
}
