package mesos

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"fmt"

	"github.com/rossmerr/marathon-autoscale/configuration"
)

type Resources []Resource

type Resource struct {
	ExecutorID   string     `json:"executor_id"`
	ExecutorName string     `json:"executor_name"`
	FrameworkID  string     `json:"framework_id"`
	Source       string     `json:"source"`
	Statistics   Statistics `json:"statistics"`
}

type Statistics struct {
	CPUsLimit             float32 `json:"cpus_limit"`
	CPUsNrPeriods         int     `json:"cpus_nr_periods"`
	CPUsNrThrottled       int     `json:"cpus_nr_throttled"`
	CPUsSystemTimeSecs    float32 `json:"cpus_system_time_secs"`
	CPUsThrottledTimeSecs float32 `json:"cpus_throttled_time_secs"`
	CPUsUserTimeSecs      float32 `json:"cpus_user_time_secs"`
	MemAnonBytes          int     `json:"mem_anon_bytes"`
	MemFileBytes          int     `json:"mem_file_bytes"`
	MemLimitBytes         int     `json:"mem_limit_bytes"`
	MemMappedFileBytes    int     `json:"mem_mapped_file_bytes"`
	MemRssBytes           int     `json:"mem_rss_bytes"`
	Timestamp             float32 `json:"timestamp"`
}

func FetchAgentStatistics(agent string, conf *configuration.Configuration) (map[string]Resource, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", conf.Mesos.Endpoint+"/slave/"+agent+"/monitor/statistics.json", nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	var resources Resources

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf(string(contents))
	err = json.Unmarshal(contents, &resources)
	if err != nil {
		return nil, err
	}

	resourcesByID := map[string]Resource{}

	for _, resource := range resources {
		resourcesByID[resource.ExecutorID] = resource
	}

	return resourcesByID, nil
}
