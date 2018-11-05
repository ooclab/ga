package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	apierrors "github.com/go-openapi/errors"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/ooclab/ga/service/etcd"
)

type vError struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	Message string `json:"message"`
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
	var errs = []*vError{}

	parameters := []spec.Parameter{}
	parameters = append(parameters, p.spi.Parameters...)
	parameters = append(parameters, p.op.Parameters...)

	allowedQueryParams := []string{}

	var bodyBuf []byte

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

		// fmt.Printf("    Name: %s\n", parameter.Name)
		// fmt.Printf("    Type: %s\n", parameter.Type)
		// fmt.Printf("    TypeName(): %s\n", parameter.TypeName())
		// fmt.Printf("    Format: %s\n", parameter.Format)
		// fmt.Printf("    In: %s\n", parameter.In)
		// fmt.Printf("    op: %#v\n", parameter)

		switch parameter.In {
		case "query":
			allowedQueryParams = append(allowedQueryParams, parameter.Name)
			if ve := p.validateQeuryParameter(&parameter, req.URL.Query()); ve != nil {
				errs = append(errs, ve)
			}
		case "path":
			if ve := p.validatePathParameter(&parameter, req); ve != nil {
				errs = append(errs, ve)
			}
		case "header":
			if ve := p.validateHeaderParameter(&parameter, req); ve != nil {
				errs = append(errs, ve)
			}
		case "body":
			if ves := p.validateBodyParameter(&parameter, req); len(ves) > 0 {
				errs = append(errs, ves...)
			}
		case "formData":
			// TODO: 和 validateBodyParameter 一起考虑
			if len(bodyBuf) == 0 {
				var err error
				if bodyBuf, err = ioutil.ReadAll(req.Body); err != nil {
					errs = append(errs, &vError{
						Name:    parameter.Name,
						Code:    "read-body-error",
						Message: fmt.Sprintf("read request body failed: %s\n", err),
					})
					return errs
				}
				req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBuf))
				// req.ParseForm()
				req.ParseMultipartForm(32 << 20)
			}
			if ve := p.validateFormDataParameter(&parameter, req); ve != nil {
				errs = append(errs, ve)
			}
		default:
			errs = append(errs, &vError{
				Name:    name,
				Code:    "unknown-parameter-in",
				Message: fmt.Sprintf("unknown parameter in: %s\n", parameter.In),
			})
		}
	}
	log.Debugf("%s:%s allowedQueryParams = %v\n", p.method, p.path, allowedQueryParams)
	if len(errs) == 0 {
		// !important! restore request body
		if len(bodyBuf) != 0 {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBuf))
		}
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
	return validateParameterStringValue(parameter, v)
}

func (p *Permission) validatePathParameter(parameter *spec.Parameter, req *http.Request) *vError {
	v, ok := p.routeMatch.Vars[parameter.Name]
	if !ok {
		// all path parameter is required
		return &vError{
			Name:    parameter.Name,
			Code:    "required",
			Message: fmt.Sprintf("path parameter (%s) is required", parameter.Name),
		}
	}

	return validateParameterStringValue(parameter, v)
}

func (p *Permission) validateHeaderParameter(parameter *spec.Parameter, req *http.Request) *vError {
	_, ok := req.Header[parameter.Name]
	if !ok {
		if parameter.Required {
			return &vError{
				Name:    parameter.Name,
				Code:    "required",
				Message: fmt.Sprintf("header parameter (%s) is required", parameter.Name),
			}
		}
	}

	v := req.Header.Get(parameter.Name)
	return validateParameterStringValue(parameter, v)
}

func (p *Permission) validateFormDataParameter(parameter *spec.Parameter, req *http.Request) *vError {
	var ok bool
	if parameter.Type == "file" {
		_, ok = req.MultipartForm.File[parameter.Name]
	} else {
		_, ok = req.MultipartForm.Value[parameter.Name]
	}
	if !ok {
		if parameter.Required {
			return &vError{
				Name:    parameter.Name,
				Code:    "required",
				Message: fmt.Sprintf("formData parameter (%s) is required", parameter.Name),
			}
		}
	}

	switch parameter.Type {
	case "integer":
	case "float":
	case "string":
		return validateParameterStringValue(parameter, req.Form.Get(parameter.Name))
	case "file":
		log.Debugf("pass parameter validate for file content (just validate required)")
	default:
		return &vError{
			Name:    parameter.Name,
			Code:    "unknown-parameter-type",
			Message: fmt.Sprintf("unknown type (%s) for parameter (%s)", parameter.Type, parameter.Name),
		}
	}

	return nil
}

func (p *Permission) validateBodyParameter(parameter *spec.Parameter, req *http.Request) (errs []*vError) {
	errs = []*vError{}

	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errs = append(errs, &vError{
			Name:    parameter.Name,
			Code:    "read-body-error",
			Message: fmt.Sprintf("read request body failed: %s\n", err),
		})
		return errs
	}

	_type := parameter.Schema.Type
	logrus.Debugf("body schema type: %v\n", _type)
	if _type.Contains("object") {
		obj := map[string]interface{}{}
		if err := json.Unmarshal(buf, &obj); err != nil {
			errs = append(errs, &vError{
				Name:    parameter.Name,
				Code:    "unmarshal-body-error",
				Message: fmt.Sprintf("unmarshal the request body (object) failed: %s\n", err),
			})
			return errs
		}
		errs = append(errs, validatebySchema(parameter.Schema, obj)...)
	} else if _type.Contains("array") {
		obj := []interface{}{}
		if err := json.Unmarshal(buf, &obj); err != nil {
			errs = append(errs, &vError{
				Name:    parameter.Name,
				Code:    "unmarshal-body-error",
				Message: fmt.Sprintf("unmarshal the request body (array) failed: %s\n", err),
			})
			return errs
		}
		errs = append(errs, validatebySchema(parameter.Schema, obj)...)
	} else {
		errs = append(errs, &vError{
			Name:    parameter.Name,
			Code:    "unknown-body-type",
			Message: fmt.Sprintf("unknown body type: %s\n", parameter.Type),
		})
	}

	if len(errs) == 0 {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	}

	return errs
}

func validatebySchema(sch *spec.Schema, data interface{}) (errs []*vError) {
	errs = []*vError{}

	err := validate.AgainstSchema(sch, data, strfmt.Default)
	ves, ok := err.(*apierrors.CompositeError)
	if ok && len(ves.Errors) > 0 {
		for _, e := range ves.Errors {
			ve := e.(*apierrors.Validation)
			log.Debugf("ve = %#v\n", ve)
			name := strings.TrimPrefix(ve.Name, ".")
			if name == "" {
				name = fmt.Sprintf("%v", ve.Value)
			}
			errs = append(errs, &vError{
				Name:    name,
				Code:    "validate-error",
				Message: strings.TrimPrefix(ve.Error(), "."),
			})
		}
	}

	return errs
}

func validateParameterStringValue(parameter *spec.Parameter, v string) *vError {
	switch parameter.Type {
	case "integer":
		return validateInteger(parameter, v)
	case "float":
		return validateFloat(parameter, v)
	case "string":
		return validateString(parameter, v)
	default:
		return &vError{
			Name:    parameter.Name,
			Code:    "unknown-parameter-type",
			Message: fmt.Sprintf("unknown type (%s) for parameter (%s)", parameter.Type, parameter.Name),
		}
	}
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
