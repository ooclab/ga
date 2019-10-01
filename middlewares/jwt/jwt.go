package main

import (
	"crypto/rsa"
	"errors"
	"net/http"

	"github.com/codegangsta/negroni"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

type config struct {
	Name          string `mapstructure:"name"` // this middleware name
	UserIDKey     string `mapstructure:"user_id_key"`
	PublicKey     string `mapstructure:"public_key"`
	PublicKeyEtcd string `mapstructure:"public_key_etcd"`
	ServiceConfig struct {
		PathPrefix string `mapstructure:"path_prefix"`
	} `mapstructure:"service"`
}

type middleware struct {
	cfg    config
	name   string
	pubKey *rsa.PublicKey
}

// NewMiddleware 创建新的UID中间件
func NewMiddleware(_cfg map[string]interface{}) (negroni.Handler, error) {
	name := _cfg["name"].(string)
	log = logrus.WithFields(logrus.Fields{
		"middleware": name,
	})

	h := &middleware{
		name: name,
		cfg:  config{},
	}

	if err := mapstructure.Decode(_cfg, &h.cfg); err != nil {
		log.Errorf("load config failed: %s", err)
		return nil, errors.New("decode config failed")
	}

	if h.cfg.UserIDKey == "" {
		h.cfg.UserIDKey = "uid"
	}

	var err error
	var pubKey []byte
	if h.cfg.PublicKeyEtcd != "" {
		pubKey, err = loadPublicKeyFromEtcd(h.cfg.PublicKeyEtcd)
		if err != nil {
			return nil, err
		}
	} else if h.cfg.PublicKey != "" {
		pubKey = []byte(h.cfg.PublicKey)
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

	h.pubKey = pub

	return h, nil
}

func (h *middleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
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

		uid, err := h.getUserID(idToken, h.pubKey)
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

func (this *middleware) stillTokenValid(tok string) error {
	// TODO: 确认当前 token 是否仍然有效（允许系统回收已经发出的 token）

	// 1. ga 配置 token 签发时间不能早于某个时刻（如 time.Now()）
	// 2. token claims 里有特殊配置
	// 3. 比如用户自己删除的 token （会话），调用第三方服务

	return nil
}

func (this *middleware) stillRequestorActive(id string) error {
	// TODO: 确认当前请求者是否仍然被许可（允许系统全局禁用用户）

	// 1. token claims 里有特殊配置
	// 2. 比如系统禁用某个用户，调用第三方服务

	return nil
}

func (this *middleware) getUserID(idToken string, pubKey *rsa.PublicKey) (userid string, err error) {
	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		// TODO: support other
		return pubKey, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims[this.cfg.UserIDKey]
		if !ok {
			logrus.WithFields(logrus.Fields{
				"claims":      claims,
				"user_id_key": this.cfg.UserIDKey,
			}).Debugf("can not find user id in claims")
			return "", nil
		}
		return id.(string), nil
	}

	return
}
