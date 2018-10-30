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
		viper.BindEnv("PORT", "SERVICE", "BACKEND")

		viper.BindPFlag("port", cmd.Flags().Lookup("port"))
		viper.BindPFlag("service", cmd.Flags().Lookup("service"))
		viper.BindPFlag("backend", cmd.Flags().Lookup("backend"))
	},
}

func init() {
	serveCmd.Flags().Int("port", 2999, "port to run ga serve on")
	serveCmd.Flags().StringP("service", "s", "", "the service name")
	serveCmd.Flags().StringP("backend", "b", "", "the backend server")
}
