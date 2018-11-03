package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/mitchellh/mapstructure"
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
	cfg  config
	spec *Spec
	auth *Auth
}

func (h *openapiMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// find the permission name
	perm, err := h.spec.SearchPermission(req)
	if err != nil {
		// TODO: response 404
		log.Errorf("match permission failed: %s\n", err)
		writeJSON(w, 403, map[string]string{"status": err.Error()})
		return
	}
	log.Debugf("match permission: %s\n", perm.Name)

	// does current user has permission ?
	if err := h.auth.HasPermission(req.Header.Get("X-User-Id"), perm.Name); err != nil {
		writeJSON(w, 403, map[string]string{"status": err.Error()})
		return
	}

	// does current request has right args ?
	if errs := perm.validateRequest(req); errs != nil {
		writeJSON(w, 403, map[string]interface{}{"status": "request-args-validate-failed", "errors": errs})
		return
	}

	next(w, req)
}

// NewMiddleware 创建新的 Auth 中间件
func NewMiddleware(_cfg map[string]interface{}) (negroni.Handler, error) {
	h := &openapiMiddleware{
		cfg: config{},
	}

	log = logrus.WithFields(logrus.Fields{
		"middleware": "openapi",
	})

	if err := mapstructure.Decode(_cfg, &h.cfg); err != nil {
		log.Errorf("load config failed: %s", err)
		return nil, errors.New("decode config failed")
	}

	if err := h.cfg.init(); err != nil {
		return nil, err
	}

	// loads openapi spec
	doc, err := loadSpecFromEtcd(h.cfg.ServiceSpec)
	if err != nil {
		return nil, err
	}

	h.spec = NewSpec(h.cfg.ServiceName, doc)

	h.auth, err = NewAuth()
	if err != nil {
		log.Errorf("create auth failed: %s\n", err)
		return nil, err
	}

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
