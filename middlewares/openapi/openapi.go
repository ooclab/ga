package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/mitchellh/mapstructure"
	"github.com/ooclab/ga/service"
)

var log *logrus.Entry

type config struct {
	Name        string `mapstructure:"name"` // this middleware name
	ServiceName string `mapstructure:"service_name"`
	ServiceSpec string `mapstructure:"service_spec"`
}

func (c *config) init() error {
	if c.ServiceName == "" {
		return errors.New("no service_name")
	}
	if c.ServiceSpec == "" {
		c.ServiceSpec = fmt.Sprintf("/ga/service/%s/openapi/spec", c.ServiceName)
		log.Debugf("use default service_spec = %s\n", c.ServiceSpec)
	}
	return nil
}

type openapiMiddleware struct {
	spec       *Spec
	authClient *service.Auth
	cfg        config
	// app         *service.App
	// authzClient *authz.AuthZ
}

func (h *openapiMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// do some stuff before
	perm, err := h.spec.SearchPermission(req)
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
			log.Warnf("not completed!")
			// if err := auth.authClient.HasPermission(userID, perm.Name); err != nil {
			// 	logrus.Errorf("check permission failed: %s\n", err)
			// 	writeJSON(w, 403, map[string]string{"status": err.Error()})
			// 	return
			// }
		}
	}
	next(w, req)
	// do some stuff after
}

// NewMiddleware 创建新的 Auth 中间件
func NewMiddleware(_cfg map[string]interface{}) (negroni.Handler, error) {
	h := &openapiMiddleware{
		cfg: config{},
	}

	log = logrus.WithFields(logrus.Fields{
		"middleware": "openapi",
	})

	log.Printf("_cfg = %#v\n", _cfg)

	if err := mapstructure.Decode(_cfg, &h.cfg); err != nil {
		log.Errorf("load config failed: %s", err)
		return nil, errors.New("decode config failed")
	}
	log.Printf("h.cfg = %#v\n", h.cfg)

	if err := h.cfg.init(); err != nil {
		return nil, err
	}

	// loads openapi spec
	doc, err := loadSpecFromEtcd(h.cfg.ServiceSpec)
	if err != nil {
		return nil, err
	}

	h.spec = NewSpec(h.cfg.ServiceName, doc)

	// app := service.NewApp()
	// if err := app.CheckAccess(); err != nil {
	// 	logrus.Errorf("create app failed: %s\n", err)
	// 	os.Exit(1)
	// }

	// authzClient := authz.NewAuthZ(app)
	// authClient := service.NewAuth()

	return h, nil
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
