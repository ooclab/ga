package main

import (
	"context"
	"encoding/json"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type TokenInfo struct {
	Token string `json:"token"`
	Count int    `json:"count"`
}

type TokenStore interface {
	Get(key string) (*TokenInfo, error)
	IncrementCount(key string) error
}

type EtcdTokenStore struct {
	etcdClient *clientv3.Client
}

func NewEtcdTokenStore(etcdAddr string) TokenStore {
	if etcdAddr == "" {
		etcdAddr = "localhost:2379"
	}
	fmt.Printf("connect to etcd server %s\n", etcdAddr)
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{etcdAddr},
	})
	if err != nil {
		panic(err)
	}
	return &EtcdTokenStore{etcdClient: client}
}

func (store *EtcdTokenStore) Get(key string) (*TokenInfo, error) {
	resp, err := store.etcdClient.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil // Token not found
	}

	var tokenInfo TokenInfo
	err = json.Unmarshal(resp.Kvs[0].Value, &tokenInfo)
	if err != nil {
		return nil, err
	}

	return &tokenInfo, nil
}

func (store *EtcdTokenStore) IncrementCount(key string) error {
	ctx := context.Background()

	resp, err := store.etcdClient.Get(ctx, key)
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 {
		// Handle the case where the key doesn't exist
	}

	var tokenInfo TokenInfo
	err = json.Unmarshal(resp.Kvs[0].Value, &tokenInfo)
	if err != nil {
		return err
	}

	// FIXME: lock
	tokenInfo.Count += 1
	newValue, err := json.Marshal(tokenInfo)
	if err != nil {
		return err
	}

	// Transaction to ensure atomic update
	txnResp, err := store.etcdClient.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", resp.Kvs[0].ModRevision)).
		Then(clientv3.OpPut(key, string(newValue))).
		Commit()

	if err != nil {
		return err
	}

	if !txnResp.Succeeded {
		// Handle the case where the transaction failed
	}

	return nil
}
