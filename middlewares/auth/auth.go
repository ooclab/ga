package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/ooclab/ga/service"
)

// Auth is the middleware for authorization by swagger ui doc
type Auth struct {
	spec       *Spec
	authClient *service.Auth
	// app         *service.App
	// authzClient *authz.AuthZ
}

func (auth *Auth) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	logrus.Debugf("call auth middleware ...")
	// do some stuff before
	fmt.Printf("[Request] %s %s\n", req.Method, req.URL.String())
	// fmt.Printf("url = %#v\n", auth.spec.url)
	perm, err := auth.spec.SearchPermission(req)
	if err != nil {
		// TODO: response 404
		logrus.Errorf("match permission failed: %s\n", err)
		writeJSON(w, 403, map[string]string{"status": err.Error()})
		return
	}
	logrus.Debugf("match permission: %s, need roles: %s\n", perm.Name, perm.Roles())

	if perm.NeedPermission() {
		userID := req.Header.Get("X-User-Id")
		if userID == "" {
			writeJSON(w, 403, map[string]string{"status": "need-authorization"})
			return
		}

		if !perm.JustAuthenticated() {
			if err := auth.authClient.HasPermission(userID, perm.Name); err != nil {
				logrus.Errorf("check permission failed: %s\n", err)
				writeJSON(w, 403, map[string]string{"status": err.Error()})
				return
			}
		}
	}
	next(w, req)
	// do some stuff after
}

// NewMiddleware 创建新的 Auth 中间件
func NewMiddleware(serviceName, path string) *Auth {
	spec := NewSpec(serviceName, path)
	spec.Load()

	// app := service.NewApp()
	// if err := app.CheckAccess(); err != nil {
	// 	logrus.Errorf("create app failed: %s\n", err)
	// 	os.Exit(1)
	// }

	// authzClient := authz.NewAuthZ(app)
	authClient := service.NewAuth()
	if err := authClient.Connect(); err != nil {
		logrus.Errorf("auth client connect failed: %s\n", err)
	}

	return &Auth{
		spec:       spec,
		authClient: authClient,
		// app:         app,
		// authzClient: authzClient,
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	jData, err := json.Marshal(data)
	if err != nil {
		// logrus.Errorf("marshal json failed: %s", err)
		statusCode = http.StatusInternalServerError
	}
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ABC", "123")
	w.Write(jData)
}
