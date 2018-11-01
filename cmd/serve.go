// cmd
// Example:
// ga serve \
//     --service authz \
//     --backend http://127.0.0.1:3000 \
//     --port 2999
//
// `--service` specify the name of this service, this is the relative path in etcd, for example: `/service/authz`
// `--backend` specify the backend server

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
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.SetEnvPrefix("GA") // will be uppercased automatically
		viper.BindEnv("PORT_EXTERNAL")
		viper.BindEnv("PORT_INTERNAL")
		viper.BindEnv("SERVICE")
		viper.BindEnv("SERVICE_EXTERNAL")
		viper.BindEnv("ETCD_ENDPOINTS")

		viper.BindPFlag("service", cmd.Flags().Lookup("service"))
		viper.BindPFlag("port_external", cmd.Flags().Lookup("port_external"))
		viper.BindPFlag("port_internal", cmd.Flags().Lookup("port_internal"))
		viper.BindPFlag("service_external", cmd.Flags().Lookup("service_external"))
		viper.BindPFlag("service_internal", cmd.Flags().Lookup("service_internal"))
		viper.BindPFlag("etcd_endpoints", cmd.Flags().Lookup("etcd_endpoints"))
	},
}

func init() {
	serveCmd.Flags().StringP("service", "s", "", "the service name")
	serveCmd.Flags().Int("port_external", 2999, "the ga external service port")
	serveCmd.Flags().Int("port_internal", 2998, "the ga internal service port")
	serveCmd.Flags().String("service_external", "http://api:3000", "the backend server for external access")
	serveCmd.Flags().String("service_internal", "http://traefik:10080", "the backend server for internal access")
	serveCmd.Flags().String("etcd_endpoints", "etcd:2379", "the etcd endpoints")
}
