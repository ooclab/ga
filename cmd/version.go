package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ga",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Current Version: %s\n", viper.GetString("ProgramVersion"))
		buildstamp := viper.GetString("ProgramBuildStamp")
		if buildstamp != "" {
			fmt.Printf("     Build Time: %s\n", buildstamp)
		}
		githash := viper.GetString("ProgramGitHash")
		if githash != "" {
			fmt.Printf("Git Commit Hash: %s\n", githash)
		}
	},
}
