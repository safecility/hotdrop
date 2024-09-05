package main

import (
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/lib/gbigquery"
	"github.com/safecility/go/lib/stream"
	"github.com/safecility/go/setup"
	"github.com/safecility/iot/devices/hotdrop/pipeline/bigquery/helpers"
	"os"
	"time"
)

func main() {
	deployment, isSet := os.LookupEnv(helpers.OSDeploymentKey)
	if !isSet {
		deployment = string(setup.Local)
	}
	config := helpers.GetConfig(deployment)

	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, config.ProjectName)
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to BigQuery")
	}
	defer func(client *bigquery.Client) {
		err := client.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close bigquery.Client")
		}
	}(client)

	bqc := gbigquery.NewBQTable(client)

	meta := getTableMetadata(config.BigQuery.Table)
	t, err := bqc.CheckOrCreateBigqueryTable(&config.BigQuery, meta)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create BigQuery table")
	}

	sClient, err := pubsub.NewSchemaClient(ctx, config.ProjectName)
	if err != nil {
		log.Fatal().Err(err).Msg("could not create schema client")
	}
	defer func(sClient *pubsub.SchemaClient) {
		err := sClient.Close()
		if err != nil {
			log.Error().Err(err).Msg("could not close schema client")
		}
	}(sClient)

	schema, err := sClient.Schema(ctx, config.BigQuery.Schema.Name, pubsub.SchemaViewFull)
	if err != nil || schema == nil {
		schema, err = gbigquery.CreateProtoSchema(sClient, config.BigQuery.Schema.Name, config.BigQuery.Schema.FilePath)
		if err != nil {
			log.Fatal().Err(err).Msg("could not create schema")
		}
	}

	gpsClient, err := pubsub.NewClient(ctx, config.ProjectName)
	if err != nil {
		log.Fatal().Err(err).Msg("could not setup pubsub")
	}

	bigqueryTopic := gpsClient.Topic(config.Pubsub.Topics.Bigquery)
	exists, err := bigqueryTopic.Exists(ctx)
	if !exists {
		bigqueryTopic, err = gbigquery.CreateBigqueryTopic(gpsClient, config.Pubsub.Topics.Bigquery, schema)
		if err != nil {
			log.Fatal().Str("sub", config.Pubsub.Subscriptions.BigQuery).Err(err).Msg("could not create bigquery topic")
		}
		log.Info().Msg("bigquery topic created")
	}
	bigQuerySubscription := gpsClient.Subscription(config.Pubsub.Subscriptions.BigQuery)
	exists, err = bigQuerySubscription.Exists(ctx)
	if !exists {
		err = gbigquery.CreateBigQuerySubscription(gpsClient, config.Pubsub.Subscriptions.BigQuery, t.FullID, bigqueryTopic)
		if err != nil {
			log.Fatal().Err(err).Msg("could not create bigquery subscription")
		}
		log.Info().Msg("created bigquery subscription")
	}

	hotdropSubscription := gpsClient.Subscription(config.Pubsub.Subscriptions.Hotdrop)
	exists, err = hotdropSubscription.Exists(ctx)
	if !exists {
		hotdropTopic := gpsClient.Topic(config.Pubsub.Topics.Hotdrop)
		if exists, err = hotdropTopic.Exists(ctx); err != nil {
			log.Fatal().Err(err).Msg("could not check if hotdrop topic exists")
		}
		if !exists {
			hotdropTopic, err = gbigquery.CreateBigqueryTopic(gpsClient, config.Pubsub.Topics.Hotdrop, schema)
			if err != nil {
				log.Fatal().Err(err).Msg("could not create hotdrop topic")
			}
			log.Info().Msg("created hotdrop topic")
		}

		r, err := time.ParseDuration("1h")
		if err != nil {
			log.Fatal().Err(err).Msg("could not parse duration")
		}
		subConfig := stream.GetDefaultSubscriptionConfig(hotdropTopic, r)
		hotdropSubscription, err = gpsClient.CreateSubscription(ctx, config.Pubsub.Subscriptions.Hotdrop, subConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("setup could not create subscription")
		}
		log.Info().Msg("created hotdrop subscription")
	}
	log.Info().Msg("finished pubsub setup")

}

// string   DeviceEUI                    = 1;
// string   Time                         = 2;
// double   Temp                         = 3;
// double   InstantaneousCurrent         = 4;
// double   MaximumCurrent               = 5;
// double   SecondsAgoForMaximumCurrent  = 6;
// double   MinimumCurrent               = 7;
// double   SecondsAgoForMinimumCurrent  = 8;
// double   AccumulatedCurrent           = 9;
// double   SupplyVoltage                = 10;
func getTableMetadata(name string) *bigquery.TableMetadata {
	tableSchema := bigquery.Schema{
		{Name: "DeviceUID", Type: bigquery.StringFieldType},
		{Name: "Time", Type: bigquery.TimestampFieldType},
		{Name: "Temp", Type: bigquery.FloatFieldType},
		{Name: "InstantaneousCurrent", Type: bigquery.FloatFieldType},
		{Name: "MaximumCurrent", Type: bigquery.FloatFieldType},
		{Name: "SecondsAgoForMaximumCurrent", Type: bigquery.FloatFieldType},
		{Name: "MinimumCurrent", Type: bigquery.FloatFieldType},
		{Name: "SecondsAgoForMinimumCurrent", Type: bigquery.FloatFieldType},
		{Name: "AccumulatedCurrent", Type: bigquery.FloatFieldType},
		{Name: "SupplyVoltage", Type: bigquery.FloatFieldType},
	}

	return &bigquery.TableMetadata{
		Name:   name,
		Schema: tableSchema,
	}
}
