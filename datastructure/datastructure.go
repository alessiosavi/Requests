package datastructure

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// Configuration is delegated to map the json configuration file to Golang struct
type Configuration struct {
	Version string
	Network struct {
		Host string
		Port int
	}
	Redis struct {
		Host string
		Port int
	}
	Cloudant struct {
		Host   string
		Apikey string
		User   string
	}
}

// Response is delegated to save the necessary information related to an HTTP call
type Response struct {
	Headers    []string
	Body       []byte
	StatusCode int
	Time       time.Duration
	Error      error
}

// LoadConfiguration is delegated to load the configuration
func LoadConfiguration() Configuration {
	log.Println("VerifyCommandLineInput | Init a new configuration from the conf file")
	filename := `./conf/test.json`
	file, err := os.Open(filename)
	if err != nil {
		log.Println("VerifyCommandLineInput | can't open config file: ", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	cfg := Configuration{}
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Println("VerifyCommandLineInput | can't decode config JSON: ", err)
	}
	log.Println("VerifyCommandLineInput | Conf loaded -> ", cfg)

	return cfg
}
