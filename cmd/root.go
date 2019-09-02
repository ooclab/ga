package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	cobra.OnInitialize(initRootConfig)
}

func initRootConfig() {
	viper.SetEnvPrefix("GA") // will be uppercased automatically
	viper.BindEnv("DEBUG")
	viper.BindPFlag("debug", rootCmd.Flags().Lookup("verbose"))

	verbose := viper.GetBool("debug")
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
