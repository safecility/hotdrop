package messages

import (
	"github.com/safecility/go/lib"
)

type PowerProfile struct {
	PowerFactor float64 `firestore:",omitempty"`
	Voltage     float64 `firestore:",omitempty"`
}

type PowerDevice struct {
	lib.Device
	PowerProfile *PowerProfile `datastore:"-" firestore:",omitempty"`
}
