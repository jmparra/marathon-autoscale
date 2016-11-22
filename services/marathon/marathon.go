package marathon

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/marathon-autoscale/configuration"
)

// HealthCheck on the application
type HealthCheck struct {
	// One of TCP, HTTP or COMMAND
	Protocol string
	// The path (if Protocol is HTTP)
	Path string
	// The position of the port targeted in the ports array
	PortIndex int
}

// Task describes an app process running
type Task struct {
	ID    string
	Host  string
	Port  int
	Ports []int
	Alive bool
}

// App may have multiple processes
type App struct {
	ID                  string
	HealthCheckPath     string
	HealthCheckProtocol string
	HealthChecks        []HealthCheck
	Tasks               []Task
	ServicePort         int
	ServicePorts        []int
	Env                 map[string]string
	Labels              map[string]string
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

type marathonTaskList []marathonTask

type marathonTasks struct {
	Tasks marathonTaskList `json:"tasks"`
}

// HealthCheckResult the results of a HealthCheck call to marathon
type HealthCheckResult struct {
	Alive bool
}

type marathonTask struct {
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

func (slice marathonTaskList) Len() int {
	return len(slice)
}

func (slice marathonTaskList) Less(i, j int) bool {
	return slice[i].ID < slice[j].ID
}

func (slice marathonTaskList) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type marathonApps struct {
	Apps []marathonApp `json:"apps"`
}

type marathonApp struct {
	ID           string                `json:"id"`
	HealthChecks []marathonHealthCheck `json:"healthChecks"`
	Ports        []int                 `json:"ports"`
	Env          map[string]string     `json:"env"`
	Labels       map[string]string     `json:"labels"`
}

type marathonHealthCheck struct {
	Path      string `json:"path"`
	Protocol  string `json:"protocol"`
	PortIndex int    `json:"portIndex"`
}

func fetchMarathonApps(endpoint string, conf *configuration.Configuration) (map[string]marathonApp, error) {
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
	var appResponse marathonApps

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &appResponse)
	if err != nil {
		return nil, err
	}

	dataByID := map[string]marathonApp{}

	for _, appConfig := range appResponse.Apps {
		dataByID[appConfig.ID] = appConfig
	}

	return dataByID, nil
}

func fetchTasks(endpoint string, conf *configuration.Configuration) (map[string]marathonTaskList, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", endpoint+"/v2/tasks", nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if len(conf.Marathon.User) > 0 && len(conf.Marathon.Password) > 0 {
		req.SetBasicAuth(conf.Marathon.User, conf.Marathon.Password)
	}
	response, err := client.Do(req)

	var tasks marathonTasks

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

	tasksByID := map[string]marathonTaskList{}
	for _, task := range taskList {
		if tasksByID[task.AppID] == nil {
			tasksByID[task.AppID] = marathonTaskList{}
		}
		tasksByID[task.AppID] = append(tasksByID[task.AppID], task)
	}

	for _, list := range tasksByID {
		sort.Sort(list)
	}

	return tasksByID, nil
}

func calculateTaskHealth(healthCheckResults []HealthCheckResult, healthChecks []marathonHealthCheck) bool {
	//If we don't even have health check results for every health check, don't count the task as healthy
	if len(healthChecks) > len(healthCheckResults) {
		return false
	}
	for _, healthCheck := range healthCheckResults {
		if !healthCheck.Alive {
			return false
		}
	}
	return true
}

func createApps(tasksByID map[string]marathonTaskList, marathonApps map[string]marathonApp) AppList {
	apps := AppList{}

	for appID, mApp := range marathonApps {

		// Try to handle old app id format without slashes
		appPath := appID
		if !strings.HasPrefix(appID, "/") {
			appPath = "/" + appID
		}

		// build App from marathonApp
		app := App{
			ID:                  appPath,
			HealthCheckPath:     parseHealthCheckPath(mApp.HealthChecks),
			HealthCheckProtocol: parseHealthCheckProtocol(mApp.HealthChecks),
			Env:                 mApp.Env,
			Labels:              mApp.Labels,
		}

		app.HealthChecks = make([]HealthCheck, 0, len(mApp.HealthChecks))
		for _, marathonCheck := range mApp.HealthChecks {
			check := HealthCheck{
				Protocol:  marathonCheck.Protocol,
				Path:      marathonCheck.Path,
				PortIndex: marathonCheck.PortIndex,
			}
			app.HealthChecks = append(app.HealthChecks, check)
		}

		if len(mApp.Ports) > 0 {
			app.ServicePort = mApp.Ports[0]
			app.ServicePorts = mApp.Ports
		}

		// build Tasks for this App
		tasks := []Task{}
		for _, mTask := range tasksByID[appID] {
			if len(mTask.Ports) > 0 {
				t := Task{
					ID:    mTask.ID,
					Host:  mTask.Host,
					Port:  mTask.Ports[0],
					Ports: mTask.Ports,
					Alive: calculateTaskHealth(mTask.HealthCheckResults, mApp.HealthChecks),
				}
				tasks = append(tasks, t)
			}
		}
		app.Tasks = tasks

		apps = append(apps, app)
	}
	return apps
}

func parseHealthCheckPath(checks []marathonHealthCheck) string {
	for _, check := range checks {
		if check.Protocol != "HTTP" && check.Protocol != "HTTPS" {
			continue
		}
		return check.Path
	}
	return ""
}

/* maybe combine this with the above? */
func parseHealthCheckProtocol(checks []marathonHealthCheck) string {
	for _, check := range checks {
		if check.Protocol != "HTTP" && check.Protocol != "HTTPS" {
			continue
		}
		return check.Protocol
	}
	return ""
}

// FetchApps returns a struct that describes Marathon current app and their
// sub tasks information.
//
// Parameters:
//	endpoint: Marathon HTTP endpoint, e.g. http://localhost:8080
func FetchApps(maraconf configuration.Marathon, conf *configuration.Configuration) (AppList, error) {

	var applist AppList
	var err error

	// try all configured endpoints until one succeeds
	for _, url := range maraconf.Endpoints() {
		applist, err = fetchApps(url, conf)
		if err == nil {
			return applist, err
		}
	}
	// return last error
	return nil, err
}

func fetchApps(url string, conf *configuration.Configuration) (AppList, error) {
	tasks, err := fetchTasks(url, conf)
	if err != nil {
		return nil, err
	}

	marathonApps, err := fetchMarathonApps(url, conf)
	if err != nil {
		return nil, err
	}

	apps := createApps(tasks, marathonApps)
	sort.Sort(apps)
	return apps, nil
}
