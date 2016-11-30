package autoscale

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"strings"

	"github.com/rossmerr/marathon-autoscale/configuration"
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
            "version": "2014-03-01T23:42:20.938Z",
            "labels": {
                "maxMemPercent" : "1",
                "maxCPUTime" : "1",
                "maxInstances" : "1"
            }
        }
    ]
}`

const tasksJSON = `{
    "tasks": [
        {
            "appId": "/product/us-east/service/myapp",
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

const slavesJSON = `{
  "slaves": [
    {
      "id": "aa53014e-04cc-49e7-975d-60c635a70c7f-S29",
      "pid": "slave(1)@{{.Endpoint}}",
      "hostname": "10.20.188.205",
      "registered_time": 1480320210.80041,
      "resources": {
        "disk": 4971,
        "mem": 999,
        "gpus": 0,
        "cpus": 1,
        "ports": "[31000-32000]"
      },
      "used_resources": {
        "disk": 0,
        "mem": 768,
        "gpus": 0,
        "cpus": 0.95,
        "ports": "[31016-31016, 31521-31521, 31890-31890]"
      },
      "offered_resources": {
        "disk": 0,
        "mem": 0,
        "gpus": 0,
        "cpus": 0
      },
      "reserved_resources": {},
      "unreserved_resources": {
        "disk": 4971,
        "mem": 999,
        "gpus": 0,
        "cpus": 1,
        "ports": "[31000-32000]"
      },
      "attributes": {
        "role": "general"
      },
      "active": true,
      "version": "1.0.1",
      "reserved_resources_full": {},
      "used_resources_full": [
        {
          "name": "cpus",
          "type": "SCALAR",
          "scalar": {
            "value": 0.95
          },
          "role": "*"
        },
        {
          "name": "mem",
          "type": "SCALAR",
          "scalar": {
            "value": 768
          },
          "role": "*"
        },
        {
          "name": "ports",
          "type": "RANGES",
          "ranges": {
            "range": [
              {
                "begin": 31016,
                "end": 31016
              },
              {
                "begin": 31521,
                "end": 31521
              },
              {
                "begin": 31890,
                "end": 31890
              }
            ]
          },
          "role": "*"
        }
      ],
      "offered_resources_full": []
    }
  ]
}`

const statisticsJSON = `[{
	"executor_id": "smartfocus-api-openid.d2060420-b541-11e6-8310-0efb52840a34",
	"executor_name": "Command Executor (Task: smartfocus-api-openid.d2060420-b541-11e6-8310-0efb52840a34) (Command: NO EXECUTABLE)",
	"framework_id": "aa53014e-04cc-49e7-975d-60c635a70c7f-0001",
	"source": "smartfocus-api-openid.d2060420-b541-11e6-8310-0efb52840a34",
	"statistics": {
		"cpus_limit": 0.6,
		"cpus_system_time_secs": 0.78,
		"cpus_user_time_secs": 4.25,
		"mem_limit_bytes": 301989888,
		"mem_rss_bytes": 160460800,
		"timestamp": 1480333639.83199
	}
}]`

type structure struct {
	Endpoint string
}

func TestAutosacle(t *testing.T) {

	tmpl, err := template.New("slaves").Parse(slavesJSON)
	var json string
	if err != nil {
		fmt.Print(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/v2/apps" {
			fmt.Fprintln(w, appsJSON)
		}
		if r.RequestURI == "/v2/tasks" {
			fmt.Fprintln(w, tasksJSON)
		}
		if r.RequestURI == "/slaves" {
			fmt.Fprintln(w, json)
		}
		if r.RequestURI == "/monitor/statistics" {
			fmt.Fprintln(w, statisticsJSON)
		}
	}))
	defer ts.Close()

	var doc bytes.Buffer

	i := strings.Index(ts.URL, "/")
	endpoint := ts.URL[i+2 : len(ts.URL)]

	err = tmpl.Execute(&doc, structure{Endpoint: endpoint})
	if err != nil {
		fmt.Print(err)
	}

	json = doc.String()

	conf := &configuration.Configuration{}
	conf.Marathon.Endpoint = ts.URL
	conf.Mesos.Endpoint = ts.URL

	Autoscale(conf)
	fmt.Printf("test")
}
