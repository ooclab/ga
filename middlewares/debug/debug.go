package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/codegangsta/negroni"
	"github.com/spf13/viper"
)

type debugMiddleware struct {
}

func (h *debugMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if viper.GetBool("debug") || req.URL.Query().Get("debug") == "true" {
		// Save a copy of this request for debugging.
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(requestDump))
	}

	next(w, req)
}

// NewMiddleware 创建新的UID中间件
func NewMiddleware(cfg map[string]interface{}) (negroni.Handler, error) {
	return &debugMiddleware{}, nil
}

// https://medium.com/doing-things-right/pretty-printing-http-requests-in-golang-a918d5aaa000
