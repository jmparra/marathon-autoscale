package marathon

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rossmerr/marathon-autoscale/configuration"
	"github.com/stretchr/testify/assert"
)

const appsJSON = `{
    "apps": [
        {
            "id": "/product/us-east/service/myapp", 
            "cmd": "env && sleep 60", 
            "constraints": [
                [
                    "hostname", 
                    "UNIQUE", 
                    ""
                ]
            ], 
            "container": null, 
            "cpus": 0.1, 
            "env": {
                "LD_LIBRARY_PATH": "/usr/local/lib/myLib"
            }, 
            "executor": "", 
            "instances": 3, 
            "mem": 5.0, 
            "ports": [
                15092, 
                14566
            ], 
            "tasksRunning": 0, 
            "tasksStaged": 1, 
            "uris": [
                "https://raw.github.com/mesosphere/marathon/master/README.md"
            ], 
            "version": "2014-03-01T23:42:20.938Z"
        }
    ]
}`

func TestFetchApps(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, appsJSON)
	}))
	defer ts.Close()

	conf := &configuration.Configuration{}
	apps, err := FetchApps(ts.URL, conf)

	if err != nil {
		log.Fatal(err)
	}

	for _, app := range apps {
		log.Print(app.ID)
		assert.Equal(t, "/product/us-east/service/myapp", app.ID)
		assert.Equal(t, 3, app.Instances)
	}
}

const tasksJSON = `{
    "tasks": [
        {
            "appId": "/bridged-webapp",
            "healthCheckResults": [
                {
                    "alive": true,
                    "consecutiveFailures": 0,
                    "firstSuccess": "2014-10-03T22:57:02.246Z",
                    "lastFailure": null,
                    "lastSuccess": "2014-10-03T22:57:41.643Z",
                    "taskId": "bridged-webapp.eb76c51f-4b4a-11e4-ae49-56847afe9799"
                }
            ],
            "host": "10.141.141.10",
            "id": "bridged-webapp.eb76c51f-4b4a-11e4-ae49-56847afe9799",
            "ports": [
                31000
            ],
            "servicePorts": [
                9000
            ],
            "stagedAt": "2014-10-03T22:16:27.811Z",
            "startedAt": "2014-10-03T22:57:41.587Z",
            "version": "2014-10-03T22:16:23.634Z"
        },
        {
            "appId": "/bridged-webapp",
            "healthCheckResults": [
                {
                    "alive": true,
                    "consecutiveFailures": 0,
                    "firstSuccess": "2014-10-03T22:57:02.246Z",
                    "lastFailure": null,
                    "lastSuccess": "2014-10-03T22:57:41.649Z",
                    "taskId": "bridged-webapp.ef0b5d91-4b4a-11e4-ae49-56847afe9799"
                }
            ],
            "host": "10.141.141.10",
            "id": "bridged-webapp.ef0b5d91-4b4a-11e4-ae49-56847afe9799",
            "ports": [
                31001
            ],
            "servicePorts": [
                9000
            ],
            "stagedAt": "2014-10-03T22:16:33.814Z",
            "startedAt": "2014-10-03T22:57:41.593Z",
            "version": "2014-10-03T22:16:23.634Z"
        }
    ]
}`

func TestFetchTasks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, tasksJSON)
	}))
	defer ts.Close()

	conf := &configuration.Configuration{}
	tasks, err := FetchTasks(ts.URL, conf)

	if err != nil {
		log.Fatal(err)
	}

	for _, task := range tasks {
		assert.Equal(t, "/bridged-webapp", task.AppID)
		assert.Equal(t, true, task.HealthCheckResults[0].Alive)
	}
}
