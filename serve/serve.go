package serve

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ooclab/ga/middlewares/auth"
	"github.com/ooclab/ga/middlewares/uid"
)

// Run run cobra subcommand
func Run(cmd *cobra.Command, args []string) {
	port := viper.GetInt("port")
	backendAddr := viper.GetString("backend")
	pubKeyPath := viper.GetString("public_key")

	pubKey, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		logrus.Errorf("read public key from file %s failed: %s", pubKeyPath, err)
		os.Exit(1)
	}

	h := getRedirectHandler(pubKey, backendAddr)
	runServe(port, h)
}

func runServe(port int, h http.Handler) {
	var srv http.Server

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	srv.Addr = fmt.Sprintf(":%d", port)
	srv.Handler = h
	logrus.Infof("starting server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Printf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}

func getRedirectHandler(pubKey []byte, backendServer string) http.Handler {

	backendURL, err := url.Parse(backendServer)
	if err != nil {
		logrus.Errorf("parse %s failed: %s\n", backendServer, err)
		os.Exit(2)
	}
	proxy := NewSingleHostReverseProxy(backendURL)

	serviceName := viper.GetString("service_name")
	swaggerPath := viper.GetString("swagger_doc")

	n := negroni.New()
	n.Use(uid.NewMiddleware(pubKey))
	n.Use(auth.NewMiddleware(serviceName, swaggerPath))
	n.UseHandler(proxy)

	return n
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
