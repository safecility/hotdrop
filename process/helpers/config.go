package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/setup"
	"os"
)

const (
	OSDeploymentKey = "HOTDROP_DEPLOYMENT"
)

type Config struct {
	ProjectName string `json:"projectName"`
	Sql         struct {
		Config setup.MySQLConfig `json:"config"`
		Secret setup.Secret      `json:"secret"`
	} `json:"sql"`
	Topics struct {
		Uplinks string `json:"uplinks"`
		Hotdrop string `json:"hotdrop"`
	} `json:"topics"`
	Subscriptions struct {
		Uplinks string `json:"uplinks"`
	} `json:"subscriptions"`
	PipeAll bool `json:"pipeAll"`
}

// GetConfig creates a config for the specified deployment
func GetConfig(deployment string) *Config {
	fileName := fmt.Sprintf("%s-config.json", deployment)

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal().Err(err).Msg("could not find config file")
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Err(err).Msg("could not close config during defer")
		}
	}(file)
	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		log.Fatal().Err(err).Str("filename", fileName).Msg("could not decode pubsub config")
	}
	return config
}
