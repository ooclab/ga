package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ooclab/ga/permission"
)

func init() {
	permissionCmd.AddCommand(permissionAddCmd)
	rootCmd.AddCommand(permissionCmd)
}

var permissionCmd = &cobra.Command{
	Use:   "permission",
	Short: "manage permission",
}

func init() {
	permissionAddCmd.Flags().String("etcd_endpoints", "127.0.0.1:2379", "the etcd endpoints")
	permissionAddCmd.Flags().String("service", "", "the service name")
	permissionAddCmd.Flags().String("openapi", "", "the file path for openapi document")
	permissionAddCmd.MarkFlagRequired("service")
	permissionAddCmd.MarkFlagRequired("openapi")
}

var permissionAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add permission",
	Run:   permission.Run,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("service_name", cmd.Flags().Lookup("service"))
		viper.BindPFlag("openapi_path", cmd.Flags().Lookup("openapi"))
		viper.BindPFlag("etcd_endpoints", cmd.Flags().Lookup("etcd_endpoints"))
	},
}
