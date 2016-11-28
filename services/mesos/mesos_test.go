package mesos

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"strings"

	"github.com/rossmerr/marathon-autoscale/configuration"
	"github.com/stretchr/testify/assert"
)

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

func TestFetchAgentStatistics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, statisticsJSON)
	}))
	defer ts.Close()

	slave := Slave{}

	index := strings.Index(ts.URL, "//")
	url := ts.URL[index+2 : len(ts.URL)]

	slave.PID = "test@" + url

	resources, err := slave.FetchAgentStatistics()

	if err != nil {
		log.Fatal(err)
	}

	for _, resource := range resources {
		assert.Equal(t, "aa53014e-04cc-49e7-975d-60c635a70c7f-0001", resource.FrameworkID)
	}
}

const slavesJSON = `{
  "slaves": [
    {
      "id": "aa53014e-04cc-49e7-975d-60c635a70c7f-S29",
      "pid": "slave(1)@10.20.188.205:5051",
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

func TestFetchAgents(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, slavesJSON)
	}))
	defer ts.Close()

	conf := &configuration.Configuration{}
	conf.Mesos.Endpoint = ts.URL

	slaves, err := FetchAgents(conf)

	if err != nil {
		log.Fatal(err)
	}

	for _, slave := range slaves {
		assert.Equal(t, "aa53014e-04cc-49e7-975d-60c635a70c7f-S29", slave.ID)
		url, _ := slave.Endpoint()
		assert.Equal(t, "10.20.188.205:5051", url)
	}
}
