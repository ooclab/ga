package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/ooclab/ga/cmd"
	"github.com/spf13/viper"
)

const programVersion = "0.1.0"

var (
	buildstamp = ""
	githash    = ""
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01/02 15:04:05",
	})

	viper.Set("ProgramVersion", programVersion)
	viper.Set("ProgramBuildStamp", buildstamp)
	viper.Set("ProgramGitHash", githash)
}

func main() {
	cmd.Execute()
}
