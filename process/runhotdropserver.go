package main

import (
	"cloud.google.com/go/pubsub"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"context"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/setup"
	"github.com/safecility/iot/devices/hotdrop/process/helpers"
	"github.com/safecility/iot/devices/hotdrop/process/server"
	"github.com/safecility/iot/devices/hotdrop/process/store"
	"os"
)

func main() {

	ctx := context.Background()

	deployment, isSet := os.LookupEnv(helpers.OSDeploymentKey)
	if !isSet {
		deployment = string(setup.Local)
	}
	config := helpers.GetConfig(deployment)

	gpsClient, err := pubsub.NewClient(ctx, config.ProjectName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create pubsub client")
	}
	if gpsClient == nil {
		log.Fatal().Err(err).Msg("Failed to create pubsub client")
		return
	}

	uplinksSubscription := gpsClient.Subscription(config.Subscriptions.Uplinks)
	exists, err := uplinksSubscription.Exists(ctx)
	if !exists {
		log.Fatal().Str("subscription", config.Subscriptions.Uplinks).Msg("no uplinks subscription")
	}

	hotdropTopic := gpsClient.Topic(config.Topics.Hotdrop)
	exists, err = hotdropTopic.Exists(ctx)
	if !exists {
		log.Fatal().Str("topic", config.Topics.Hotdrop).Msg("no hotdrop topic")
	}
	if err != nil {
		log.Fatal().Err(err).Str("topic", config.Topics.Hotdrop).Msg("could not get topic")
	}
	defer hotdropTopic.Stop()

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
	sqlSecret := setup.GetNewSecrets(config.ProjectName, secretsClient)
	password, err := sqlSecret.GetSecret(config.Sql.Secret)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get secret")
	}
	config.Sql.Config.Password = string(password)

	s, err := setup.NewSafecilitySql(config.Sql.Config)
	if err != nil {
		log.Fatal().Err(err).Msg("could not setup safecility sql")
	}
	c, err := store.NewDeviceSql(s)
	if err != nil {
		log.Fatal().Err(err).Msg("could not setup safecility device sql")
	}

	hotDropServer := server.NewHotDropServer(c, uplinksSubscription, hotdropTopic, config.PipeAll)
	hotDropServer.Start()

}
