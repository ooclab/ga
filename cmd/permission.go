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
	cobra.OnInitialize(initPermissionAddConfig)
	permissionAddCmd.Flags().StringP("service_name", "n", "", "The service name")
	permissionAddCmd.Flags().StringP("service_doc", "f", "", "the file path for swaggerui api document")
	permissionAddCmd.MarkFlagRequired("service_name")
	permissionAddCmd.MarkFlagRequired("service_doc")
}

func initPermissionAddConfig() {
	viper.BindPFlag("permission_service_name", permissionAddCmd.Flags().Lookup("service_name"))
	viper.BindPFlag("permission_service_doc", permissionAddCmd.Flags().Lookup("service_doc"))
}

var permissionAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add permission",
	Run:   permission.Run,
}
