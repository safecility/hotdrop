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

func NewDeviceFirestore(client *firestore.Client, timeout time.Duration) *DeviceFirestore {
	return &DeviceFirestore{client: client, contextDeadline: timeout}
}

func (df DeviceFirestore) GetDevice(uid string) (*messages.PowerDevice, error) {
	ctx, cancel := context.WithTimeout(context.Background(), df.contextDeadline)
	defer cancel()

	m, err := df.client.Collection("device").Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}
	d := lib.Device{
		DeviceMeta: &lib.DeviceMeta{
			Processors: &lib.Processor{},
			Firmware:   &lib.Firmware{},
		},
	}

	pd := &messages.PowerDevice{
		Device:       d,
		PowerProfile: &messages.PowerProfile{},
	}
	err = m.DataTo(pd)

	return pd, err
}

func (df DeviceFirestore) Close() error {
	return df.client.Close()
}
