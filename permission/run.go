package permission

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ooclab/ga/middlewares/auth"
	"github.com/ooclab/ga/service"
)

// Run run cobra subcommand
func Run(cmd *cobra.Command, args []string) {

	// configPath := viper.GetString("configPath")
	// fmt.Printf("configPath = %s\n", configPath)
	// config, err := ioutil.ReadFile(configPath)
	// if err != nil {
	// 	fmt.Printf("read config (%s) failed: %s\n", configPath, err)
	// 	return
	// }
	// fmt.Printf("viper.AllKeys() = %#v\n", viper.AllKeys())
	//
	// viper.SetConfigType("toml")
	// viper.ReadConfig(bytes.NewBuffer(config))

	serviceName := viper.GetString("service_name")
	openapiPath := viper.GetString("openapi_path")

	logrus.Debugf("service name = %s\n", serviceName)
	logrus.Debugf("openapi path = %s\n", openapiPath)
	// loads openapi spec
	doc, err := auth.LoadSpecFromPath(openapiPath)
	if err != nil {
		return
	}

	spec := auth.NewSpec(serviceName, doc)

	// app := service.NewApp()
	// if err := app.CheckAccess(); err != nil {
	// 	logrus.Errorf("create app failed: %s\n", err)
	// 	return
	// }

	authClient := service.NewAuth()

	for k, v := range spec.GetPermissionMap() {
		if err := authClient.AddPermission(k, v.Roles()); err != nil {
			logrus.Errorf("add permission failed: %s\n", err)
			return
		}
		logrus.Debugf("add perm = %s, roles = %s\n", k, strings.Join(v.Roles(), ","))
	}
}
