package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ooclab/ga/service/etcd"
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

func getUserID(idToken string, pubKey []byte) (userid string, err error) {

	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		return jwt.ParseRSAPublicKeyFromPEM(pubKey)
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userid := fmt.Sprintf("%v", claims["uid"])
		return userid, nil
	}

	return
}

func getIDToken(req *http.Request) string {
	// 从 HTTP Request Header 中获取 Authorization 值
	az := req.Header.Get("Authorization")
	if az != "" {
		l := strings.Split(az, " ")
		if l[0] == "Bearer" {
			if len(l) == 2 {
				delete(req.Header, "Authorization")
				return l[1]
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
