package main

import (
	"flag"
	"github.com/golang/glog"
	"os"
	"time"
	"github.com/BurntSushi/toml"
	"github.com/michaeldye/hzn-policy-api/api"
)


func parseConfig(configFile string)(*api.PolicyHandlerConfig) {
	config := &api.PolicyHandlerConfig{"0.0.0.0:8091","/etc/horizon/anax/policy.d/", "", "server.key", "server.crt", false}
	if _, err := os.Stat(configFile); err != nil {
		glog.Fatalf("Config file missing at '%s'", configFile)
	}
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		glog.Fatal(err)
	}
	if config.SecretToken == "" {
		glog.Fatal("SecretToken must be non-empty in config")
	}
	if config.NoSec == true {
		glog.V(5).Info("WARNING: security disabled" )
		os.Stderr.WriteString("WARNING: security disabled\n")
	}

	return config
}


func main() {
	configfile := flag.String("configfile", "config.toml", "TOML formatted configuration file")

	flag.Parse()

	ph := parseConfig(*configfile)

	listen := ph.ListenAddr

	api.Listen(listen, ph)

	// sleep forever
	for {
		time.Sleep(10 * time.Second)
	}
}
