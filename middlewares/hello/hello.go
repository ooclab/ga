package main

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/codegangsta/negroni"
)

type helloMiddleware struct {
}

func (h *helloMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// do some stuff before
	logrus.Warnf("==> this is from hello middleware plugin: %s", req.RequestURI)
	next(w, req)
	// do some stuff after
}

// NewMiddleware 创建新的UID中间件
func NewMiddleware(cfg map[string]interface{}) (negroni.Handler, error) {
	return &helloMiddleware{}, nil
}
