package uid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	jwt "github.com/dgrijalva/jwt-go"
)

// UID is a middleware handle authorization token(jwt)
type UID struct {
	pubKey []byte
}

func (uid *UID) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// do some stuff before

	idToken := getIDToken(req)
	if idToken != "" {
		uid, err := getUserID(idToken, uid.pubKey)
		if err != nil {
			logrus.Errorf("get uid failed: %s", err)
			writeJSON(w, 403, map[string]string{"status": err.Error()})
			return
		}
		req.Header["X-User-Id"] = []string{uid}
	}

	next(w, req)
	// do some stuff after
}

// NewMiddleware 创建新的UID中间件
func NewMiddleware(pubKey []byte) *UID {
	return &UID{
		pubKey: pubKey,
	}
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

	// Fetch access_token from URL query param
	idToken := req.URL.Query().Get("access_token")
	req.URL.Query().Del("access_token")
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
