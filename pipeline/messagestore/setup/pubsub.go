package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/lib/stream"
	"github.com/safecility/go/setup"
	"github.com/safecility/iot/devices/hotdrop/pipeline/messagestore/helpers"
	"os"
	"time"
)

func main() {

	deployment, isSet := os.LookupEnv("Deployment")
	if !isSet {
		deployment = string(setup.Local)
	}
	config := helpers.GetConfig(deployment)

	ctx := context.Background()

	gpsClient, err := pubsub.NewClient(ctx, config.ProjectName)
	if err != nil {
		log.Fatal().Err(err).Msg("could not setup pubsub")
	}

	hotdropSubscription := gpsClient.Subscription(config.Subscriptions.Hotdrop)
	exists, err := hotdropSubscription.Exists(ctx)
	if !exists {
		hotdropTopic := gpsClient.Topic(config.Topics.Hotdrop)
		exists, err = hotdropTopic.Exists(ctx)
		if !exists {
			hotdropTopic, err = gpsClient.CreateTopic(ctx, config.Topics.Hotdrop)
			if err != nil {
				log.Fatal().Err(err).Str("topic", config.Topics.Hotdrop).Msg("setup could not create topic")
			}
			log.Info().Str("topic", hotdropTopic.String()).Msg("created topic")
		}

		r, err := time.ParseDuration("1h")
		if err != nil {
			log.Fatal().Err(err).Msg("could not parse duration")
		}
		subConfig := stream.GetDefaultSubscriptionConfig(hotdropTopic, r)
		hotdropSubscription, err = gpsClient.CreateSubscription(ctx, config.Subscriptions.Hotdrop, subConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("setup could not create subscription")
		}
	}

	log.Info().Msg("setup complete")
}
