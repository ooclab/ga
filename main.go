package main

import (
	"time"

	"github.com/sirupsen/logrus"
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
		TimestampFormat: time.RFC3339Nano,
	})

	viper.Set("program.version", programVersion)
	viper.Set("program.buildstamp", buildstamp)
	viper.Set("program.githash", githash)
}

func main() {
	cmd.Execute()
}
