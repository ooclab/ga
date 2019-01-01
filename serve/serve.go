package serve

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"plugin"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ooclab/ga/forward"
)

var middlewareMap = map[string]func(cfg map[string]interface{}) (negroni.Handler, error){
	"logger": func(map[string]interface{}) (negroni.Handler, error) { return negroni.NewLogger(), nil },
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
			var cfg []interface{}
			switch v.(type) {
			case nil:
				logrus.Debugf("no middlewares found for %s, continue", name)
			case []interface{}:
				cfg = v.([]interface{})
			default:
				logrus.Errorf("unsupport middlewares config type: %T\n", v)
				os.Exit(2)
			}
			var err error
			middlewares, err = loadMiddlewares(cfg)
			if err != nil {
				logrus.Errorf("load middlewares failed: %s\n", err)
				return
			}
			logrus.Debugf("load middlewares success for %s\n", name)
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

func loadMiddlewares(cfgs []interface{}) ([]negroni.Handler, error) {
	var middlewares []negroni.Handler
	for _, v := range cfgs {
		cfg := map[string]interface{}{}
		switch v.(type) {
		case map[string]interface{}:
			cfg = v.(map[string]interface{})
		case map[interface{}]interface{}:
			for key, value := range v.(map[interface{}]interface{}) {
				cfg[key.(string)] = value
			}
		default:
			logrus.Errorf("unsupport middleware config type: %T\n", v)
			os.Exit(3)
		}

		name := cfg["name"].(string)
		var err error
		var mw negroni.Handler
		if fc, ok := middlewareMap[name]; ok {
			mw, err = fc(cfg)
		} else {
			logrus.Debugf("try load middleware (%s) as plugin\n", name)
			mw, err = loadPlugin(cfg)
		}
		if err != nil {
			logrus.Errorf("load %s middleware failed: %s\n", name, err)
			return nil, err
		}
		middlewares = append(middlewares, mw)
	}
	return middlewares, nil
}

func loadPlugin(cfg map[string]interface{}) (negroni.Handler, error) {
	name := cfg["name"].(string)

	var err error
	var p *plugin.Plugin
	var path string

	hd, _ := homedir.Expand(fmt.Sprintf("~/.ga/middlewares/%s.so", name))

	for _, path = range []string{
		fmt.Sprintf("middlewares/%s.so", name),
		fmt.Sprintf("/etc/ga/middlewares/%s.so", name),
		hd,
	} {
		p, err = plugin.Open(path)
		if err == nil {
			logrus.Debugf("load plugin (%s) success", path)
			break
		}
		logrus.Debugf("load plugin (%s) failed: %s", path, err)
	}
	if p == nil {
		logrus.Errorf("can not load middleware plugin %s\n", name)
		return nil, errors.New("can not load plugin")
	}

	f, err := p.Lookup("NewMiddleware")
	if err != nil {
		logrus.Errorf("lookup NewMiddleware func from plugin (%s) failed: %s", path, err)
		return nil, err
	}
	return f.(func(map[string]interface{}) (negroni.Handler, error))(cfg)
}
