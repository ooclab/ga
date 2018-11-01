package serve

import (
	"context"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ooclab/ga/forward"
	"github.com/ooclab/ga/middlewares/auth"
	"github.com/ooclab/ga/middlewares/uid"
)

// Run run cobra subcommand
func Run(cmd *cobra.Command, args []string) {
	// check service name
	serviceName := viper.GetString("service")
	if serviceName == "" {
		logrus.Debugf("all settings: \n%s\n", viper.AllSettings())
		logrus.Errorf("the service name must not be empty !")
		os.Exit(1)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		runExternalForward(ctx, serviceName)
		cancel() // TODO: how to cancel gracefully ?
	}()

	go func() {
		defer wg.Done()
		runInternalForward(ctx, serviceName)
		cancel() // TODO: how to cancel gracefully ?
	}()

	wg.Wait()
}

func runExternalForward(ctx context.Context, serviceName string) {
	logrus.Debugf("try to run external forwarder ...")
	// check public key
	pubKey, err := uid.LoadPublicKey()
	if err != nil {
		return
	}

	uidMiddleware := uid.NewMiddleware(pubKey)

	// loads openapi spec
	doc, err := auth.LoadSpec(serviceName)
	if err != nil {
		return
	}

	authMiddleware := auth.NewMiddleware(serviceName, doc)

	middlewares := []negroni.Handler{
		uidMiddleware,
		authMiddleware,
	}
	port := viper.GetInt("port_external")
	backend := viper.GetString("service_external")
	forwarder := forward.NewHTTPForward(ctx, middlewares, port, backend)
	forwarder.Run()
	logrus.Debugf("external forwarder quit")
}

func runInternalForward(ctx context.Context, serviceName string) {
	logrus.Debugf("try to run internal forwarder ...")
	middlewares := []negroni.Handler{}
	port := viper.GetInt("port_internal")
	backend := viper.GetString("service_internal")
	forwarder := forward.NewHTTPForward(ctx, middlewares, port, backend)
	forwarder.Run()
	logrus.Debugf("internal forwarder quit")
}
