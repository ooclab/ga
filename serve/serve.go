package serve

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ooclab/ga/forward"
	"github.com/ooclab/ga/middlewares/auth"
	"github.com/ooclab/ga/middlewares/uid"
)

var middlewareMap = map[string]func(cfg map[string]interface{}) (negroni.Handler, error){
	"uid":  uid.NewMiddleware,
	"auth": auth.NewMiddleware,
}

// Run run cobra subcommand
func Run(cmd *cobra.Command, args []string) {
	if viper.GetBool("config-example") {
		fmt.Printf("%s\n", yamlConfigExample)
		os.Exit(0)
	}

	readConfig()

	servers := viper.GetStringMap("servers")
	if len(servers) == 0 {
		logrus.Warnf("no servers found, quit now")
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup

	for name, _srv := range servers {
		srv := _srv.(map[string]interface{})

		logrus.Debugf("try to run forwarder %s", name)

		// load middlewares
		var middlewares []negroni.Handler
		if v, ok := srv["middlewares"]; ok {
			cfg := v.(map[string]interface{})
			var err error
			middlewares, err = loadMiddlewares(cfg)
			if err != nil {
				logrus.Errorf("load middlewares failed: %s\n", err)
				return
			}
		}

		backend := srv["backend"].(string)
		backendAddr, err := url.Parse(backend)
		if err != nil {
			logrus.Errorf("parse %s failed: %s\n", backend, err)
			return
		}

		listenTo := srv["listen"].(string)
		forwarder := forward.NewHTTPForward(ctx, middlewares, listenTo, backendAddr)

		wg.Add(1)
		go func() {
			defer wg.Done()
			forwarder.Run()
			cancel() // TODO: how to cancel gracefully ?
			logrus.Debugf("forwarder %s quit", name)
		}()
	}

	// go func() {
	// 	defer wg.Done()
	// 	runExternalForward(ctx)
	// 	cancel() // TODO: how to cancel gracefully ?
	// }()
	//
	// go func() {
	// 	defer wg.Done()
	// 	runInternalForward(ctx)
	// 	cancel() // TODO: how to cancel gracefully ?
	// }()
	//
	wg.Wait()
}

func readConfig() {

	// 1. try read config from command lie
	configPath := viper.GetString("config")
	if configPath != "" {
		logrus.Debugf("load config from %s\n", configPath)
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			logrus.Errorf("read config (%s) failed: %s\n", configPath, err)
			os.Exit(2)
		}
		if strings.HasSuffix(configPath, "yml") || strings.HasSuffix(configPath, "yaml") {
			viper.SetConfigType("yaml") // or viper.SetConfigType("YAML")
		} else if strings.HasSuffix(configPath, "json") {
			viper.SetConfigType("json")
		} else if strings.HasSuffix(configPath, "toml") {
			viper.SetConfigType("toml")
		} else {
			logrus.Errorf("unsupported config file type, just [yaml, json, toml]")
			os.Exit(2)
		}
		if err := viper.ReadConfig(bytes.NewBuffer(data)); err != nil {
			logrus.Errorf("load failed: %s\n", err)
		}
		return
	}

	// 2. try read config from default paths

	viper.SetConfigName("config")    // name of config file (without extension)
	viper.AddConfigPath("/etc/ga/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.ga") // call multiple times to add many search paths
	viper.AddConfigPath(".")         // optionally look for config in the working directory
	err := viper.ReadInConfig()      // Find and read the config file
	if err != nil {                  // Handle errors reading the config file
		logrus.Errorf("read config file failed: %s \n", err)
		os.Exit(2)
	}
}

func loadMiddlewares(cfgs map[string]interface{}) ([]negroni.Handler, error) {
	var middlewares []negroni.Handler
	for name, _cfg := range cfgs {
		cfg := _cfg.(map[string]interface{})
		if fc, ok := middlewareMap[name]; ok {
			mw, err := fc(cfg)
			if err != nil {
				logrus.Errorf("load %s middleware failed: %s\n", name, err)
				return nil, err
			}
			middlewares = append(middlewares, mw)
		} else {
			logrus.Warnf("unknown middleware %s, pass\n", name)
		}
	}
	return middlewares, nil
}
