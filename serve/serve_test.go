package serve

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var pubKey = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA3KaaRKEjVxY+cX3mEUxW
Kpp8D9RPF8fvb4vCsQ6g3MkBpDPeQ+AcMUcHOYkbyzP9aVIY1RXWhRsb083MXpjq
03D1pSkWhog11nrffJQotnxT106N/TdZ6J06DAxF4sD4acscz24FP0IvO9IS9e92
+9H/Pmj+tB+j3RyuKoAibriTgcE6VEpFu2F9zERJ0mvMT2ycoDAVq73mhesLOraA
mVPX+TLyC3NqU2k3AFA7FDD3RVnOBTmkwBwhxpgF66wmlwsZP7SW+RhQ8kK13xkM
8U4GrXaDQGmRHn1tZIlMk2dRdt4VA3DqFJJPU5Oq8qZHBfiw3X8dJqsgYnwdpFWP
KwIDAQAB
-----END PUBLIC KEY-----
`

var token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImE1NjRlZjFkLTU3NjAtNGMxMS1iMzJlLTNmZmI3YzRmYmU4NSJ9.eyJzdWIiOiJvb2NsYWIiLCJleHAiOjMyNTAzNjc5OTk5LCJpYXQiOjE1Mzk1MjY2MDYsImlzcyI6Im9vY2xhYiIsInVpZCI6ODMyfQ.xfICxLoOtrDGpbVWBglARN4rj58q8tVo2Kte3bMzaniLDp0sebttYmOJ7n8pon2N__PmP5aN1WJIaLjNUvVUy1IARatqc7gBi2Xr_qkIIQBIwH_qWn6uKN1NHqllhtfrIsVxxKQmhN_C9WFNNOg6VyR7ckkl1h3ZA9LJqlgPZHL_t6LcB5YZKD42F0Y01mL5bGbmTkGuAJrUr2KK1Z-yLXKvKOjnW74EJF9iht1c6moNOZeqOsjcY8w6W3qEI9_Xm_OrwMlRQR66Ese4fC1iNAk5d0KSPnUfd38ln7V6CmumzlHCDr1USMjJEZAV8-ThJO9o0hNsG16DzPHCYvhteA"

func TestAuthorization(t *testing.T) {

	redirectHandler := getRedirectHandler([]byte(pubKey), "http://httpbin.ooclab.com")

	// Create server using the a router initialized elsewhere. The router
	// can be a Gorilla mux as in the question, a net/http ServeMux,
	// http.DefaultServeMux or any value that statisfies the net/http
	// Handler interface.
	ts := httptest.NewServer(redirectHandler)
	defer ts.Close()

	newreq := func(method, url string, header http.Header, body io.Reader) *http.Request {
		r, err := http.NewRequest(method, url, body)
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range header {
			for _, vv := range v {
				r.Header.Set(k, vv)
			}
		}
		return r
	}

	tests := []struct {
		name string
		r    *http.Request
	}{
		{
			name: "1: testing get by access_token",
			r: newreq(
				"GET",
				ts.URL+"/anything?access_token="+token,
				nil,
				nil,
			),
		},
		{
			name: "2: testing get by Authorization",
			r: newreq(
				"GET",
				ts.URL+"/anything",
				http.Header{"Authorization": []string{"Bearer" + " " + token}},
				nil,
			),
		},
		{
			name: "3: testing post by access_token",
			r: newreq(
				"POST",
				ts.URL+"/anything?access_token="+token,
				nil,
				nil,
			),
		},
		{
			name: "4: testing post by Authorization",
			r: newreq(
				"POST",
				ts.URL+"/anything",
				http.Header{"Authorization": []string{"Bearer" + " " + token}},
				nil,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.DefaultClient.Do(tt.r)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			// fmt.Printf("resp body = %s\n", body)

			v := map[string]interface{}{}
			if err := json.Unmarshal(body, &v); err != nil {
				t.Fatal(err)
			}
			headers := v["headers"].(map[string]interface{})
			// fmt.Printf("headers = %v\n", headers)
			uid := headers["X-User-Id"].(string)
			if uid == "" {
				t.Fatal("no X-User-Id")
			}
			if uid != "832" {
				t.Fatal("uid != 832")
			}
			// fmt.Printf("uid = %s\n", uid)
		})
	}

}
