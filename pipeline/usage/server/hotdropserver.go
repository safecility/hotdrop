package server

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/lib/stream"
	"github.com/safecility/iot/devices/hotdrop/pipeline/usage/messages"
	"net/http"
	"os"
)

type HotdropServer struct {
	usageTopic *pubsub.Topic
	sub        *pubsub.Subscription
}

func NewHotdropServer(u *pubsub.Topic, s *pubsub.Subscription) HotdropServer {
	return HotdropServer{usageTopic: u, sub: s}
}

func (es *HotdropServer) Start() {
	go es.receive()
	es.serverHttp()
}

func (es *HotdropServer) receive() {

	err := es.sub.Receive(context.Background(), func(ctx context.Context, message *pubsub.Message) {
		r := &messages.HotdropDeviceReading{}

		log.Debug().Str("data", fmt.Sprintf("%s", message.Data)).Msg("raw data")
		err := json.Unmarshal(message.Data, r)
		message.Ack()
		if err != nil {
			log.Err(err).Msg("could not unmarshall data")
			return
		}

		usage, err := r.Usage()
		if err != nil {
			log.Err(err).Msg("could not get usage")
			return
		}

		topic, err := stream.PublishToTopic(usage, es.usageTopic)
		if err != nil {
			log.Err(err).Msg("could not publish data")
			return
		}
		log.Debug().Str("topic", *topic).Msg("published usage")
	})
	if err != nil {
		log.Err(err).Msg("could not receive from sub")
		return
	}
}

func (es *HotdropServer) serverHttp() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "started")
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("could write to http.ResponseWriter"))
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	log.Debug().Msg(fmt.Sprintf("starting http server port %s", port))
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start http")
	}
}
