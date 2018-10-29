package cmd

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var inDebugMode bool
var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "ga",
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
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.ga/config.toml)")
	rootCmd.PersistentFlags().BoolVarP(&inDebugMode, "debug", "d", false, "show debug log")
}

func initConfig() {

	if inDebugMode {
		logrus.SetLevel(logrus.DebugLevel)
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
