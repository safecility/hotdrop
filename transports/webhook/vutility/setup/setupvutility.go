package main

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/setup"
	"github.com/safecility/iot/devices/transports/webhook/vutility/helpers"
	"os"
	"time"
)

func main() {
	deployment, isSet := os.LookupEnv("Deployment")
	if !isSet {
		deployment = string(setup.Local)
	}

	ctx := context.Background()
	config := helpers.GetConfig(deployment)

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
	ourSecrets := setup.GetNewSecrets(config.ProjectName, secretsClient)
	sigSecret, err := ourSecrets.GetSecret(config.Secret)

	sig := hmac.New(sha256.New, sigSecret)

	hmacSecret := sig.Sum(nil)

	tokenString, err := getTokenString(hmacSecret)

	fo, err := os.Create("jwt.txt")
	if err != nil {
		log.Error().Err(err).Msg("could not create file")
	}
	defer func() {
		if err := fo.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close output file")
		}
	}()
	write, err := fo.WriteString(tokenString)
	if err != nil {
		log.Error().Err(err).Msg("could not write to file")
	} else {
		log.Info().Msgf("wrote %d bytes", write)
	}
}

func getTokenString(hmacSecret []byte) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"service": "vutility",
		"created": now.Format(time.RFC3339),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(hmacSecret)
	if err != nil {
		return "", err
	}

	// check we can parse before exit
	token, err = jwt.Parse(tokenString,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return hmacSecret, nil
		},
	)

	if token == nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["service"], claims["created"])
	} else {
		return "", fmt.Errorf("invalid claims")
	}

	return tokenString, nil
}
