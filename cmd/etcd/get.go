package etcd

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ooclab/ga/service/etcd"
)

// GetRun run cobra subcommand
func GetRun(cmd *cobra.Command, args []string) {

	if len(args) != 1 {
		logrus.Errorf("just KEY need, args = %v\n", args)
		return
	}

	key := args[0]

	session, err := etcd.GetSession()
	if err != nil {
		logrus.Errorf("get etcd session failed: %s\n", err)
		return
	}
	defer session.Close()
	value, err := session.Get(key)
	if err != nil {
		logrus.Errorf("get etcd key(%s) failed: %s\n", key, err)
	} else {
		fmt.Printf(value)
	}
}
