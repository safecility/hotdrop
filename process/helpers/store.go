package helpers

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/safecility/iot/devices/hotdrop/process/store"
)

func GetStore(config *Config) (store.DeviceStore, error) {
	if config.Store.Rest != nil {
		return store.CreateDeviceClient(config.Store.Rest.Address()), nil
	}
	return getFirestore(config)
}

func getFirestore(config *Config) (*store.DeviceFirestore, error) {
	ctx := context.Background()

	id := config.Store.Firestore.Database

	fsClient, err := firestore.NewClientWithDatabase(ctx, config.ProjectName, *id)
	if err != nil {
		return nil, err
	}

	fs := store.NewDeviceFirestore(fsClient, config.ContextDeadline)

	return fs, err
}
