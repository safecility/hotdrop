package main

import (
	"cloud.google.com/go/pubsub"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"context"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/lib"
	"github.com/safecility/go/setup"
	"github.com/safecility/iot/devices/transports/webhook/vutility/helpers"
	"github.com/safecility/iot/devices/transports/webhook/vutility/server"
	"os"
)

func main() {
	deployment, isSet := os.LookupEnv("Deployment")
	if !isSet {
		deployment = string(setup.Local)
	}
	config := helpers.GetConfig(deployment)

	ctx := context.Background()
	secretsClient, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create secrets client")
	}
	defer func(secretsClient *secretmanager.Client) {
		err := secretsClient.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close secrets client")
		}
	}(secretsClient)
	secrets := setup.GetNewSecrets(config.ProjectName, secretsClient)
	jwtSecret, err := secrets.GetSecret(config.Secret)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get secret")
	}

	psClient, err := pubsub.NewClient(ctx, config.ProjectName)
	if err != nil {
		log.Fatal().Err(err).Msg("could not get pubsub client")
	}
	defer func(psClient *pubsub.Client) {
		err := psClient.Close()
		if err != nil {
			log.Err(err).Msg("Failed to close pubsub client")
		}
	}(psClient)

	uplinksTopic := psClient.Topic(config.Topics.Uplinks)

	exists, err := uplinksTopic.Exists(ctx)
	if !exists || err != nil {
		log.Fatal().Err(err).Str("name", config.Topics.Uplinks).Msg("topic does not exist or error")
	}

	jwtParser := lib.NewJWTParser(string(jwtSecret))

	s := server.NewVutilityServer(&jwtParser, uplinksTopic)

	s.Start()
}
