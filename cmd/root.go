package cmd

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
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
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

func initRootConfig() {
	if Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
