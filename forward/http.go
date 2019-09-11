package forward

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"plugin"
	"strings"
	"syscall"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

// https://github.com/urfave/negroni#third-party-middleware
var middlewareMap = map[string]func(cfg map[string]interface{}) (negroni.Handler, error){
	"logger": func(map[string]interface{}) (negroni.Handler, error) { return negroni.NewLogger(), nil },
	"cors":   func(map[string]interface{}) (negroni.Handler, error) { return cors.AllowAll(), nil },
}

type ProxyConfig struct {
	PathPrefix  string
	BackendAddr *url.URL
}

// Idea:
// 1. support a list of backend ?

// HTTPForward forward a http request to the backend with custom middlewares
type HTTPForward struct {
	ctx      context.Context // for cancel
	listenTo string          // address to listen on
	srv      http.Server
	services map[string]interface{}
}

// NewHTTPForward create a new http forward
func NewHTTPForward(ctx context.Context, listenTo string, services map[string]interface{}) *HTTPForward {
	return &HTTPForward{
		ctx:      ctx,
		listenTo: listenTo,
		srv:      http.Server{},
		services: services,
	}
}

// Run start http forward
func (this *HTTPForward) Run() error {

	r := mux.NewRouter()

	for name, _cfg := range this.services {
		logrus.Debugf("loading config for services %s", name)
		cfg := _cfg.(map[string]interface{})

		hdr, err := this.getRedirectHandler(name, cfg)
		if err != nil {
			logrus.Errorf("get redirect handler failed: %s\n", err)
			return err
		}

		pathPrefix := cfg["path_prefix"].(string)
		r.PathPrefix(pathPrefix).Handler(hdr)
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := this.srv.Shutdown(this.ctx); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	this.srv.Addr = this.listenTo
	this.srv.Handler = r
	logrus.Infof("starting server on %s", this.srv.Addr)
	if err := this.srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Printf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
	return nil
}

func (f *HTTPForward) getRedirectHandler(svc string, serviceConfig map[string]interface{}) (http.Handler, error) {
	// load middlewares
	var middlewares []negroni.Handler
	if v, ok := serviceConfig["middlewares"]; ok {
		var cfg []interface{}
		switch v.(type) {
		case nil:
			logrus.Debugf("no middlewares found for %s, continue", svc)
		case []interface{}:
			cfg = v.([]interface{})
		default:
			logrus.Errorf("unsupport middlewares config type: %T\n", v)
			os.Exit(2)
		}
		var err error
		middlewares, err = loadMiddlewares(serviceConfig, cfg)
		if err != nil {
			logrus.Errorf("load middlewares failed: %s\n", err)
			return nil, err
		}
		logrus.Debugf("load middlewares success for %s\n", svc)
	}

	backend := serviceConfig["backend"].(string)
	backendAddr, err := url.Parse(backend)
	if err != nil {
		logrus.Errorf("parse %s failed: %s\n", backend, err)
		return nil, err
	}

	// fmt.Printf("cfg = %#v\n", cfg)
	proxy := NewSingleHostReverseProxy(&ProxyConfig{
		PathPrefix:  serviceConfig["path_prefix"].(string),
		BackendAddr: backendAddr,
	})

	n := negroni.New()
	for _, mw := range middlewares {
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
func NewSingleHostReverseProxy(cfg *ProxyConfig) *httputil.ReverseProxy {
	fmt.Printf("cfg = %#v\n", cfg)
	target := cfg.BackendAddr
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		if len(cfg.PathPrefix) != 0 {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, cfg.PathPrefix)
		}
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

func loadMiddlewares(serviceConfig map[string]interface{}, cfgs []interface{}) ([]negroni.Handler, error) {
	var middlewares []negroni.Handler
	for _, v := range cfgs {
		cfg := map[string]interface{}{}
		switch v.(type) {
		case map[string]interface{}:
			cfg = v.(map[string]interface{})
		case map[interface{}]interface{}:
			for key, value := range v.(map[interface{}]interface{}) {
				cfg[key.(string)] = value
			}
		default:
			logrus.Errorf("unsupport middleware config type: %T\n", v)
			os.Exit(3)
		}

		cfg["service"] = map[string]string{
			"path_prefix": serviceConfig["path_prefix"].(string),
		}

		name := cfg["name"].(string)
		var err error
		var mw negroni.Handler
		if fc, ok := middlewareMap[name]; ok {
			mw, err = fc(cfg)
		} else {
			logrus.Debugf("try load middleware (%s) as plugin\n", name)
			mw, err = loadPlugin(cfg)
		}
		if err != nil {
			logrus.Errorf("load %s middleware failed: %s\n", name, err)
			return nil, err
		}
		middlewares = append(middlewares, mw)
	}
	return middlewares, nil
}

func loadPlugin(cfg map[string]interface{}) (negroni.Handler, error) {
	name := cfg["name"].(string)

	var err error
	var p *plugin.Plugin
	var path string

	hd, _ := homedir.Expand(fmt.Sprintf("~/.ga/middlewares/%s.so", name))

	for _, path = range []string{
		fmt.Sprintf("middlewares/%s.so", name),
		fmt.Sprintf("/etc/ga/middlewares/%s.so", name),
		hd,
	} {
		p, err = plugin.Open(path)
		if err == nil {
			logrus.Debugf("load plugin (%s) success", path)
			break
		}
		logrus.Debugf("load plugin (%s) failed: %s", path, err)
	}
	if p == nil {
		logrus.Errorf("can not load middleware plugin %s\n", name)
		return nil, errors.New("can not load plugin")
	}

	f, err := p.Lookup("NewMiddleware")
	if err != nil {
		logrus.Errorf("lookup NewMiddleware func from plugin (%s) failed: %s", path, err)
		return nil, err
	}
	return f.(func(map[string]interface{}) (negroni.Handler, error))(cfg)
}
