package authz

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/ooclab/ga/service"
	"github.com/spf13/viper"
)

// AuthZ 封装 authz 服务接口
type AuthZ struct {
	baseURL string
	app     *service.App
}

// NewAuthZ 创建 AuthZ 对象
func NewAuthZ(app *service.App) *AuthZ {
	baseURL := viper.GetString("service.authz.baseurl")

	logrus.Debugf("authz baseurl: %s\n", baseURL)

	return &AuthZ{
		baseURL: baseURL,
		app:     app,
	}
}

func (c *AuthZ) getFullURL(url string) string {
	return c.baseURL + url
}

// UpdatePermission 创建/更新 permission
func (c *AuthZ) UpdatePermission(name string, summary string, description string) (string, error) {
	url := c.getFullURL("/permission")
	resp, err := c.app.Post(url, map[string]string{
		"name":        name,
		"summary":     summary,
		"description": description,
	})
	if err != nil {
		logrus.Errorf("authz update permission failed: %s\n", err)
		return "", err
	}

	if resp.Success() {
		return resp.Body["id"].(string), nil
	}

	if resp.Status == "name-exist" {
		// get id by name
		return c.getPermissionIDByName(name)
	}

	logrus.Debugf("ceate permission response body: %s", resp.Body)
	return "", errors.New("create permission error")
}

// UpdateRole 创建/更新 role
func (c *AuthZ) UpdateRole(name string) (string, error) {
	url := c.getFullURL("/role")
	resp, err := c.app.Post(url, map[string]string{
		"name": name,
	})
	if err != nil {
		logrus.Errorf("authz update role failed: %s\n", err)
		return "", err
	}

	if resp.Success() {
		return resp.Body["id"].(string), nil
	}

	if resp.Status == "name-exist" {
		// get id by name
		return c.getRoleIDByName(name)
	}

	logrus.Debugf("ceate role response body: %s", resp.Body)
	return "", errors.New("create role error")
}

func (c *AuthZ) getPermissionIDByName(name string) (string, error) {
	url := c.getFullURL("/permission/id?name=" + name)
	resp, err := c.app.Get(url)
	if err != nil {
		logrus.Errorf("authz get permission id by name failed: %s\n", err)
		return "", err
	}
	if resp.Success() {
		return resp.Body["id"].(string), nil
	}

	logrus.Debugf("get permission id response body: %s", resp.Body)
	return "", errors.New("get permission id error")
}

func (c *AuthZ) getRoleIDByName(name string) (string, error) {
	url := c.getFullURL("/role/id?name=" + name)
	resp, err := c.app.Get(url)
	if err != nil {
		logrus.Errorf("authz get role id by name failed: %s\n", err)
		return "", err
	}
	if resp.Success() {
		return resp.Body["id"].(string), nil
	}

	logrus.Debugf("get role id response body: %s", resp.Body)
	return "", errors.New("get role id error")
}

// RoleAddPermission 为角色增加权限
func (c *AuthZ) RoleAddPermission(roleID string, permID string) error {
	url := c.getFullURL(fmt.Sprintf("/role/%s/permission/append", roleID))
	resp, err := c.app.Post(url, map[string]interface{}{
		"permissions": []string{permID},
	})
	if err != nil {
		logrus.Errorf("authz get role id by name failed: %s\n", err)
		return err
	}
	if resp.Success() {
		return nil
	}

	logrus.Debugf("add permission to role response body: %s", resp.Body)
	return errors.New("add permission to role error")
}

// HasPermission 为角色增加权限
func (c *AuthZ) HasPermission(userID string, permName string) error {
	url := c.getFullURL("/has_permission")
	resp, err := c.app.Post(url, map[string]string{
		"user_id":         userID,
		"permission_name": permName,
	})
	if err != nil {
		logrus.Errorf("authz has permission failed: %s\n", err)
		return err
	}
	if resp.Success() {
		return nil
	}

	logrus.Debugf("has permission body: %s", resp.Body)
	return errors.New("no-permission")
}
