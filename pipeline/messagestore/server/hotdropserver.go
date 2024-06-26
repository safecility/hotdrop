package server

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/safecility/iot/devices/hotdrop/pipeline/messagestore/messages"
	"github.com/safecility/iot/devices/hotdrop/pipeline/messagestore/store"
	"net/http"
	"os"
)

type HotDropServer struct {
	store    *store.DatastoreHotdrop
	sub      *pubsub.Subscription
	storeAll bool
}

func NewHotDropServer(store *store.DatastoreHotdrop, sub *pubsub.Subscription, storeAll bool) HotDropServer {

	return HotDropServer{sub: sub, store: store, storeAll: storeAll}
}

func (es *HotDropServer) Start() {
	go es.receive()
	es.serverHttp()
}

func (es *HotDropServer) receive() {

	err := es.sub.Receive(context.Background(), func(ctx context.Context, message *pubsub.Message) {
		r := &messages.HotdropDeviceReading{}

		log.Debug().Str("data", fmt.Sprintf("%s", message.Data)).Msg("raw data")
		err := json.Unmarshal(message.Data, r)
		message.Ack()
		if err != nil {
			log.Err(err).Msg("could not unmarshall data")
			return
		}

		if r.PowerDevice == nil && es.storeAll == false {
			log.Debug().Str("eui", r.DeviceEUI).Msg("skipping message as no device and storeAll == false")
			return
		}

		go func() {
			crr := es.store.AddHotdropMessage(r)
			if crr != nil {
				log.Err(crr).Msg("could not add hotdrop data")
			}
		}()
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
		port = "8081"
	}
	log.Debug().Msg(fmt.Sprintf("starting http server port %s", port))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start http")
	}
}
