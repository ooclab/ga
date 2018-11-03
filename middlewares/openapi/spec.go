package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/ooclab/ga/service/etcd"
)

type vError struct {
	Name    string
	Code    string
	Message string
}

func loadSpec(content []byte) (*loads.Document, error) {
	doc, err := loads.Analyzed(json.RawMessage([]byte(content)), "")
	if err != nil {
		logrus.Errorf("load spec failed: %v\n", err)
		return nil, err
	}

	validate.SetContinueOnErrors(true) // Set global options
	// Validates spec with default Swagger 2.0 format definitions
	if err = validate.Spec(doc, strfmt.Default); err != nil {
		logrus.Errorf("The spec has some validation error: %v\n", err)
	}

	doc, err = doc.Expanded()
	if err != nil {
		logrus.Errorf("failed to expand spec: %s\n", err)
		return nil, err
	}

	return doc, err
}

func loadSpecFromEtcd(openapiSpecPath string) (*loads.Document, error) {

	// get public key
	session, err := etcd.GetSession()
	if err != nil {
		logrus.Errorf("get etcd session failed: %s\n", err)
		return nil, err
	}

	specData, err := session.Get(openapiSpecPath)
	if err != nil {
		logrus.Errorf("get openapi spec path from etcd failed: %s\n", err)
		return nil, err
	}
	logrus.Debugf("load openapi spec path (%s) success\n", openapiSpecPath)

	return loadSpec([]byte(specData))
}

// LoadSpecFromPath try to load the swagger spec specified by path
func LoadSpecFromPath(path string) (*loads.Document, error) {
	doc, err := loads.Spec(path)
	if err != nil {
		logrus.Errorf("load spec (%s) failed: %v\n", path, err)
		return nil, err
	}
	validate.SetContinueOnErrors(true) // Set global options
	// Validates spec with default Swagger 2.0 format definitions
	if err = validate.Spec(doc, strfmt.Default); err != nil {
		logrus.Errorf("The spec (%s) has some validation error: %v\n", path, err)
		return nil, err
	}

	// expanded
	doc, err = doc.Expanded(&spec.ExpandOptions{RelativeBase: path})
	if err != nil {
		logrus.Errorf("failed to expand spec: %s\n", err)
		os.Exit(1)
	}

	return doc, nil
}

// Spec is a object to store info about swagger ui spec
type Spec struct {
	serviceName string
	router      *mux.Router
	doc         *loads.Document

	// 最大等待 path (url) 的秒数
	pathReadTimeout int

	// find swagger path by route
	routePathMap      map[*mux.Route]string
	routePathMapMutex *sync.Mutex

	// find swagger op by permission name
	permissionMap      map[string]*Permission
	permissionMapMutex *sync.Mutex
}

// NewSpec create a new Spec struct
func NewSpec(serviceName string, doc *loads.Document) *Spec {
	spec := &Spec{
		serviceName:     serviceName,
		doc:             doc,
		pathReadTimeout: 16, // 16 秒

		routePathMap:      make(map[*mux.Route]string),
		routePathMapMutex: &sync.Mutex{},

		permissionMap:      make(map[string]*Permission),
		permissionMapMutex: &sync.Mutex{},
	}
	spec.load()
	return spec
}

func (s *Spec) load() {
	// s.router = mux.NewRouter().PathPrefix("/api/auth").Subrouter()
	s.router = mux.NewRouter()

	for path, v := range s.doc.Spec().Paths.Paths {
		if v.Get != nil {
			s.addOperation("GET", path, v)
		}
		if v.Post != nil {
			s.addOperation("POST", path, v)
		}
		if v.Put != nil {
			s.addOperation("PUT", path, v)
		}
		if v.Delete != nil {
			s.addOperation("DELETE", path, v)
		}
		if v.Options != nil {
			s.addOperation("Options", path, v)
		}
		if v.Head != nil {
			s.addOperation("HEAD", path, v)
		}
		if v.Patch != nil {
			s.addOperation("PATCH", path, v)
		}
	}
}

func (s *Spec) getRoutePath(route *mux.Route) string {
	s.routePathMapMutex.Lock()
	defer s.routePathMapMutex.Unlock()
	return s.routePathMap[route]
}

func (s *Spec) addRoutePath(route *mux.Route, path string) {
	s.routePathMapMutex.Lock()
	defer s.routePathMapMutex.Unlock()
	s.routePathMap[route] = path
}

func (s *Spec) getPermission(permName string) *Permission {
	s.permissionMapMutex.Lock()
	defer s.permissionMapMutex.Unlock()
	return s.permissionMap[permName]
}

func (s *Spec) addPermission(perm *Permission) {
	s.permissionMapMutex.Lock()
	defer s.permissionMapMutex.Unlock()
	s.permissionMap[perm.Name] = perm
}

func (s *Spec) addOperation(method string, path string, spi spec.PathItem) {
	route := s.router.NewRoute().Methods(method).Path(path)
	s.addRoutePath(route, path)

	perm := NewPermssion(s, method, path, &spi)
	s.addPermission(perm)
}

// GetPermissionMap 返回 permission 数据
func (s *Spec) GetPermissionMap() map[string]*Permission {
	m := make(map[string]*Permission)
	s.permissionMapMutex.Lock()
	defer s.permissionMapMutex.Unlock()
	for key, value := range s.permissionMap {
		m[key] = value
	}
	return m
}

// SearchPermission 查询匹配当前请求的权限名
func (s *Spec) SearchPermission(req *http.Request) (*Permission, error) {
	var match mux.RouteMatch
	if ok := s.router.Match(req, &match); ok {
		path := s.getRoutePath(match.Route)
		permName := genPermissionName(s.serviceName, req.Method, path)
		perm := s.getPermission(permName)
		perm.routeMatch = &match
		return perm, nil
	}
	log.Debugf("match = %#v\n", match)
	return nil, errors.New("not match")
}

func getPermissionID(route *mux.Route) string {
	methods, _ := route.GetMethods()
	path, _ := route.GetPathRegexp()
	return strings.Join([]string{methods[0], path}, ":")
}

// Permission store the properites needed by permission
type Permission struct {
	Spec       *Spec
	Name       string
	method     string
	path       string
	spi        *spec.PathItem
	op         *spec.Operation
	routeMatch *mux.RouteMatch // hold the path value
	roles      []string
}

func genPermissionName(serviceName, method, path string) string {
	return strings.Join([]string{serviceName, strings.ToLower(method), path}, ":")
}

// NewPermssion create a new Permission object
func NewPermssion(gaspec *Spec, method string, path string, spi *spec.PathItem) *Permission {
	var op *spec.Operation
	switch method {
	case "GET":
		op = spi.Get
	case "POST":
		op = spi.Post
	case "PUT":
		op = spi.Put
	case "DELETE":
		op = spi.Delete
	case "OPTIONS":
		op = spi.Options
	case "HEAD":
		op = spi.Head
	case "PATCH":
		op = spi.Patch
	}
	return &Permission{
		Spec:   gaspec,
		Name:   genPermissionName(gaspec.serviceName, method, path),
		method: strings.ToLower(method),
		path:   path,
		spi:    spi,
		op:     op,
	}
}

// Summary 返回权限描述
func (p *Permission) Summary() string {
	desc := p.op.Summary
	if desc == "" {
		desc = p.op.Description
	}
	return desc
}

func (p *Permission) String() string {
	return fmt.Sprintf("%s : %s", p.Name, p.Summary())
}

// Code return the code of permission
func (p *Permission) Code() string {
	return fmt.Sprintf("%s:%s:%s", p.Spec.serviceName, p.method, p.path)
}

// validateRequest validate the current request args
func (p *Permission) validateRequest(req *http.Request) []*vError {
	var verr *vError
	var errs = []*vError{}

	parameters := []spec.Parameter{}
	parameters = append(parameters, p.spi.Parameters...)
	parameters = append(parameters, p.op.Parameters...)

	log.Warnf("in developing ...")
	allowedQueryParams := []string{}

	for _, parameter := range parameters {
		name := parameter.Name

		// Authorization
		if name == "Authorization" {
			if req.Header.Get("X-User-Id") == "" && parameter.Required {
				errs = append(errs, &vError{
					Name:    p.Name,
					Code:    "need-x-user-id",
					Message: fmt.Sprint("the http request header X-User-Id is needed"),
				})
			}
			continue
		}

		// fmt.Printf("    name: %s, in: %s, type: %s, type_name: %s, format: %s, op: %#v\n", parameter.Name, parameter.In, parameter.Type, parameter.TypeName(), parameter.Format, parameter)

		switch parameter.In {
		case "query":
			allowedQueryParams = append(allowedQueryParams, parameter.Name)
			verr = p.validateQeuryParameter(&parameter, req.URL.Query())
		case "path":
			verr = p.validatePathParameter(&parameter, req.URL.Query())
		case "header":
			verr = p.validateHeaderParameter(&parameter, req.URL.Query())
		default:
			verr = &vError{
				Name:    name,
				Code:    "unknown-parameter-in",
				Message: fmt.Sprintf("unknown parameter in: %s\n", parameter.In),
			}
		}
		if verr != nil {
			errs = append(errs, verr)
		}
	}
	// log.Debugf("%s:%s allowedQueryParams = %v\n", p.method, p.path, allowedQueryParams)
	if len(errs) == 0 {
		return nil
	}
	return errs
}

func (p *Permission) validateQeuryParameter(parameter *spec.Parameter, q url.Values) *vError {
	_, ok := q[parameter.Name]
	if !ok {
		if parameter.Required {
			return &vError{
				Name:    p.Name,
				Code:    "required",
				Message: fmt.Sprintf("query parameter (%s) is required", parameter.Name),
			}
		}
		return nil
	}

	v := q.Get(p.Name)
	// fmt.Printf("p.Name = %s, p.Type = %s, p.Enum = %#v, v = %s\n", parameter.Name, parameter.Type, parameter.Enum, v)

	switch parameter.Type {
	case "integer":
		return validateInteger(parameter, v)
	case "float":
		return validateFloat(parameter, v)
	case "string":
		return validateString(parameter, v)
	}

	return nil
}

func (p *Permission) validatePathParameter(parameter *spec.Parameter, q url.Values) *vError {
	v, ok := p.routeMatch.Vars[parameter.Name]
	if !ok {
		// all path parameter is required
		return &vError{
			Name:    parameter.Name,
			Code:    "required",
			Message: fmt.Sprintf("path parameter (%s) is required", parameter.Name),
		}
	}

	switch parameter.Type {
	case "integer":
		return validateInteger(parameter, v)
	case "float":
		return validateFloat(parameter, v)
	case "string":
		return validateString(parameter, v)
	}

	return nil
}

func (p *Permission) validateHeaderParameter(parameter *spec.Parameter, q url.Values) *vError {
	log.Warnf("uncompleted header parameter validate ...")
	return nil
}

func validateInteger(p *spec.Parameter, v string) *vError {
	iv, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return &vError{
			Name:    p.Name,
			Code:    "need-integer",
			Message: fmt.Sprintf("value (%s) is not a integer: %s\n", v, err),
		}
	}

	if p.Maximum != nil {
		maximum := int64(*p.Maximum)
		if maximum < iv {
			return &vError{
				Name:    p.Name,
				Code:    "greater-than-maximum",
				Message: fmt.Sprintf("value (%s) is greater than %d", v, maximum),
			}
		}
	}

	if p.Minimum != nil {
		minimum := int64(*p.Minimum)
		if minimum > iv {
			return &vError{
				Name:    p.Name,
				Code:    "less-than-minimum",
				Message: fmt.Sprintf("value (%s) is less than %d", v, minimum),
			}
		}
	}

	return nil
}

func validateFloat(p *spec.Parameter, v string) *vError {
	iv, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return &vError{
			Name:    p.Name,
			Code:    "need-number",
			Message: fmt.Sprintf("value (%s) is not a number: %s\n", v, err),
		}
	}

	if p.Maximum != nil && *p.Maximum < iv {
		return &vError{
			Name:    p.Name,
			Code:    "greater-than-maximum",
			Message: fmt.Sprintf("value (%s) is greater than %f", v, *p.Maximum),
		}
	}

	if p.Minimum != nil && *p.Minimum > iv {
		return &vError{
			Name:    p.Name,
			Code:    "less-than-minimum",
			Message: fmt.Sprintf("value (%s) is less than %f", v, *p.Minimum),
		}
	}

	return nil
}

func validateString(p *spec.Parameter, v string) *vError {

	// validate

	// validate maxlenght
	if p.MaxLength != nil && *p.MaxLength < int64(len(v)) {
		return &vError{
			Name:    p.Name,
			Code:    "greater-than-maxlength",
			Message: fmt.Sprintf("value (%s) is greater than %d", v, p.MaxLength),
		}
	}

	// validate minlength
	if p.MinLength != nil && *p.MinLength > int64(len(v)) {
		return &vError{
			Name:    p.Name,
			Code:    "less-than-minlength",
			Message: fmt.Sprintf("value (%s) is less than %d", v, p.MinLength),
		}
	}

	// validate enum
	if len(p.Enum) > 0 {
		for _, enum := range p.Enum {
			if enum.(string) == v {
				return nil
			}
		}
		return &vError{
			Name:    p.Name,
			Code:    "not-in-enum",
			Message: fmt.Sprintf("value (%s) is not in enum %v", v, p.Enum),
		}
	}

	// validate format
	if p.Format != "" {
		switch p.Format {
		case "uuid":
			if !isValidUUID(v) {
				return &vError{
					Name:    p.Name,
					Code:    "invalid-uuid",
					Message: fmt.Sprintf("value (%s) is not a valid uuid", v),
				}
			}
		default:
			return &vError{
				Name:    p.Name,
				Code:    "unknown-format",
				Message: fmt.Sprintf("value (%s) : format (%s) is unsupported now", v, p.Format),
			}
		}
	}

	return nil
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
