package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/michaeldye/hzn-policy-api/api"
	"time"
)

func main() {
	flag.Parse()

	listen := "0.0.0.0:8091"

	glog.Infof("Passing %v to HTTP API server", listen)
	api.Listen(listen)

	// report forever
	for {
		glog.Infof("Some stats reporting...")
		time.Sleep(10 * time.Second)
	}
}
