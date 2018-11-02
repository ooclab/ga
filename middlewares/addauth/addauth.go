package main

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/ooclab/ga/service"
)

type addauthMiddleware struct {
	app *service.App
}

func (h *addauthMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// do some stuff before
	if h.app.AccessToken != "" {
		req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", h.app.AccessToken)}
	}
	next(w, req)
	// do some stuff after
}

// NewMiddleware 创建新的UID中间件
func NewMiddleware(cfg map[string]interface{}) (negroni.Handler, error) {

	app := service.NewApp()
	if err := app.CheckAccess(); err != nil {
		logrus.Errorf("app access failed: %s\n", err)
		return nil, err
	}

	return &addauthMiddleware{
		app: app,
	}, nil
}
