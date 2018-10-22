package auth

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Auth is the middleware for authorization by swagger ui doc
type Auth struct {
	spec *Spec
}

func (auth *Auth) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// do some stuff before
	fmt.Printf("[Request] %s %s\n", req.Method, req.URL.String())
	// fmt.Printf("url = %#v\n", auth.spec.url)
	var match mux.RouteMatch
	if ok := auth.spec.router.Match(req, &match); ok {
		fmt.Println("--> matched !")
		fmt.Printf("match = %#v\n", match)
		perm := NewPermssion(auth.spec.serviceName, match.Route, "")
		fmt.Printf("permission = %s, %s\n", perm.Name, perm.Code())
	} else {
		// TODO: response 404
		fmt.Println("---> not matched !")
	}
	next(w, req)
	// do some stuff after
}

// NewMiddleware 创建新的 Auth 中间件
func NewMiddleware(serviceName, path string) *Auth {
	spec := NewSpec(serviceName, path)
	spec.Load()

	return &Auth{
		spec: spec,
	}
}
