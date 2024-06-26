package store

import (
	"cloud.google.com/go/datastore"
	"context"
	"github.com/rs/zerolog/log"
	"github.com/safecility/iot/devices/hotdrop/pipeline/messagestore/messages"
)

type DatastoreHotdrop struct {
	client *datastore.Client
}

func NewDatastoreHotdrop(client *datastore.Client) (*DatastoreHotdrop, error) {
	rd := &DatastoreHotdrop{client: client}
	return rd, nil
}

// AddHotdropMessage TODO store a single value from the hot mess hotdrop units
func (d *DatastoreHotdrop) AddHotdropMessage(m *messages.HotdropDeviceReading) error {
	ctx := context.Background()
	k := datastore.IncompleteKey("Hotdrop", nil)
	k, err := d.client.Put(ctx, k, m)
	if err != nil {
		return err
	}
	log.Debug().Str("uid", m.DeviceEUI).Msg("putting new Hotdrop message")
	return nil
}
