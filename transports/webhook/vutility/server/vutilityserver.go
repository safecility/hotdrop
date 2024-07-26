package server

import (
	"cloud.google.com/go/pubsub"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/lib"
	"github.com/safecility/go/lib/stream"
	"github.com/safecility/iot/devices/transports/webhook/vutility/messages"
	"io"
	"net/http"
)

const bearerPrefix = "Bearer "

type VutilityServer struct {
	jwtParser *lib.JWTParser
	uplinks   *pubsub.Topic
	port      string
}

func NewVutilityServer(jwtParser *lib.JWTParser, uplinks *pubsub.Topic, port string) VutilityServer {
	return VutilityServer{jwtParser: jwtParser, uplinks: uplinks, port: port}
}

// Start listen at the given port for /vutility messages
func (en *VutilityServer) Start() error {
	handler := http.HandlerFunc(en.handleRequest)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "started")
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("could write to http.ResponseWriter"))
		}
	})

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	http.Handle("/vutility", handler)

	log.Info().Msgf("Starting webhook on port %s", en.port)
	return http.ListenAndServe(":"+en.port, nil)
}

func (en *VutilityServer) handleAuth(r *http.Request) error {
	auth := r.Header.Get("Authorization")
	log.Debug().Interface("header", auth).Msg("auth")

	if auth == "" || len(auth) < len(bearerPrefix) {
		return fmt.Errorf("invalid authorization header")
	}
	token := auth[len(bearerPrefix):]

	claims, err := en.jwtParser.ParseToken(token)
	if err != nil {
		log.Err(err).Msg("could not parse token")
		return err
	}
	// for the moment we're not interested in the claims
	_ = claims
	return nil
}

func (en *VutilityServer) handleRequest(w http.ResponseWriter, r *http.Request) {

	err := en.handleAuth(r)
	if err != nil {
		log.Err(err).Msg("could not handle request")
		return
	}

	all, err := io.ReadAll(r.Body)

	if err != nil {
		log.Err(err).Msg("no body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	go func() {
		vuMessage, grr := messages.DecodeVutilityJson(all)
		if grr != nil {
			log.Err(grr).Str("body", fmt.Sprintf("%s", all)).Msg("error decoding message")
		}
		id, grr := stream.PublishToTopic(vuMessage, en.uplinks)
		if grr != nil {
			log.Err(grr).Msg("could not publish to topic")
		}
		log.Debug().Str("id", *id).Msg("published")
	}()
	w.WriteHeader(200)

	return
}
