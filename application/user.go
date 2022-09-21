package application

import (
	"fmt"

	"github.com/eolinker/eosc"
)

type IUser interface {
	Username() string
}

type User struct {
	Expire         int64             `json:"expire"`
	Labels         map[string]string `json:"labels"`
	HideCredential bool              `json:"hide_credential"`
}

type UserInfo struct {
	Name           string
	Value          string
	Expire         int64
	HideCredential bool
	Labels         map[string]string
	TokenName      string
	Position       string
	App            IApp
}

var _ IUserManager = (*UserManager)(nil)

type IUserManager interface {
	Get(name string) (*UserInfo, bool)
	Set(appID string, user []*UserInfo)
	Del(name string)
	Check(appID string, driver string, users []IUser) error
	DelByAppID(appID string)
	List() []*UserInfo
	Count() int
}

type UserManager struct {
	// users map[string]IUser
	users   eosc.IUntyped
	connApp eosc.IUntyped
}

func (u *UserManager) Check(appID string, driver string, users []IUser) error {
	us := make(map[string]IUser)
	for _, user := range users {
		t, ok := u.get(user.Username())
		if ok {
			if t.App.Id() != appID {
				return fmt.Errorf("[%s] user(%s) is existed", driver, user.Username())
			}
		}
		if _, ok = us[user.Username()]; ok {
			return fmt.Errorf("[%s] user(%s) is repeated", driver, user.Username())
		}
		us[user.Username()] = user
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

func NewUserManager() *UserManager {
	return &UserManager{users: eosc.NewUntyped(), connApp: eosc.NewUntyped()}
}

func (u *UserManager) Get(name string) (*UserInfo, bool) {
	return u.get(name)
}

func (u *UserManager) get(name string) (*UserInfo, bool) {
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

type Auth struct {
	Type      string `json:"type"`
	Position  string `json:"position"`
	TokenName string `json:"token_name"`
}
