package cmd

import (
	"fmt"

	"github.com/ooclab/ga/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ga",
	Run: func(cmd *cobra.Command, args []string) {
		versionInfo := version.Get()
		fmt.Println(versionInfo.GitVersion)
		fmt.Println(versionInfo.BuildDate)
		fmt.Println(versionInfo.GitCommit)
		fmt.Println(versionInfo.GoVersion)
	},
}
