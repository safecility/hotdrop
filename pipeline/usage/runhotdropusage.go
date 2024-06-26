package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/setup"
	"github.com/safecility/iot/devices/hotdrop/pipeline/usage/helpers"
	"github.com/safecility/iot/devices/hotdrop/pipeline/usage/server"
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
	defer func(gpsClient *pubsub.Client) {
		err := gpsClient.Close()
		if err != nil {
			log.Err(err).Msg("Error closing pubsub client")
		}
	}(gpsClient)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create pubsub client")
	}
	if gpsClient == nil {
		log.Fatal().Err(err).Msg("Failed to create pubsub client")
		return // this is here so golang doesn't complain about gpsClient being possibly nil
	}

	usageTopic := gpsClient.Topic(config.Topics.Usage)
	exists, err := usageTopic.Exists(ctx)
	if !exists {
		log.Fatal().Str("topic", config.Topics.Usage).Msg("no eastron topic found")
	}

	eastronSubscription := gpsClient.Subscription(config.Subscriptions.Eastron)
	exists, err = eastronSubscription.Exists(ctx)
	if !exists {
		log.Fatal().Str("subscription", config.Subscriptions.Eastron).Msg("no eastron subscription")
	}

	eastronServer := server.NewEastronServer(usageTopic, eastronSubscription)
	eastronServer.Start()
}
