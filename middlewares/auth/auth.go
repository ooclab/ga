package auth

import (
	"fmt"
	"net/http"
)

// Auth is the middleware for authorization by swagger ui doc
type Auth struct {
	spec *Spec
}

func (auth *Auth) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// do some stuff before
	fmt.Printf("[Request] %s %s\n", req.Method, req.URL.String())
	// fmt.Printf("url = %#v\n", auth.spec.url)
	perm, err := auth.spec.SearchPermission(req)
	if err != nil {
		// TODO: response 404
		fmt.Printf("search permission failed: %s\n", err)
	}
	fmt.Printf("match permission: %s\n", perm.Name)
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
