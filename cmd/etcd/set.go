package etcd

import (
	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ooclab/ga/service/etcd"
)

// SetRun run cobra subcommand
func SetRun(cmd *cobra.Command, args []string) {

	if len(args) != 2 {
		logrus.Errorf("just KEY VALUE need, args = %v\n", args)
		return
	}

	key := args[0]
	value := args[1]

	if viper.GetBool("value_is_file") {
		data, err := ioutil.ReadFile(value)
		if err != nil {
			logrus.Errorf("read file (%s) failed: %s\n", value, err)
		}
		value = string(data)
	}

	session, err := etcd.GetSession()
	if err != nil {
		logrus.Errorf("get etcd session failed: %s\n", err)
		return
	}
	defer session.Close()
	if err := session.Set(key, value); err != nil {
		logrus.Errorf("set etcd key(%s) failed: %s\n", key, err)
	} else {
		logrus.Debugf("set etcd key(%s) success", key)
	}
}
