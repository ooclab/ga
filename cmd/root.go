package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configExample = []byte(`# This is a TOML document.

title = "ga config"

[service]

	[service.authn]
	baseurl = "http://127.0.0.1:10080/authn"
	app_id = ""
	app_secret = ""

	[service.authz]
	baseurl = "http://127.0.0.1:10080/authz"
`)

// Verbose 输出详细日志
var Verbose bool
var cfgFile string

// rootCmd 是主命令对象
var rootCmd = &cobra.Command{
	Use:   "ga SUBCOMMAND ARGS",
	Short: "A lightweight middleware for service-oriented architecture",
}

// Execute execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initRootConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.ga/config.toml)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("example-config", "", false, "dump a example config")
}

func initRootConfig() {

	if Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if rootCmd.Flags().Lookup("example-config").Value.String() == "true" {
		configPath := "example-config.toml"
		ioutil.WriteFile(configPath, configExample, 0644)
		fmt.Printf("save the example config to %s\n", configPath)
		os.Exit(0)
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc/ga/")
		viper.AddConfigPath("$HOME/.ga")
		viper.AddConfigPath(".")
		viper.SetConfigName("config") // name of config file (without extension)
	}

	if err := viper.ReadInConfig(); err != nil {
		logrus.Errorf("Can't read config: %s\n", err)
		os.Exit(1)
	}
	logrus.Debugf("use config %s\n", viper.ConfigFileUsed())
}
