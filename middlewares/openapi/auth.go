package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ooclab/ga/service/etcd"
)

var errPermissionNotSet = errors.New("permission-not-set")
var errInternalEtcd = errors.New("internal-etcd-error")
var errNeedLogin = errors.New("need-login")
var errNoPermission = errors.New("no-permission")

// Auth 定义了Auth结构
type Auth struct {
	etcdSession *etcd.Session
}

// NewAuth 创建新的 Auth 对象
func NewAuth() (*Auth, error) {
	session, err := etcd.GetSession()
	if err != nil {
		log.Errorf("can not create new auth: get etcd session failed: %s\n", err)
		return nil, err
	}
	return &Auth{
		etcdSession: session,
	}, nil
}

// HasPermission 检查用户是否有指定权限
func (auth *Auth) HasPermission(userID, permName string) error {
	// TODO：在缓存中查询，需要：监听 etcd , 以确定本地缓存是否需要刷新

	// get the roles needed
	prl, err := auth.getPermissionRoleList(permName)
	if err != nil {
		if err == etcd.ErrKeyNotExist {
			return errPermissionNotSet
		}
		return errInternalEtcd
	}
	prm := stringSliceToMap(prl)

	// anonymous
	if userID == "" {
		if _, ok := prm["anonymous"]; ok {
			return nil
		}
	}

	// authenticated
	if _, ok := prm["authenticated"]; ok {
		if len(prm) == 1 {
			return nil
		}
	}

	// now we need the user's roles

	if userID == "" {
		return errNeedLogin
	}

	url, err := auth.getUserRoleList(userID)
	if err != nil {
		return errInternalEtcd
	}

	for _, role := range url {
		if _, ok := prm[role]; ok {
			return nil
		}
	}
	log.Debugf("need permissions: %s", prl)
	return errNoPermission
}

func (auth *Auth) getPermissionRoleList(permName string) ([]string, error) {
	key := fmt.Sprintf("ga.auth.permissions.%s.roles", permName)
	value, err := auth.etcdSession.Get(key)
	if err != nil {
		// 注意：权限需要的角色列表不可为空，这个需要管理员注意设置准确！
		log.Errorf("get roles of permission (%s) from etcd failed: %s", permName, err)
		return nil, err
	}

	roles := []string{}
	if err := json.Unmarshal([]byte(value), &roles); err != nil {
		log.Errorf("unmarshal permission roles data failed: %s\n", err)
		return nil, err
	}

	return roles, nil
}

func (auth *Auth) getUserRoleList(userID string) ([]string, error) {
	key := fmt.Sprintf("ga.auth.users.%s.roles", userID)
	value, err := auth.etcdSession.Get(key)
	if err != nil {
		// 注意：用户角色列表可以为空，即无权限（仅 authenticated）
		if err == etcd.ErrKeyNotExist {
			return []string{}, nil
		}
		log.Errorf("get roles of user (%s) from etcd failed: %s", userID, err)
		return nil, err
	}

	// TODO: handle no roles

	roles := []string{}
	if err := json.Unmarshal([]byte(value), &roles); err != nil {
		log.Errorf("unmarshal user roles data failed: %s\n", err)
		return nil, err
	}

	return roles, nil
}

func stringSliceToMap(s []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
