package datastructure

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
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

func (resp *Response) Dump() string {
	var sb strings.Builder

	sb.WriteString("=========================\n")
	sb.WriteString("Headers: ")
	sb.WriteString(fmt.Sprintf("%s", resp.Headers))
	sb.WriteString("\n")

	sb.WriteString("Status Code: ")
	sb.WriteString(fmt.Sprintf("%d", resp.StatusCode))
	sb.WriteString("\n")

	sb.WriteString("Time elapsed: ")
	sb.WriteString(fmt.Sprintf("%v", resp.Time))
	sb.WriteString("\n")

	sb.WriteString("Body: ")
	sb.WriteString(fmt.Sprint(string(resp.Body)))
	sb.WriteString("\n")

	sb.WriteString("Error: ")
	sb.WriteString(fmt.Sprintf("%v", resp.Error))
	sb.WriteString("\n")
	sb.WriteString("=========================\n")
	return sb.String()
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
