package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

type config struct {
	Name     string `mapstructure:"name"` // this middleware name
	EtcdAddr string `mapstructure:"etcd_addr"`
	LimitDay int    `mapstructure:"limit_day"`
}

type middleware struct {
	cfg   config
	store TokenStore
}

func (this *middleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// get the bearer token from the Authorization header
	// and verify that it's valid (use the verifyToken function)
	// if valid, set the real apikey on the request context
	// if not valid, return a 401 Unauthorized response
	// if no Authorization header, return a 401 Unauthorized response
	v := req.Header.Get("Authorization")
	if v == "" || !strings.HasPrefix(v, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		// abort the middleware chain
		logrus.Error("no Authorization header")
		return
	}
	token := strings.TrimPrefix(v, "Bearer ")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		// abort the middleware chain
		logrus.Error("no token")
		return
	}

	// check token
	// if token not exists, return 401
	// if token exists, check count
	// if count > limit, return 429
	// if count <= limit, increment count
	// if increment failed, return 500
	// if increment success, continue
	ti, err := this.store.Get(token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("get token failed: %s", err)
		return
	}
	if ti == nil {
		w.WriteHeader(http.StatusUnauthorized)
		logrus.Errorf("token not exists: %s", token)
		return
	}

	if this.cfg.LimitDay > 0 && ti.Count > this.cfg.LimitDay {
		w.WriteHeader(http.StatusTooManyRequests)
		logrus.Errorf("token count >= limit: %s", token)
		return
	}

	// set req Authorization with "Bearer " + ti.Token
	req.Header.Set("Authorization", "Bearer "+ti.Token)

	next(w, req)

	// FIXME: check above request is success or not
	// TODO: check limit_day
	// increment count
	err = this.store.IncrementCount(token)
	if err != nil {
		logrus.Errorf("increment count failed: %s", err)
	}
}

func NewMiddleware(cfg map[string]interface{}) (negroni.Handler, error) {
	midd := &middleware{
		cfg: config{},
	}

	log = logrus.WithFields(logrus.Fields{
		"middleware": "openai",
	})

	if err := mapstructure.Decode(cfg, &midd.cfg); err != nil {
		log.Errorf("load config failed: %s", err)
		return nil, errors.New("decode config failed")
	}

	midd.store = NewEtcdTokenStore(midd.cfg.EtcdAddr)

	return midd, nil
}
