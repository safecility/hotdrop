package store

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/safecility/iot/devices/hotdrop/process/messages"
	"time"
)

type DeviceCache struct {
	rClient         *redis.Client
	dbClient        DeviceStore
	contextDeadline time.Duration
}

func NewDeviceCache(rClient *redis.Client, dbClient DeviceStore, deadline int) *DeviceCache {
	return &DeviceCache{rClient: rClient, dbClient: dbClient, contextDeadline: time.Duration(deadline)}
}

func (dc DeviceCache) GetDevice(uid string) (*messages.PowerDevice, error) {

	ctx, cancel := context.WithTimeout(context.Background(), dc.contextDeadline)

	defer cancel()

	pd := &messages.PowerDevice{}
	err := dc.rClient.MGet(ctx, "pDevice", uid).Scan(pd)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			pd, err = dc.dbClient.GetDevice(uid)
			if err != nil {
				return nil, err
			}
			sc := dc.rClient.MSet(ctx, "pDevice", pd)
			_, err = sc.Result()
			if err != nil {
				log.Warn().Err(err).Msg("Failed to set device to cache")
			}
			return pd, nil
		}
		return nil, err
	}

	return pd, nil
}

func (dc DeviceCache) Close() error {
	defer func(rClient *redis.Client) {
		err := rClient.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close Redis")
		}
	}(dc.rClient)
	return dc.dbClient.Close()
}
