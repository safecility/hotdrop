package store

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/safecility/go/lib"
	"github.com/safecility/iot/devices/hotdrop/process/messages"
	"time"
)

type DeviceFirestore struct {
	client          *firestore.Client
	contextDeadline time.Duration
}

func NewDeviceFirestore(client *firestore.Client, deadline int) *DeviceFirestore {
	return &DeviceFirestore{client: client, contextDeadline: time.Duration(deadline)}
}

func (df DeviceFirestore) GetDevice(uid string) (*messages.PowerDevice, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*df.contextDeadline)
	defer cancel()

	m, err := df.client.Collection("device").Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}
	d := &messages.PowerDevice{
		Device: lib.Device{
			DeviceMeta: &lib.DeviceMeta{
				Listing: &lib.Listing{},
				Version: &lib.DeviceVersion{},
			},
		},
	}
	err = m.DataTo(d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (df DeviceFirestore) Close() error {
	return df.client.Close()
}
