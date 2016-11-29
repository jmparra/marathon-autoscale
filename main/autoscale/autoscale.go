package autoscale

import (
	"strconv"

	"github.com/rossmerr/marathon-autoscale/configuration"
	"github.com/rossmerr/marathon-autoscale/services/marathon"
	"github.com/rossmerr/marathon-autoscale/services/mesos"
)

type application struct {
	AppID               string
	MaxMemPercent       int
	MaxCPUTime          int
	MaxInstances        int
	TriggerMode         string
	AutoscaleMultiplier float64
	Statistics          []mesos.Resource
}

func Autoscale(conf *configuration.Configuration) error {

	table := make(map[string]application)
	iteration := int64(0)
	for {
		iteration++

		resources := make([]mesos.Resource, 0)

		apps, err := marathon.FetchApps(conf)
		if err != nil {
			panic(err)
		}

		tasks, err := marathon.FetchTasks(conf)
		if err != nil {
			panic(err)
		}

		agents, err := mesos.FetchAgents(conf)
		if err != nil {
			panic(err)
		}

		for _, agent := range agents {
			statistics, err := agent.FetchAgentStatistics()
			if err != nil {
				return err
			}

			resources = append(resources, statistics...)
		}

		for _, app := range apps {

			maxMemPercent, err := strconv.Atoi(app.Labels["maxMemPercent"])
			if err != nil {
				continue
			}

			maxCPUTime, err := strconv.Atoi(app.Labels["maxCPUTime"])
			if err != nil {
				continue
			}

			maxInstances, err := strconv.Atoi(app.Labels["maxInstances"])
			if err != nil {
				continue
			}

			triggerMode := app.Labels["triggerMode"]
			if triggerMode != "" {
				triggerMode = "both"
			}

			autoscaleMultiplier, err := strconv.ParseFloat(app.Labels["autoscaleMultiplier"], 64)
			if err != nil {
				autoscaleMultiplier = 1.5
			}

			appTasks := findAppTasks(tasks, func(appID string) bool {
				return app.ID == appID
			})

			statistics := filterStatistics(resources, appTasks, func(executorID string) bool {
				for _, task := range appTasks {
					if task.ID == executorID {
						return true
					}
				}
				return false
			})

			application := application{AppID: app.ID, MaxMemPercent: maxMemPercent, MaxCPUTime: maxCPUTime,
				MaxInstances: maxInstances, TriggerMode: triggerMode, AutoscaleMultiplier: autoscaleMultiplier}

			if app1, ok := table[app.ID]; ok {
				application = app1
			}

			application.Statistics = append(application.Statistics, statistics...)

			table[app.ID] = application
		}

		// remove old not running apps
		for id := range table {
			if _, ok := apps[id]; !ok {
				delete(table, id)
				break
			}
		}

		// for _, app := range table {
		// 	for stats := app.Statistics {
		// 		stats..CPUsLimit
		// 	}
		// }
	}
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

func filterStatistics(s []mesos.Resource, m []marathon.Task, fn func(marathonApp string) bool) []mesos.Resource {
	var p []mesos.Resource
	for _, v := range s {
		if fn(v.ExecutorID) {
			p = append(p, v)
		}
	}
	return p
}
