package main

import (
	"github.com/rossmerr/marathon-autoscale/configuration"
	"github.com/rossmerr/marathon-autoscale/services/autoscale"
)

func main() {
	conf := &configuration.Configuration{}
	autoscale.Autoscale(conf)
}
