package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type override struct {
	Header string
	Match  string
	Host   string
	Path   string
}

type config struct {
	Path     string
	Host     string
	Override override
}

func generateProxy(conf config) http.Handler {
	proxy := &httputil.ReverseProxy{Director: func(req *http.Request) {
		logrus.WithFields(logrus.Fields{
			"req.URL.Path": req.URL.Path,
			"req.URL.Host": req.URL.Host,
			"req.Host":     req.Host,
			"originHost":   conf.Host,
		}).Info("proxy handle")
		req.URL.Path = strings.TrimPrefix(req.URL.Path, conf.Path)
		originHost := conf.Host
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", originHost)
		req.Host = originHost
		req.URL.Host = originHost
		req.URL.Scheme = "http"

		if conf.Override.Header != "" && conf.Override.Match != "" {
			if req.Header.Get(conf.Override.Header) == conf.Override.Match {
				req.URL.Path = conf.Override.Path
			}
		}
	}, Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
	}}

	return proxy
}

func main() {
	r := mux.NewRouter()

	configuration := []config{
		config{
			Path: "/ooclab",
			Host: "httpbin.ooclab.com",
		},
		config{
			Path: "/{path:anything/(?:foo|bar)}",
			Host: "httpbin.org",
		},
		config{
			Path: "/org",
			Host: "httpbin.org",
			Override: override{
				Header: "X-BF-Testing",
				Match:  "integralist",
				Path:   "/anything/newthing",
			},
		},
	}

	for _, conf := range configuration {
		proxy := generateProxy(conf)

		r.PathPrefix(conf.Path).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})
	}

	log.Fatal(http.ListenAndServe(":9001", r))
}
