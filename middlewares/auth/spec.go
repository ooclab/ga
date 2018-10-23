package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

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

	// find swagger path by route
	routePathMap      map[*mux.Route]string
	routePathMapMutex *sync.Mutex
}

// NewSpec create a new Spec struct
func NewSpec(serviceName, path string) *Spec {
	return &Spec{
		serviceName: serviceName,
		path:        path,

		routePathMap:      make(map[*mux.Route]string),
		routePathMapMutex: &sync.Mutex{},
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

func (s *Spec) addOperation(method string, path string, op *spec.Operation) {
	desc := op.Summary
	if desc == "" {
		desc = op.Description
	}
	route := s.router.NewRoute().Methods(method).Path(path)
	perm := NewPermssion(s.serviceName, route, path, desc)
	s.addRoutePath(route, path)
	fmt.Printf("%s\n", perm)
}

// Load try to load the swagger spec specified by path
func (s *Spec) Load() *loads.Document {
	doc, err := loads.Spec(s.path)
	if err == nil {
		validate.SetContinueOnErrors(true)         // Set global options
		errs := validate.Spec(doc, strfmt.Default) // Validates spec with default Swagger 2.0 format definitions

		if errs == nil {
			fmt.Println("This spec is valid")
		} else {
			fmt.Printf("The spec %s has some validation errors: %v\n", s.path, errs)
		}
	} else {
		fmt.Printf("Could not load spec %s: %v\n", s.path, err)
	}

	// s.router = mux.NewRouter().PathPrefix("/api/auth").Subrouter()
	s.router = mux.NewRouter()

	for path, v := range doc.Spec().Paths.Paths {
		if v.Get != nil {
			s.addOperation("GET", path, v.Get)
		}
		if v.Post != nil {
			s.addOperation("POST", path, v.Post)
		}
		if v.Put != nil {
			s.addOperation("PUT", path, v.Put)
		}
		if v.Delete != nil {
			s.addOperation("DELETE", path, v.Delete)
		}
		if v.Options != nil {
			s.addOperation("Options", path, v.Options)
		}
		if v.Head != nil {
			s.addOperation("HEAD", path, v.Head)
		}
		if v.Patch != nil {
			s.addOperation("PATCH", path, v.Patch)
		}
	}

	s.doc = doc
	return s.doc
}

func (s *Spec) SearchPermission(req *http.Request) (*Permission, error) {
	var match mux.RouteMatch
	if ok := s.router.Match(req, &match); ok {
		path := s.getRoutePath(match.Route)
		perm := NewPermssion(s.serviceName, match.Route, path, "")
		return perm, nil
	} else {
		return nil, errors.New("not match")
	}
}

func getPermissionID(route *mux.Route) string {
	methods, _ := route.GetMethods()
	path, _ := route.GetPathRegexp()
	return strings.Join([]string{methods[0], path}, ":")
}

// Permission store the properites needed by permission
type Permission struct {
	serviceName string
	Name        string
	Method      string
	Path        string
	Summary     string
}

// NewPermssion create a new Permission object
func NewPermssion(serviceName string, route *mux.Route, path string, summary string) *Permission {
	methods, _ := route.GetMethods()
	method := strings.ToLower(methods[0])
	code := strings.Join([]string{serviceName, method, path}, ":")
	return &Permission{
		serviceName: serviceName,
		Name:        code,
		Method:      method,
		Path:        path,
		Summary:     summary,
	}
}

func (p *Permission) String() string {
	return fmt.Sprintf("%s : %s", p.Name, p.Summary)
}

// Code return the code of permission
func (p *Permission) Code() string {
	return fmt.Sprintf("%s:%s:%s", p.serviceName, p.Method, p.Path)
}
