package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ooclab/ga/serve"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve ARGS",
	Short: "start ga serve",
	Run:   serve.Run,
}

func init() {

	// cobra.OnInitialize(initConfig)

	serveCmd.Flags().Int("port", 2999, "Port to run ga serve on")
	serveCmd.Flags().StringP("service_name", "", "", "The service name")
	serveCmd.Flags().StringP("public_key", "", "", "the path of public key")
	serveCmd.Flags().StringP("backend", "b", "http://127.0.0.1:3000", "the backend server to forward")
	serveCmd.Flags().StringP("swagger_doc", "", "http://127.0.0.1:3000", "the swagger document path")

	serveCmd.MarkFlagRequired("service_name")
	serveCmd.MarkFlagRequired("public_key")

	viper.BindPFlag("port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("service_name", serveCmd.Flags().Lookup("service_name"))
	viper.BindPFlag("public_key", serveCmd.Flags().Lookup("public_key"))
	viper.BindPFlag("backend", serveCmd.Flags().Lookup("backend"))
	viper.BindPFlag("swagger_doc", serveCmd.Flags().Lookup("swagger_doc"))

	// viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	// viper.SetDefault("license", "apache")
}
