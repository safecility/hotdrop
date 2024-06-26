package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/setup"
	"os"
)

type Config struct {
	ProjectName string `json:"projectName"`
	Topics      struct {
		Uplinks          string `json:"uplinks"`
		DownlinkReceipts string `json:"downlinkReceipts"`
		Signal           string `json:"signal"`
		Location         string `json:"location"`
	} `json:"topics"`
	Secret setup.Secret `json:"secret"`
}

// GetConfig for Vutility
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
