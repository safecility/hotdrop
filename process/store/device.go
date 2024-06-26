package store

import (
	"github.com/safecility/iot/devices/hotdrop/process/messages"
)

type DeviceStore interface {
	GetDevice(uid string) (*messages.PowerDevice, error)
}
