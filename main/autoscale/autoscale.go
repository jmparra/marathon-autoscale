package autoscale

import (
	"github.com/rossmerr/marathon-autoscale/configuration"
	"github.com/rossmerr/marathon-autoscale/services/marathon"
	"github.com/rossmerr/marathon-autoscale/services/mesos"
)

type Autoscale struct {
}

func (autoscale Autoscale) Init() error {
	conf := &configuration.Configuration{}

	for {
		resources := make([]mesos.Resource, 0)

		apps, err := marathon.FetchApps(conf)

		if err != nil {
			return err
		}
		tasks, err := marathon.FetchTasks(conf)

		if err != nil {
			return err
		}
		agents, err := mesos.FetchAgents(conf)

		if err != nil {
			return err
		}

		for _, agent := range agents {
			statistics, err := agent.FetchAgentStatistics()

			if err != nil {
				return err
			}

			resources = append(resources, statistics...)
		}

		for _, app := range apps {

			if err != nil {
				return err
			}

			appTasks := findAppTasks(tasks, func(appID string) bool {
				return app.ID == appID
			})

			// for _, task := range appTasks {
			// 	resourcess, err := mesos.FetchAgentStatistics(task.Host, conf)

			// 	if err != nil {
			// 		return err
			// 	}

			// 	statistics := filterStatistics(resources, func(executorID string) bool {
			// 		return task.ID == executorID
			// 	})

			// 	for _, statistic := range statistics {

			// 	}
			// }
		}
	}
	return nil
}

func findAppTasks(s map[string]marathon.Task, fn func(marathonApp string) bool) []marathon.Task {
	var p []marathon.Task
	for _, v := range s {
		if fn(v.AppID) {
			p = append(p, v)
		}
	}
	return p
}

func filterStatistics(s map[string]mesos.Resource, fn func(marathonApp string) bool) []mesos.Resource {
	var p []mesos.Resource
	for _, v := range s {
		if fn(v.ExecutorID) {
			p = append(p, v)
		}
	}
	return p
}
