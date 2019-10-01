package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ooclab/ga/service/etcd"
	"github.com/sirupsen/logrus"
)

func loadPublicKeyFromEtcd(publicKeyPath string) ([]byte, error) {
	// get public key
	session, err := etcd.GetSession()
	if err != nil {
		logrus.Errorf("get etcd session failed: %s\n", err)
		return nil, err
	}

	pubKey, err := session.Get(publicKeyPath)
	if err != nil {
		logrus.Errorf("get public key from etcd failed: %s\n", err)
		return nil, err
	}
	logrus.Debugf("load public key (%s) success\n", publicKeyPath)

	return []byte(pubKey), nil
}

func getIDToken(req *http.Request) string {
	// 从 HTTP Request Header 中获取 Authorization 值
	az := req.Header.Get("Authorization")
	if az != "" {
		l := strings.Split(az, " ")
		if l[0] == "Bearer" {
			if len(l) == 2 {
				// FIXME: openapi3 校验需要 Authorization 头依赖, 此处不可以删除
				// delete(req.Header, "Authorization")
				return l[1]
			} else {
				logrus.Warnf("Authorization has more than two components: %#v", l)
			}
		}
	}

	// from cookies
	if cookie, err := req.Cookie("access_token"); err == nil {
		return cookie.Value
	}

	// Fetch access_token from URL query param
	idToken := req.URL.Query().Get("access_token")
	if idToken != "" {
		req.URL.Query().Del("access_token")
	}
	return idToken
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	jData, err := json.Marshal(data)
	if err != nil {
		// logrus.Errorf("marshal json failed: %s", err)
		statusCode = http.StatusInternalServerError
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jData)
}
