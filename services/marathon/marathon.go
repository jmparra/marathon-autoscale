package marathon

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

	"github.com/rossmerr/marathon-autoscale/configuration"
)

// App may have multiple processes
type App struct {
	ID           string            `json:"id"`
	HealthChecks []HealthCheck     `json:"healthChecks"`
	Ports        []int             `json:"ports"`
	Env          map[string]string `json:"env"`
	Labels       map[string]string `json:"labels"`
	Instances    int               `json:"instances"`
	TasksRunning int               `json:"tasksRunning"`
	TasksStaged  int               `json:"tasksStaged"`
}

type taskList []Task

type tasks struct {
	Tasks taskList `json:"tasks"`
}

type Task struct {
	AppID              string
	ID                 string
	Host               string
	Ports              []int
	ServicePorts       []int
	StartedAt          string
	StagedAt           string
	Version            string
	HealthCheckResults []HealthCheckResult
}

type HealthCheckResult struct {
	Alive               bool     `json:"Alive"`
	ConsecutiveFailures int      `json:"consecutiveFailures"`
	FirstSuccess        JSONDate `json:"firstSuccess"`
	LastFailure         JSONDate `json:"lastFailure"`
	LastSuccess         JSONDate `json:"lastSuccess"`
	TaskID              string   `json:"taskId"`
}

type apps struct {
	Apps []App `json:"apps"`
}

// HealthCheck on the application
type HealthCheck struct {
	// The path (if Protocol is HTTP)
	Path string `json:"path"`
	// One of TCP, HTTP or COMMAND
	Protocol string `json:"protocol"`
	// The position of the port targeted in the ports array
	PortIndex int `json:"portIndex"`
}

func FetchApps(conf *configuration.Configuration) (map[string]App, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", conf.Marathon.Endpoint+"/v2/apps", nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if len(conf.Marathon.User) > 0 && len(conf.Marathon.Password) > 0 {
		req.SetBasicAuth(conf.Marathon.User, conf.Marathon.Password)
	}
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	var appResponse apps

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &appResponse)
	if err != nil {
		return nil, err
	}

	dataByID := map[string]App{}

	for _, appConfig := range appResponse.Apps {
		dataByID[appConfig.ID] = appConfig
	}

	return dataByID, nil
}

func FetchTasks(conf *configuration.Configuration) (map[string]Task, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", conf.Marathon.Endpoint+"/v2/tasks", nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if len(conf.Marathon.User) > 0 && len(conf.Marathon.Password) > 0 {
		req.SetBasicAuth(conf.Marathon.User, conf.Marathon.Password)
	}
	response, err := client.Do(req)

	var tasks tasks

	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &tasks)
	if err != nil {
		return nil, err
	}

	tasksByID := map[string]Task{}

	for _, taskConfig := range tasks.Tasks {
		tasksByID[taskConfig.ID] = taskConfig
	}

	return tasksByID, nil
}

func (app App) FetchDetails() (map[string]Task, error) {
	return nil, nil
}

func (app App) ScaleApp(conf *configuration.Configuration, marathonApp string) error {

	autoscaleMultiplier, err := strconv.Atoi(app.Labels["autoscaleMultiplier"])
	if err != nil {
		return err
	}

	maxInstances, err := strconv.Atoi(app.Labels["maxInstances"])
	if err != nil {
		return err
	}

	targetInstancesFloat := float64(app.Instances * autoscaleMultiplier)
	targetInstances := int(math.Ceil(targetInstancesFloat))

	if targetInstances > maxInstances {
		targetInstances = maxInstances
	}

	client := &http.Client{}
	var jsonStr = []byte(`{"instances": ` + strconv.Itoa(targetInstances) + `}`)
	req, _ := http.NewRequest("PUT", conf.Marathon.Endpoint+"/v2/apps/"+app.ID, bytes.NewBuffer(jsonStr))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if len(conf.Marathon.User) > 0 && len(conf.Marathon.Password) > 0 {
		req.SetBasicAuth(conf.Marathon.User, conf.Marathon.Password)
	}
	_, err = client.Do(req)

	if err != nil {
		return err
	}

	return nil
}
