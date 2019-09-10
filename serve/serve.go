package serve

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ooclab/ga/forward"
)

// Run run cobra subcommand
func Run(cmd *cobra.Command, args []string) {
	if viper.GetBool("config-example") {
		fmt.Printf("%s\n", yamlConfigExample)
		os.Exit(0)
	}

	readConfig()

	services := viper.GetStringMap("services")
	if len(services) == 0 {
		logrus.Warnf("no services found, quit now")
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	listenTo := viper.GetString("listen")
	forwarder := forward.NewHTTPForward(ctx, listenTo, services)

	forwarder.Run()
	cancel() // TODO: how to cancel gracefully ?
	logrus.Debug("forwarder quit")
}

func readConfig() {

	// 1. try read config from command lie
	configPath := viper.GetString("config")
	if configPath != "" {
		logrus.Debugf("load config from %s\n", configPath)
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			logrus.Errorf("read config (%s) failed: %s\n", configPath, err)
			os.Exit(2)
		}
		if strings.HasSuffix(configPath, "yml") || strings.HasSuffix(configPath, "yaml") {
			viper.SetConfigType("yaml") // or viper.SetConfigType("YAML")
		} else if strings.HasSuffix(configPath, "json") {
			viper.SetConfigType("json")
		} else if strings.HasSuffix(configPath, "toml") {
			viper.SetConfigType("toml")
		} else {
			logrus.Errorf("unsupported config file type, just [yaml, json, toml]")
			os.Exit(2)
		}
		if err := viper.ReadConfig(bytes.NewBuffer(data)); err != nil {
			logrus.Errorf("load failed: %s\n", err)
		}
		return
	}

	// 2. try read config from default paths

	viper.SetConfigName("config")    // name of config file (without extension)
	viper.AddConfigPath("/etc/ga/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.ga") // call multiple times to add many search paths
	viper.AddConfigPath(".")         // optionally look for config in the working directory
	err := viper.ReadInConfig()      // Find and read the config file
	if err != nil {                  // Handle errors reading the config file
		logrus.Errorf("read config file failed: %s \n", err)
		os.Exit(2)
	}
}
