package server

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/lib/stream"
	"github.com/safecility/iot/devices/hotdrop/process/messages"
	"github.com/safecility/iot/devices/hotdrop/process/store"
	"net/http"
	"os"
)

type HotDropServer struct {
	cache   store.DeviceStore
	sub     *pubsub.Subscription
	hotdrop *pubsub.Topic
	pipeAll bool
}

func NewHotDropServer(cache store.DeviceStore, sub *pubsub.Subscription, hotdrop *pubsub.Topic, pipeAll bool) HotDropServer {
	return HotDropServer{sub: sub, cache: cache, hotdrop: hotdrop, pipeAll: pipeAll}
}

func (es *HotDropServer) Start() {
	go es.receive()
	es.serverHttp()
}

func (es *HotDropServer) receive() {

	err := es.sub.Receive(context.Background(), func(ctx context.Context, message *pubsub.Message) {
		newPowerData := &messages.VuSensorMessage{}

		log.Debug().Str("data", fmt.Sprintf("%s", message.Data)).Msg("raw data")
		err := json.Unmarshal(message.Data, newPowerData)
		message.Ack()
		if err != nil {
			log.Err(err).Msg("could not unmarshall data")
			return
		}
		log.Debug().Int("dataSize", len(newPowerData.Data)).Msg("vu hotdrop readings")

		hdData := newPowerData.GetHotdropDeviceReadings()
		for _, r := range hdData {
			log.Debug().Str("eui", r.DeviceEUI).Msg("hotdrop message")
			var pd *messages.PowerDevice

			if es.cache != nil {
				pd, err = es.cache.GetDevice(r.DeviceEUI)
				if err != nil {
					log.Warn().Err(err).Str("uid", r.DeviceEUI).Msg("could not get device")
				}
				r.PowerDevice = pd
			}
			if r.PowerDevice == nil && !es.pipeAll {
				log.Debug().Str("device", r.DeviceEUI).Msg("no device in cache and pipeAll == false")
				continue
			}

			topic, err := stream.PublishToTopic(r, es.hotdrop)
			if err != nil {
				log.Err(err).Msg("could not publish usage to topic")
				continue
			}
			log.Debug().Str("topic", *topic).Msg("published usage to topic")
		}
	})
	if err != nil {
		log.Err(err).Msg("could not receive from sub")
		return
	}
}

func (es *HotDropServer) serverHttp() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "started")
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("could write to http.ResponseWriter"))
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8089"
	}
	log.Debug().Msg(fmt.Sprintf("starting http server port %s", port))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start http")
	}
}
