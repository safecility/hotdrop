package server

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/safecility/iot/devices/hotdrop/pipeline/bigquery/messages"
	"github.com/safecility/iot/devices/hotdrop/pipeline/bigquery/protobuffer"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net/http"
	"os"
)

type HotDropServer struct {
	sub      *pubsub.Subscription
	pub      *pubsub.Topic
	encoding pubsub.SchemaEncoding
	storeAll bool
}

func NewHotDropServer(sub *pubsub.Subscription, pub *pubsub.Topic, storeAll bool) *HotDropServer {
	return &HotDropServer{sub: sub, pub: pub, storeAll: storeAll, encoding: pubsub.EncodingBinary}
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
			m := protobuffer.CreateProtobufMessage(r)
			crr := es.publishProtoMessages(m)
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

func (es *HotDropServer) publishProtoMessages(hotDrop *protobuffer.Hotdrop) error {

	var msg []byte
	var err error

	switch es.encoding {
	case pubsub.EncodingBinary:
		msg, err = proto.Marshal(hotDrop)
		if err != nil {
			return fmt.Errorf("proto.Marshal err: %v", err)
		}
	case pubsub.EncodingJSON:
		msg, err = protojson.Marshal(hotDrop)
		if err != nil {
			return fmt.Errorf("protojson.Marshal err: %v", err)
		}
	default:
		return fmt.Errorf("invalid encoding: %v", es.encoding)
	}

	ctx := context.Background()
	result := es.pub.Publish(ctx, &pubsub.Message{
		Data: msg,
	})
	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("result.Get: %v", err)
	}
	log.Debug().Str("message", string(msg)).Msg("Published proto message")
	return nil
}
