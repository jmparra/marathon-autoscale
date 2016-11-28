package autoscale

import (
	"strconv"
	"time"

	"github.com/rossmerr/marathon-autoscale/configuration"
	"github.com/rossmerr/marathon-autoscale/services/marathon"
	"github.com/rossmerr/marathon-autoscale/services/mesos"
	"github.com/saromanov/go-memdb"
)

type state struct {
	AppID     string
	CPU       float32
	MEM       int
	Limit     int
	Date      time.Time
	Timestamp float32
}

func (s state) Utilization() float32 {
	return float32(100 * (s.MEM / s.Limit))

}

func (s state) Usage(s2 state) float32 {
	cpuDelta := s.CPU - s2.CPU
	timeDelta := s.Timestamp - s2.Timestamp
	return float32(float32(cpuDelta/timeDelta) * 100)
}

func Autoscale(conf *configuration.Configuration) error {

	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"state": &memdb.TableSchema{
				Name: "state",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "AppID"},
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)

	if err != nil {
		panic(err)
	}

	for {
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

		// Create a write transaction
		txn := db.Txn(true)

		for _, app := range apps {

			maxMemPercent, err := strconv.Atoi(app.Labels["maxMemPercent"])
			if err != nil {
				continue
			}

			maxCpuTime, err := strconv.Atoi(app.Labels["maxCpuTime"])
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

			autoscaleMultiplier, err := strconv.ParseFloat(app.Labels["autoscaleMultiplier"], 32)
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

			state := state{AppID: app.ID, CPU: 0, MEM: 0, Limit: 0, Timestamp: 0, Date: time.Now()}
			for _, statistic := range statistics {
				state.CPU = state.CPU + statistic.Statistics.CPUsSystemTimeSecs + statistic.Statistics.CPUsUserTimeSecs
				state.MEM = state.MEM + statistic.Statistics.MemRssBytes
				state.Limit = state.Limit + statistic.Statistics.MemLimitBytes
				state.Timestamp = state.Timestamp + statistic.Statistics.Timestamp
			}

			if err := txn.Insert("state", state); err != nil {
				panic(err)
			}
		}

		// Commit the transaction
		txn.Commit()
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
