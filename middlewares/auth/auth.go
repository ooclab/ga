package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/go-openapi/loads"
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
	// do some stuff before
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
func NewMiddleware(cfg map[string]interface{}) (negroni.Handler, error) {

	// loads openapi spec
	var err error
	var doc *loads.Document

	if v, ok := cfg["openapi_spec_etcd"]; ok {
		doc, err = loadSpecFromEtcd(v.(string))
	} else if v, ok := cfg["openapi_spec_path"]; ok {
		doc, err = loadSpecFromEtcd(v.(string))
	} else if v, ok := cfg["openapi_spec"]; ok {
		doc, err = loadSpec([]byte(v.(string)))
	}
	if err != nil {
		return nil, err
	}

	var serviceName string
	if v, ok := cfg["service_name"]; ok {
		serviceName = v.(string)
	}
	if serviceName == "" {
		logrus.Errorf("the service name is empty")
		return nil, errors.New("service name is empty")
	}

	spec := NewSpec(serviceName, doc)

	// app := service.NewApp()
	// if err := app.CheckAccess(); err != nil {
	// 	logrus.Errorf("create app failed: %s\n", err)
	// 	os.Exit(1)
	// }

	// authzClient := authz.NewAuthZ(app)
	authClient := service.NewAuth()

	return &Auth{
		spec:       spec,
		authClient: authClient,
		// app:         app,
		// authzClient: authzClient,
	}, nil
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	jData, err := json.Marshal(data)
	if err != nil {
		// logrus.Errorf("marshal json failed: %s", err)
		statusCode = http.StatusInternalServerError
	}
	w.Header().Set("Content-Type", "application/json")

	// !Important! Headers can only be written once
	// https://stackoverflow.com/questions/39427544/golang-http-response-headers-being-removed
	w.WriteHeader(statusCode)
	w.Write(jData)
}
