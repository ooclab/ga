package forward

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
)

// Idea:
// 1. support a list of backend ?

// HTTPForward forward a http request to the backend with custom middlewares
type HTTPForward struct {
	ctx         context.Context // for cancel
	middlewares []negroni.Handler
	listenTo    string   // address to listen on
	backendAddr *url.URL // forward request to this backend
	srv         http.Server
}

// NewHTTPForward create a new http forward
func NewHTTPForward(ctx context.Context, middlewares []negroni.Handler, listenTo string, backendAddr *url.URL) *HTTPForward {
	return &HTTPForward{
		ctx:         ctx,
		middlewares: middlewares,
		listenTo:    listenTo,
		backendAddr: backendAddr,
		srv:         http.Server{},
	}
}

// Run start http forward
func (f *HTTPForward) Run() error {
	handler, err := f.getRedirectHandler()
	if err != nil {
		logrus.Errorf("get redirect handler failed: %s\n", err)
		return err
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := f.srv.Shutdown(f.ctx); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	f.srv.Addr = f.listenTo
	f.srv.Handler = handler
	logrus.Infof("starting server on %s", f.srv.Addr)
	if err := f.srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Printf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
	return nil
}

func (f *HTTPForward) getRedirectHandler() (http.Handler, error) {
	proxy := NewSingleHostReverseProxy(f.backendAddr)

	n := negroni.New()
	for _, mw := range f.middlewares {
		n.Use(mw)
	}
	n.UseHandler(proxy)

	return n, nil
}

// NewSingleHostReverseProxy returns a new ReverseProxy that routes
// URLs to the scheme, host, and base path provided in target. If the
// target's path is "/base" and the incoming request was for "/dir",
// the target request will be for /base/dir.
// NewSingleHostReverseProxy does not rewrite the Host header.
// To rewrite Host headers, use ReverseProxy directly with a custom
// Director policy.
func NewSingleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}

		// TODO: make a choice
		req.Host = target.Host
	}
	return &httputil.ReverseProxy{Director: director}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
