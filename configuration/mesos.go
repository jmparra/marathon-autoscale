package configuration

import "strings"

type Mesos struct {
	// comma separated mesos master http endpoints including port number
	Endpoint string
	User     string
	Password string
}

func (m Mesos) Endpoints() []string {
	return strings.Split(m.Endpoint, ",")
}
