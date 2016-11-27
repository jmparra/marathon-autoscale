package marathon

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"

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

// AppList a array of App's
type AppList []App

func (slice AppList) Len() int {
	return len(slice)
}

func (slice AppList) Less(i, j int) bool {
	return slice[i].ID < slice[j].ID
}

func (slice AppList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type TaskList []Task

type Tasks struct {
	Tasks TaskList `json:"tasks"`
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

func (slice TaskList) Len() int {
	return len(slice)
}

func (slice TaskList) Less(i, j int) bool {
	return slice[i].ID < slice[j].ID
}

func (slice TaskList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type Apps struct {
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

func FetchApps(endpoint string, conf *configuration.Configuration) (map[string]App, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", endpoint+"/v2/apps", nil)
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
	var appResponse Apps

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

func FetchTasks(endpoint string, conf *configuration.Configuration) (map[string]Task, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", endpoint+"/v2/tasks", nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if len(conf.Marathon.User) > 0 && len(conf.Marathon.Password) > 0 {
		req.SetBasicAuth(conf.Marathon.User, conf.Marathon.Password)
	}
	response, err := client.Do(req)

	var tasks Tasks

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

	taskList := tasks.Tasks
	sort.Sort(taskList)

	tasksByID := map[string]Task{}

	for _, taskConfig := range tasks.Tasks {
		tasksByID[taskConfig.ID] = taskConfig
	}

	return tasksByID, nil
}

// func createApps(tasksByID map[string]marathonTaskList, marathonApps map[string]marathonApp) AppList {
// 	apps := AppList{}

// 	for appID, mApp := range marathonApps {

// 		// Try to handle old app id format without slashes
// 		appPath := appID
// 		if !strings.HasPrefix(appID, "/") {
// 			appPath = "/" + appID
// 		}

// 		// build App from marathonApp
// 		app := App{
// 			ID:     appPath,
// 			Env:    mApp.Env,
// 			Labels: mApp.Labels,
// 		}

// 		app.HealthChecks = make([]HealthCheck, 0, len(mApp.HealthChecks))
// 		for _, marathonCheck := range mApp.HealthChecks {
// 			check := HealthCheck{
// 				Protocol:  marathonCheck.Protocol,
// 				Path:      marathonCheck.Path,
// 				PortIndex: marathonCheck.PortIndex,
// 			}
// 			app.HealthChecks = append(app.HealthChecks, check)
// 		}

// 		if len(mApp.Ports) > 0 {
// 			app.ServicePort = mApp.Ports[0]
// 			app.ServicePorts = mApp.Ports
// 		}

// 		// build Tasks for this App
// 		tasks := []Task{}
// 		for _, mTask := range tasksByID[appID] {
// 			if len(mTask.Ports) > 0 {
// 				t := Task{
// 					ID:                 mTask.ID,
// 					Host:               mTask.Host,
// 					Port:               mTask.Ports[0],
// 					Ports:              mTask.Ports,
// 					HealthCheckResults: mTask.HealthCheckResults,
// 				}
// 				tasks = append(tasks, t)
// 			}
// 		}
// 		app.Tasks = tasks

// 		apps = append(apps, app)
// 	}
// 	return apps
// }
