package configuration

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

/*
	Service configuration struct
*/
type Configuration struct {
	// Marathon configuration
	Marathon Marathon

	// MarathonAutoscale specific configuration
	Autoscale Autoscale
}

/*
	Returns Configuration struct from a given file path

	Parameters:
		filePath: full file path to the JSON configuration
*/
func (config *Configuration) FromFile(filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return json.Unmarshal(content, &config)
}

func FromFile(filePath string) (Configuration, error) {
	conf := &Configuration{}
	err := conf.FromFile(filePath)
	setValueFromEnv(&conf.Marathon.Endpoint, "MARATHON_ENDPOINT")
	setValueFromEnv(&conf.Marathon.User, "MARATHON_USER")
	setValueFromEnv(&conf.Marathon.Password, "MARATHON_PASSWORD")
	setBoolValueFromEnv(&conf.Marathon.UseEventStream, "MARATHON_USE_EVENT_STREAM")

	setValueFromEnv(&conf.Autoscale.Endpoint, "BAMBOO_ENDPOINT")

	return *conf, err
}

func setValueFromEnv(field *string, envVar string) {
	env := os.Getenv(envVar)
	if len(env) > 0 {
		log.Printf("Using environment override %s=%s", envVar, env)
		*field = env
	}
}

func setBoolValueFromEnv(field *bool, envVar string) {
	env := os.Getenv(envVar)
	if len(env) > 0 {
		log.Printf("Using environment override %s=%s", envVar, env)
		x, err := strconv.ParseBool(env)
		if err != nil {
			log.Printf("Error converting boolean value: %s\n", err)
		}
		*field = x
	} else {
		log.Printf("Environment variable not set: %s", envVar)
	}
}
