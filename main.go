package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/michaeldye/hzn-policy-api/api"
	"time"
)

func main() {
	policyDir := flag.String("policydir", "", "the directory containing policy.d files")

	flag.Parse()

	glog.Infof("Using '%s' for policy.d directory", *policyDir)

	listen := "0.0.0.0:8091"

	api.Listen(listen)

	// report forever
	for {
		time.Sleep(10 * time.Second)
	}
}
