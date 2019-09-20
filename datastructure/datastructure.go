package datastructure

import (
	"encoding/json"
	"os"

	"go.uber.org/zap"
)

type DB2User struct {
	User string `json:"userid"`
	Pass string `json:"password"`
}

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
	DB2 struct {
		Host   string
		User   string
		Pass   string
		DBName string
	}
}

type RequestResponse struct {
	Headers    []string
	Body       []byte
	StatusCode int
	Error      error
}

// LoadConfiguration is delegated to load the configuration
func LoadConfiguration() Configuration {
	zap.S().Info("VerifyCommandLineInput | Init a new configuration from the conf file")
	filename := `./conf/test.json`
	file, err := os.Open(filename)
	if err != nil {
		zap.S().Info("VerifyCommandLineInput | can't open config file: ", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	cfg := Configuration{}
	err = decoder.Decode(&cfg)
	if err != nil {
		zap.S().Info("VerifyCommandLineInput | can't decode config JSON: ", err)
	}
	zap.S().Info("VerifyCommandLineInput | Conf loaded -> ", cfg)

	return cfg
}
