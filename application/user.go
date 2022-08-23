package application

import "github.com/eolinker/eosc"

type User struct {
	Expire         int64             `json:"expire"`
	Labels         map[string]string `json:"labels"`
	Pattern        map[string]string `json:"pattern"`
	HideCredential bool              `json:"hide_credential"`
}

type UserInfo struct {
	Name           string
	Expire         int64
	Labels         map[string]string
	HideCredential bool
	AppLabels      map[string]string
	Disable        bool
}

var _ IUserManager = (*UserManager)(nil)

type IUserManager interface {
	Get(name string) (*UserInfo, bool)
	Set(appID string, user []*UserInfo)
	Del(name string)
	DelByAppID(appID string)
	List() []*UserInfo
}

type UserManager struct {
	users   eosc.IUntyped
	connApp eosc.IUntyped
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
