package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/gorilla/mux"
)

// Spec is a object to store info about swagger ui spec
type Spec struct {
	serviceName string
	path        string
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
func NewSpec(serviceName, path string) *Spec {
	return &Spec{
		serviceName:     serviceName,
		path:            path,
		pathReadTimeout: 16, // 16 秒

		routePathMap:      make(map[*mux.Route]string),
		routePathMapMutex: &sync.Mutex{},

		permissionMap:      make(map[string]*Permission),
		permissionMapMutex: &sync.Mutex{},
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

// Load try to load the swagger spec specified by path
func (s *Spec) Load() *loads.Document {
	doc, err := s.loadSpecAndWait()
	if err == nil {
		validate.SetContinueOnErrors(true)         // Set global options
		errs := validate.Spec(doc, strfmt.Default) // Validates spec with default Swagger 2.0 format definitions

		if errs == nil {
			logrus.Debugf("The spec (%s) is valid", s.path)
		} else {
			logrus.Errorf("The spec (%s) has some validation errors: %v\n", s.path, errs)
		}
	} else {
		logrus.Errorf("could not load spec (%s): %v\n", s.path, err)
		os.Exit(1)
	}

	doc, err = doc.Expanded(&spec.ExpandOptions{RelativeBase: s.path})
	if err != nil {
		logrus.Errorf("failed to expand spec: %s\n", err)
		os.Exit(1)
	}

	// s.router = mux.NewRouter().PathPrefix("/api/auth").Subrouter()
	s.router = mux.NewRouter()

	for path, v := range doc.Spec().Paths.Paths {
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

	s.doc = doc
	return s.doc
}

func (s *Spec) loadSpecAndWait() (doc *loads.Document, err error) {
	timeout := 0
	for {
		doc, err = loads.Spec(s.path)
		if err != nil {
			if e, ok := err.(*url.Error); ok {
				if strings.HasSuffix(e.Err.Error(), "connection refused") ||
					strings.HasSuffix(e.Err.Error(), "no such host") {
					if timeout < s.pathReadTimeout {
						fmt.Printf(".")
						time.Sleep(1 * time.Second)
						timeout++
						continue
					}
				}
			}
		}
		return
	}
}

// SearchPermission 查询匹配当前请求的权限名
func (s *Spec) SearchPermission(req *http.Request) (*Permission, error) {
	var match mux.RouteMatch
	if ok := s.router.Match(req, &match); ok {
		path := s.getRoutePath(match.Route)
		permName := genPermissionName(s.serviceName, req.Method, path)
		perm := s.getPermission(permName)
		return perm, nil
	}
	return nil, errors.New("not match")
}

func getPermissionID(route *mux.Route) string {
	methods, _ := route.GetMethods()
	path, _ := route.GetPathRegexp()
	return strings.Join([]string{methods[0], path}, ":")
}

// Permission store the properites needed by permission
type Permission struct {
	Spec   *Spec
	Name   string
	method string
	path   string
	spi    *spec.PathItem
	op     *spec.Operation
	roles  []string
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

// Roles 返回该权限需要的角色
func (p *Permission) Roles() []string {
	if p.roles != nil {
		return p.roles
	}

	p.roles = []string{}

	// 1. 检查自定义的权限
	extensions := p.op.VendorExtensible.Extensions
	if extensions != nil {
		for extName, extValue := range extensions {
			if extName == "x-roles" {
				for _, roleName := range extValue.([]interface{}) {
					p.roles = append(p.roles, roleName.(string))
				}
				break
			}
		}
	}
	if len(p.roles) != 0 {
		return p.roles
	}

	// 2. 如果没有自定义权限再判断 Authorization 判断
	if needAuth(p.spi.PathItemProps.Parameters) || needAuth(p.op.OperationProps.Parameters) {
		p.roles = append(p.roles, "authenticated")
		return p.roles
	}

	// 3. 如果其他权限都没有，表明只需要匿名权限
	p.roles = append(p.roles, "anonymous")
	return p.roles
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

// NeedPermission 是否需要权限
func (p *Permission) NeedPermission() bool {
	for _, roleName := range p.Roles() {
		if roleName == "anonymous" {
			return false
		}
	}
	return true
}

// JustAuthenticated 检测是否仅仅需要登录权限
func (p *Permission) JustAuthenticated() bool {
	for _, roleName := range p.Roles() {
		if roleName == "authenticated" {
			return true
		}
	}
	return false
}

func needAuth(parameters []spec.Parameter) bool {
	for _, parameter := range parameters {
		if parameter.ParamProps.Name == "Authorization" {
			return true
		}
	}
	return false
}