package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/client"
)

// Auth 定义了Auth结构
type Auth struct {
	endpoints []string
	client    client.Client
}

// NewAuth 创建新的 Auth 对象
func NewAuth() *Auth {
	return &Auth{
		endpoints: viper.GetStringSlice("etcd.endpoints"),
	}
}

// Connect 连接到 etcd
func (auth *Auth) Connect() error {
	cfg := client.Config{
		Endpoints: auth.endpoints,
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		logrus.Errorf("create etcd client failed: %s\n", err)
		return err
	}
	auth.client = c
	return nil
}

func (auth *Auth) getChecksum(v string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(v)))
}

func (auth *Auth) etcdSetString(key, value string) error {
	kapi := client.NewKeysAPI(auth.client)
	// 在权限下创建一个角色
	resp, err := kapi.Set(context.Background(), key, value, nil)
	if err != nil {
		if err == context.Canceled {
			// ctx is canceled by another routine
			logrus.Error("ctx is canceled by another routine")
		} else if err == context.DeadlineExceeded {
			// ctx is attached with a deadline and it exceeded
			logrus.Error("ctx is attached with a deadline and it exceeded")
		} else if err, ok := err.(*client.ClusterError); ok {
			// process (cerr.Errors)
			logrus.Errorf("add role to permission failed: %#v: %s\n", resp, err)
		} else {
			// bad cluster endpoints, which are not etcd servers
			logrus.Errorf("bad cluster endpoints, which are not etcd servers: %s\n", err)
		}
		return err
	}
	return nil
}

func (auth *Auth) etcdGetString(key string) (string, error) {
	kapi := client.NewKeysAPI(auth.client)
	resp, err := kapi.Get(context.Background(), key, nil)
	if err != nil {
		if err == context.Canceled {
			// ctx is canceled by another routine
			logrus.Error("ctx is canceled by another routine")
		} else if err == context.DeadlineExceeded {
			// ctx is attached with a deadline and it exceeded
			logrus.Error("ctx is attached with a deadline and it exceeded")
		} else if err, ok := err.(*client.ClusterError); ok {
			// process (cerr.Errors)
			logrus.Errorf("get permission's roles failed: %#v: %s\n", resp, err)
		} else {
			// bad cluster endpoints, which are not etcd servers
			logrus.Errorf("bad cluster endpoints, which are not etcd servers: %s\n", err)
		}
		return "", err
	}

	// print common key info
	logrus.Debugf("Get is done. Metadata is %q\n", resp)
	// print value
	logrus.Debugf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)

	return resp.Node.Value, nil
}

func (auth *Auth) etcdGet(key string) (*client.Response, error) {
	kapi := client.NewKeysAPI(auth.client)
	resp, err := kapi.Get(context.Background(), key, nil)
	if err != nil {
		if err == context.Canceled {
			// ctx is canceled by another routine
			logrus.Error("ctx is canceled by another routine")
		} else if err == context.DeadlineExceeded {
			// ctx is attached with a deadline and it exceeded
			logrus.Error("ctx is attached with a deadline and it exceeded")
		} else if err, ok := err.(*client.ClusterError); ok {
			// process (cerr.Errors)
			logrus.Errorf("get permission's roles failed: %#v: %s\n", resp, err)
		} else {
			// bad cluster endpoints, which are not etcd servers
			logrus.Errorf("bad cluster endpoints, which are not etcd servers: %s\n", err)
		}
		return nil, err
	}

	// print common key info
	logrus.Debugf("Get is done. Metadata is %q\n", resp)
	// print value
	logrus.Debugf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)

	return resp, nil
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
	var key string

	// 查找指定权限需要哪些角色
	key = fmt.Sprintf("/auth/permission/%s/role", auth.getChecksum(permName))
	resp, err := auth.etcdGet(key)
	fmt.Printf("get permission role resp = %#v\n", resp)
	if !resp.Node.Dir {
		logrus.Errorf("this must be dir!")
	}
	if err != nil {
		return err
	}

	// 查找指定用户有哪些角色
	key = fmt.Sprintf("/auth/user/%s/role", userID)
	resp, err = auth.etcdGet(key)
	fmt.Printf("get user role resp = %#v\n", resp)

	return err
}
