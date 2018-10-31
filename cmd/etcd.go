package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ooclab/ga/cmd/etcd"
)

func init() {
	etcdCmd.AddCommand(etcdSetCmd)
	etcdCmd.AddCommand(etcdGetCmd)
	rootCmd.AddCommand(etcdCmd)
}

var etcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "manage etcd",
}

func init() {
	etcdSetCmd.Flags().String("endpoints", "127.0.0.1:2379", "the etcd endpoints")
	etcdSetCmd.Flags().Bool("value-is-file", false, "update file content")
	etcdGetCmd.Flags().String("endpoints", "127.0.0.1:2379", "the etcd endpoints")
}

var etcdSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set KEY VALUE",
	Run:   etcd.SetRun,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("etcd_endpoints", cmd.Flags().Lookup("endpoints"))
		viper.BindPFlag("value_is_file", cmd.Flags().Lookup("value-is-file"))
	},
}

var etcdGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get KEY",
	Run:   etcd.GetRun,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("etcd_endpoints", cmd.Flags().Lookup("endpoints"))
	},
}
