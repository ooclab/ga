package forward

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
	errHdr := func(w http.ResponseWriter, req *http.Request, err error) {
		logrus.Errorf("w = %#v\nreq = %#v\nerr = %#v\n", w, req, err)
		switch err := err.(type) {
		case *net.OpError:
			logrus.Debugf("proxy error: %s", err)
			// https://stackoverflow.com/questions/19929386/handling-connection-reset-errors-in-go/49822466
			if syscallErr, ok := err.Err.(*os.SyscallError); ok {
				// https://golang.org/pkg/syscall/
				// fmt.Printf("%#v\n", syscallErr)
				switch syscallErr.Err {
				case syscall.ENODATA:
					logrus.Errorf("maybe backend server is offline: %s", err)
					// TODO: set backend status & notify all middlewares ?
				case syscall.ECONNRESET:
					logrus.Errorf("ECONNRESET, forward to backend failed: %s", err)
					// TODO: set backend status & notify all middlewares ?
				case syscall.ECONNREFUSED:
					logrus.Errorf("ECONNREFUSED, forward to backend failed: %s", err)
					// TODO: set backend status & notify all middlewares ?
				default:
					logrus.Errorf("unknown syscall error: %s\n", err)
					// TODO: set backend status & notify all middlewares ?
				}
			}
		default:
			logrus.Errorf("unknown proxy error: %s", err)
		}
	}
	return &httputil.ReverseProxy{Director: director, ErrorHandler: errHdr}
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
