package etcd

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
)

// ErrKeyNotExist 表示该 etcd key 不存在（查无结果）
var ErrKeyNotExist = errors.New("key not exist")

var defaultEndpoints = []string{"127.0.0.1:2379"}
var session *Session

// Session is a etcd client session
type Session struct {
	endpoints []string
	cli       *clientv3.Client
}

func newSession() *Session {
	endpoints := viper.GetStringSlice("services.etcd.endpoints")
	if len(endpoints) == 0 {
		logrus.Warnf("no etcd endpoints found, use default: %s\n", defaultEndpoints)
		endpoints = defaultEndpoints
	}

	return &Session{
		endpoints: endpoints,
	}
}

// GetSession return the etcd client session
func GetSession() (*Session, error) {
	if session == nil {
		session = newSession()
	}
	if err := session.connect(); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Session) connect() error {
	if s.cli != nil {
		return nil
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   s.endpoints,
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		logrus.Errorf("create etcd client failed: %s\n", err)
		return err
	}
	s.cli = cli
	return nil
}

// Close the etcd client
func (s *Session) Close() {
	s.cli.Close()
	s.cli = nil
}

// Set store a key, value
func (s *Session) Set(key, value string) error {
	ctx := context.Background()
	resp, err := s.cli.Put(ctx, key, value)
	if err != nil {
		switch err {
		case context.Canceled:
			log.Fatalf("ctx is canceled by another routine: %v", err)
		case context.DeadlineExceeded:
			log.Fatalf("ctx is attached with a deadline is exceeded: %v", err)
		case rpctypes.ErrEmptyKey:
			log.Fatalf("client-side error: %v", err)
		default:
			log.Fatalf("bad cluster endpoints, which are not etcd servers: %v", err)
		}
		logrus.Errorf("resp = %#v", resp)
		return err
	}
	return nil
}

// Get the value with key
func (s *Session) Get(key string) (string, error) {
	ctx := context.Background()
	resp, err := s.cli.Get(ctx, key)
	if err != nil {
		switch err {
		case context.Canceled:
			log.Fatalf("ctx is canceled by another routine: %v", err)
		case context.DeadlineExceeded:
			log.Fatalf("ctx is attached with a deadline is exceeded: %v", err)
		case rpctypes.ErrEmptyKey:
			log.Fatalf("client-side error: %v", err)
		default:
			log.Fatalf("bad cluster endpoints, which are not etcd servers: %v", err)
		}
		return "", err
	}
	if len(resp.Kvs) >= 1 {
		return string(resp.Kvs[0].Value), nil
	}

	logrus.Errorf("no result found for key (%s): %#v\n", key, resp.Kvs)
	return "", ErrKeyNotExist
}
