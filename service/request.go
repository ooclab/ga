package service

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

// Response 是 service client http 返回结构
type Response struct {
	HTTPResponse *http.Response
	Body         map[string]interface{}
	Status       string
}

// Success 表示请求是否成功
func (r *Response) Success() bool {
	return r.Status == "success"
}

// Client 封装简单的方法访问 ga 内部服务
type Client struct {
	requestTimeout time.Duration
}

// NewClient 创建新的 Request 对象
func NewClient() *Client {
	return &Client{
		requestTimeout: 10 * time.Second,
	}
}

// Post 发起 HTTP POST 请求
func (client *Client) Post(url string, body interface{}) (*Response, error) {
	return client.SendRequest("POST", url, nil, nil, body)
}

// SendRequest 发起 HTTP Request
func (client *Client) SendRequest(
	method string,
	url string,
	queryParams map[string]string,
	headers map[string]string,
	body interface{}) (*Response, error) {

	var bodyReader io.Reader

	if body != nil {
		bodyByte, err := json.Marshal(body)
		if err != nil {
			logrus.Errorf("marshal body failed: %s\n", err)
			return nil, err
		}
		bodyReader = bytes.NewBuffer(bodyByte)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		logrus.Errorf("create new request failed: %s\n", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	httpClient := &http.Client{
		Timeout: client.requestTimeout,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		logrus.Errorf("do http request failed: %s\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("read response body failed: %s\n", err)
		return nil, err
	}
	serviceResp := &Response{HTTPResponse: resp}
	if err := json.Unmarshal(respBody, &serviceResp.Body); err != nil {
		logrus.Errorf("unmarshal body json failed: %s\n", err)
		return nil, err
	}

	serviceResp.Status = serviceResp.Body["status"].(string)
	return serviceResp, nil
}
