package mesos

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

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
	CPUsLimit          float32 `json:"cpus_limit"`
	CPUsSystemTimeSecs float32 `json:"cpus_system_time_secs"`
	CPUsUserTimeSecs   float32 `json:"cpus_user_time_secs"`
	MemLimitBytes      int     `json:"mem_limit_bytes"`
	MemRssBytes        int     `json:"mem_rss_bytes"`
	Timestamp          float32 `json:"timestamp"`
}

func (s Slave) FetchAgentStatistics() ([]Resource, error) {
	client := &http.Client{}
	endpoint, err := s.Endpoint()
	req, _ := http.NewRequest("GET", "http://"+endpoint+"/monitor/statistics", nil)
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

	err = json.Unmarshal(contents, &resources)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

type slaves struct {
	Slaves []Slave
}

type Slave struct {
	ID                  string         `json:"id"`
	PID                 string         `json:"pid"`
	Hostname            string         `json:"hostname"`
	RegisteredTime      float32        `json:"registered_time"`
	Resources           SlaveResources `json:"resources"`
	UsedResources       SlaveResources `json:"used_resources"`
	OfferedResources    SlaveResources `json:"offered_resources"`
	ReservedResources   SlaveResources `json:"reserved_resources"`
	UnReservedResources SlaveResources `json:"unreserved_resources"`
	Active              bool           `json:"active"`
	Version             string         `json:"version"`
	//	Attributes          []string       `json:"attributes"`
}

func (s Slave) Endpoint() (string, error) {

	index := strings.Index(s.PID, "@")

	if index != -1 {
		substring := s.PID[index+1 : len(s.PID)]
		return substring, nil
	}

	return s.PID, errors.New("Hostname not found within pid")
}

type SlaveResources struct {
	Disk int     `json:"disk"`
	Mem  int     `json:"mem"`
	GPUS float32 `json:"gpus"`
	CPUS float32 `json:"cpus"`
}

func FetchAgents(conf *configuration.Configuration) (map[string]Slave, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", conf.Mesos.Endpoint+"/slaves", nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	var slaves slaves

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &slaves)
	if err != nil {
		return nil, err
	}

	slaveByID := map[string]Slave{}

	for _, slave := range slaves.Slaves {
		slaveByID[slave.ID] = slave
	}

	return slaveByID, nil
}
