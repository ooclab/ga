package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
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

// permRequest 权限请求
type permRequest struct {
	service   string // 服务名称，限定 resource 范围
	requestor string // 请求者
	action    string // 动作：GET / POST / PUT / DELETE / PATCH / HEAD
	resource  string // `/v1/gateway/{id}`
}

func (this *permRequest) String() string {
	return fmt.Sprintf("requestor %s want %s %s in service %s scope", this.requestor, this.action, this.resource, this.service)
}

type middleware struct {
	cfg    config
	spec   *openapi3.Swagger
	router *openapi3filter.Router
}

func (this *middleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	ctx := context.TODO()

	// Find route
	route, pathParams, err := this.router.FindRoute(req.Method, req.URL)
	if err != nil {
		logrus.Errorf("can not find operation for request: %s", err)
		writeJSON(w, 400, map[string]interface{}{"error": err.Error()})
		return
	}

	// TODO: 定制化获取方式，如通过哪个 Header 获取
	requestor := req.Header.Get("X-User-Id")
	fmt.Printf("requestor = %#v\n", requestor)
	if requestor == "" {
		requestor = "anonymous"
	}

	pr := &permRequest{
		service:   this.cfg.ServiceName,
		requestor: requestor,
		action:    route.Method,
		resource:  route.Path,
	}

	// 验证请求者是否有权限访问 api
	if err := this.auth(pr); err != nil {
		logrus.Errorf("auth for \"%s\" failed: %s", pr.String(), err)
		writeJSON(w, 400, map[string]interface{}{"error": err.Error()})
		return
	}

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}
	if err := openapi3filter.ValidateRequest(ctx, requestValidationInput); err != nil {
		logrus.Errorf("valide request failed: %s", err)
		writeJSON(w, 400, map[string]interface{}{"error": "request invalid", "message": err.Error()})
		return
	}

	next(w, req)
}

// auth the requestor has perms for resource
func (this *middleware) auth(pr *permRequest) error {
	fmt.Println(pr.String())
	return nil
}

// NewMiddleware create a new middleware
func NewMiddleware(_cfg map[string]interface{}) (negroni.Handler, error) {
	var err error
	h := &middleware{
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

	// Fix UUID
	openapi3.DefineStringFormat("uuid", `^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	// loads openapi spec
	var spec *openapi3.Swagger
	addr := h.cfg.ServiceSpec
	if addr[0] == '/' {
		spec, err = openapi3.NewSwaggerLoader().LoadSwaggerFromFile(addr)
	} else if strings.HasPrefix(addr, "http") {
		var urlAddr *url.URL
		urlAddr, err = url.Parse(addr)
		if err != nil {
			logrus.Errorf("parse url %s failed: %s", addr, err)
			return nil, err
		}
		spec, err = openapi3.NewSwaggerLoader().LoadSwaggerFromURI(urlAddr)
	}
	if err != nil {
		logrus.Errorf("load spec from \"%s\" failed: %s", addr, err)
		return nil, err
	}

	// https://github.com/danielgtaylor/apisprout/blob/master/apisprout.go
	// Clear the server list so no validation happens. Note: this has a side
	// effect of no longer parsing any server-declared parameters.
	spec.Servers = make([]*openapi3.Server, 0)

	h.spec = spec
	h.router = openapi3filter.NewRouter().WithSwagger(spec)

	return h, nil
}

func writeJSON(w http.ResponseWriter, statusCode int, data map[string]interface{}) {
	data["middleware"] = "openapi3"
	data["status"] = "fail"
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
