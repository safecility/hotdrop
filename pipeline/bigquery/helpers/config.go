package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/lib/gbigquery"
	"os"
)

const (
	OSDeploymentKey = "HOTDROP_DEPLOYMENT"
)

type Config struct {
	ProjectName string `json:"projectName"`
	Pubsub      struct {
		Topics struct {
			Hotdrop  string `json:"hotdrop"`
			Bigquery string `json:"bigquery"`
		} `json:"topics"`
		Subscriptions struct {
			BigQuery string `json:"bigquery"`
			Hotdrop  string `json:"hotdrop"`
		} `json:"subscriptions"`
	} `json:"pubsub"`
	BigQuery gbigquery.BQTableConfig `json:"bigQuery"`
	StoreAll bool                    `json:"storeAll"`
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
