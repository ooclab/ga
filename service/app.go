package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

type createTokenResponseData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
}

// App 是 authn 的一个访问实体
type App struct {
	baseURL   string
	appID     string // app_id
	appSecret string // app_secret

	client *Client

	AccessToken  string
	refreshToken string
	expiresIn    time.Time
}

// NewApp 创建 App 结构
func NewApp() *App {
	baseURL := viper.GetString("services.authn.baseurl")
	appID := viper.GetString("services.authn.app_id")
	appSecret := viper.GetString("services.authn.app_secret")

	logrus.Debugf("authn baseurl: %s\n", baseURL)

	return &App{
		baseURL:   baseURL,
		appID:     appID,
		appSecret: appSecret,
		client:    NewClient(),
	}
}

func (app *App) getFullURL(url string) string {
	return app.baseURL + url
}

// curl http://127.0.0.1:10080/authn/app_token \
//     -d '{"app_id": "9dc52220-f611-4300-9242-2aff30f7902a", "app_secret": "RqNIleKoWqrPozsizYWcveUSmglQZfSf"}'
func (app *App) login() error {
	url := app.getFullURL("/app_token")
	resp, err := app.client.Post(url, map[string]string{
		"app_id":     app.appID,
		"app_secret": app.appSecret,
	})
	if err != nil {
		logrus.Errorf("get app access_token failed: %s\n", err)
		return err
	}
	if status, ok := resp.Body["status"]; ok {
		if status.(string) != "success" {
			return fmt.Errorf("app login with error %s", status)
		}
	}

	data, ok := resp.Body["data"].(map[string]interface{})
	if !ok {
		logrus.Errorf("unknown data in body: %v\n", resp.Body)
		return errors.New("respone body.data error")
	}

	expiresIn, err := time.Parse(time.RFC3339, data["expires_in"].(string))
	if err != nil {
		logrus.Errorf("parse expires_in time failed: %s\n", err)
		return err
	}

	app.AccessToken = data["access_token"].(string)
	app.refreshToken = data["refresh_token"].(string)
	app.expiresIn = expiresIn
	return nil
}

func (app *App) doRefreshToken() error {
	url := app.getFullURL("/app_token/refresh")
	resp, err := app.Post(url, map[string]string{
		"app_id":        app.appID,
		"refresh_token": app.refreshToken,
	})
	if err != nil {
		logrus.Errorf("app access_token refresh failed: %s\n", err)
		return err
	}

	data, ok := resp.Body["data"].(map[string]interface{})
	if !ok {
		logrus.Errorf("unknown data in body: %v\n", resp.Body)
		return errors.New("respone body.data error")
	}

	app.expiresIn, err = time.Parse(time.RFC3339, data["expires_in"].(string))
	if err != nil {
		logrus.Errorf("parse expires_in time failed: %s\n", err)
		return err
	}

	app.AccessToken = data["access_token"].(string)
	app.refreshToken = data["refresh_token"].(string)
	return nil
}

func (app *App) beforeRequest() error {
	if app.AccessToken == "" {
		logrus.Debug("app: first time to login")
		return app.login()
	}

	if time.Now().UTC().Before(app.expiresIn) {
		logrus.Debugf("need refresh token, expires in %s\n", app.expiresIn)
		return app.doRefreshToken()
	}
	return nil
}

// CheckAccess 验证可访问性
func (app *App) CheckAccess() error {
	return app.beforeRequest()
}

func (app *App) sendRequest(
	method string,
	url string,
	queryParams map[string]string,
	body interface{}) (*Response, error) {
	// Set Access Token
	headers := map[string]string{}
	if app.AccessToken != "" {
		headers["Authorization"] = "Bearer " + app.AccessToken
	}
	return app.client.SendRequest(method, url, queryParams, headers, body)
}

// Get Send a HTTP Get Request
func (app *App) Get(url string) (*Response, error) {
	return app.sendRequest("GET", url, nil, nil)
}

// Post Send a HTTP Post Request
func (app *App) Post(url string, body interface{}) (*Response, error) {
	return app.sendRequest("POST", url, nil, body)
}
