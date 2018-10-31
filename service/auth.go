package service

import (
	"crypto/md5"
	"fmt"

	"github.com/Sirupsen/logrus"

	"github.com/ooclab/ga/service/etcd"
)

// Auth 定义了Auth结构
type Auth struct {
	etcdSession *etcd.Session
}

// NewAuth 创建新的 Auth 对象
func NewAuth() *Auth {
	session, err := etcd.GetSession()
	if err != nil {
		logrus.Errorf("can not create new auth: get etcd session failed: %s\n", err)
		return nil
	}
	return &Auth{
		etcdSession: session,
	}
}

func (auth *Auth) etcdSetString(key, value string) error {
	return auth.etcdSession.Set(key, value)
}

func (auth *Auth) etcdGetString(key string) (string, error) {
	return auth.etcdSession.Get(key)
}

func (auth *Auth) getChecksum(v string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(v)))
}

// AddPermission 添加新的权限
func (auth *Auth) AddPermission(permName string, roles []string) error {
	var err error
	var permMD5, roleMD5, key string
	permMD5 = auth.getChecksum(permName)
	for _, roleName := range roles {
		roleMD5 = auth.getChecksum(roleName)
		key = fmt.Sprintf("/auth/permission/%s/role/%s", permMD5, roleMD5)
		if err = auth.etcdSetString(key, roleName); err != nil {
			return err
		}
	}
	return nil
}

// HasPermission 检查用户是否有指定权限
func (auth *Auth) HasPermission(userID, permName string) error {
	// TODO：在缓存中查询，需要：监听 etcd , 以确定本地缓存是否需要刷新

	var err error
	var value, key string

	// 查找指定权限需要哪些角色
	key = fmt.Sprintf("/auth/permission/%s/role", auth.getChecksum(permName))
	value, err = auth.etcdGetString(key)
	fmt.Printf("get permission role = %#v\n", value)
	if err != nil {
		logrus.Errorf("find roles of permission failed: %s\n", err)
		return err
	}

	// 查找指定用户有哪些角色
	key = fmt.Sprintf("/auth/user/%s/role", userID)
	value, err = auth.etcdGetString(key)
	fmt.Printf("get user role = %#v\n", value)
	if err != nil {
		logrus.Errorf("find roles of auth failed: %s\n", err)
		return err
	}

	logrus.Warning("this is uncompleted!")
	return err
}
