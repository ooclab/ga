package main

import (
	"crypto/rsa"
	"errors"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	jwt "github.com/dgrijalva/jwt-go"
)

var log *logrus.Entry

type jwtMiddleware struct {
	name   string
	pubKey *rsa.PublicKey
	cfg    map[string]string
}

// NewMiddleware 创建新的UID中间件
func NewMiddleware(_cfg map[string]interface{}) (negroni.Handler, error) {
	name := _cfg["name"].(string)
	log = logrus.WithFields(logrus.Fields{
		"middleware": name,
	})

	cfg := map[string]string{}
	for k, v := range _cfg {
		switch v.(type) {
		case string:
			cfg[k] = v.(string)
		default:
			log.Errorf("unsuppported value type: %T\n", v)
		}
	}

	var err error
	var pubKey []byte
	if v, ok := cfg["public_key_etcd"]; ok {
		pubKey, err = loadPublicKeyFromEtcd(v)
		if err != nil {
			return nil, err
		}
	} else if v, ok := cfg["public_key"]; ok {
		pubKey = []byte(v)
	}
	if pubKey == nil {
		log.Errorf("no public key found!")
		return nil, errors.New("no public key")
	}

	pub, err := jwt.ParseRSAPublicKeyFromPEM(pubKey)
	if err != nil {
		logrus.Errorf("load public key failed: %s\n", err)
		return nil, err
	}
	return &jwtMiddleware{
		name:   name,
		pubKey: pub,
		cfg:    cfg,
	}, nil
}

func (h *jwtMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// do some stuff before

	idToken := getIDToken(req)
	if idToken == "" {
		// FIXME! deny a custom "X-User-Id" Header (supplied by bad user)
		req.Header["X-User-Id"] = []string{}
	} else {
		if err := h.stillTokenValid(idToken); err != nil {
			logrus.Errorf("token is invalid now: %s", err)
			writeJSON(w, 403, map[string]string{"status": err.Error()})
			return
		}

		uid, err := getUserID(idToken, h.pubKey)
		if err != nil {
			writeJSON(w, 403, map[string]string{"status": err.Error()})
			return
		}

		if err := h.stillRequestorActive(uid); err != nil {
			logrus.Errorf("requestor is inactive now: %s", err)
			writeJSON(w, 403, map[string]string{"status": err.Error()})
			return
		}

		req.Header["X-User-Id"] = []string{uid}
	}

	next(w, req)
	// do some stuff after
}

func (this *jwtMiddleware) stillTokenValid(tok string) error {
	// TODO: 确认当前 token 是否仍然有效（允许系统回收已经发出的 token）

	// 1. ga 配置 token 签发时间不能早于某个时刻（如 time.Now()）
	// 2. token claims 里有特殊配置
	// 3. 比如用户自己删除的 token （会话），调用第三方服务

	return nil
}

func (this *jwtMiddleware) stillRequestorActive(id string) error {
	// TODO: 确认当前请求者是否仍然被许可（允许系统全局禁用用户）

	// 1. token claims 里有特殊配置
	// 2. 比如系统禁用某个用户，调用第三方服务

	return nil
}
