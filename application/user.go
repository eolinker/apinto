package application

import (
	"fmt"

	"github.com/eolinker/eosc"
)

type IUser interface {
	Username() string
}

type User struct {
	Labels         map[string]string `json:"labels" label:"用户标签"`
	Expire         int64             `json:"expire" label:"过期时间" format:"date-time"`
	HideCredential bool              `json:"hide_credential" label:"是否隐藏证书"`
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
	Additional     interface{}
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
	users   eosc.Untyped[string, *UserInfo]
	connApp eosc.Untyped[string, []string]
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
	return u.users.List()

}

func NewUserManager() *UserManager {
	return &UserManager{users: eosc.BuildUntyped[string, *UserInfo](), connApp: eosc.BuildUntyped[string, []string]()}
}

func (u *UserManager) Get(name string) (*UserInfo, bool) {
	return u.get(name)
}

func (u *UserManager) get(name string) (*UserInfo, bool) {
	return u.users.Get(name)
}

func (u *UserManager) Set(appID string, users []*UserInfo) {

	userMap := make(map[string]struct{})
	names, has := u.getByAppID(appID)
	if has {
		for _, name := range names {
			userMap[name] = struct{}{}
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
	return u.connApp.Del(appID)
}

func (u *UserManager) getByAppID(appID string) ([]string, bool) {
	return u.connApp.Get(appID)
}

type Auth struct {
	Type      string `json:"type" label:"鉴权类型" skip:""`
	Position  string `json:"position" label:"token位置" enum:"header,query,body"`
	TokenName string `json:"token_name" label:"token名称"`
}
