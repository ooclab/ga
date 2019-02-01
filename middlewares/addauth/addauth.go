package main

import (
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/ooclab/ga/service"
	"github.com/Sirupsen/logrus"
)

type addauthMiddleware struct {
	app *service.App
}

func (h *addauthMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if err := h.app.CheckAccess(); err != nil {
		logrus.Warnf("app access failed: %s\n", err)
	}
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
	// 启动过程中不要测试可访问性，应为此时该服务很可能还没启动。容易形成循环依赖。
	// if err := app.CheckAccess(); err != nil {
	// 	logrus.Errorf("app access failed: %s\n", err)
	// 	return nil, err
	// }

	return &addauthMiddleware{
		app: app,
	}, nil
}
