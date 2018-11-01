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
		viper.SetEnvPrefix("GA")
		viper.BindEnv("CONFIG")
		viper.BindPFlag("config-example", cmd.Flags().Lookup("config-example"))
		viper.BindPFlag("config", cmd.Flags().Lookup("config"))
	},
}

func init() {
	serveCmd.Flags().Bool("config-example", false, "dump the example config")
	serveCmd.Flags().StringP("config", "c", "", "the config file")
}
