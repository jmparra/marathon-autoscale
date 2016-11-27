package autoscale

import (
	"github.com/rossmerr/marathon-autoscale/configuration"
	"github.com/rossmerr/marathon-autoscale/services/marathon"
)

type Autoscale struct {
}

func (autoscale Autoscale) Init() error {
	for {
		conf := &configuration.Configuration{}
		apps, err := marathon.FetchApps(conf)

		if err != nil {
			return err
		}

		for _, app := range apps {
			tasks, err := app.FetchDetails()

			if err != nil {
				return err
			}

			for _, task := range tasks {

			}
		}
	}
	return nil
}
