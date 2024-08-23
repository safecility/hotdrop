package messages

import (
	"github.com/safecility/go/lib"
)

type PowerProfile struct {
	PowerFactor float64 `firestore:"powerFactor,omitempty" json:"powerFactor,omitempty"`
	Voltage     float64 `firestore:"voltage,omitempty" json:"voltage,omitempty"`
}

type PowerDevice struct {
	lib.Device
	PowerProfile *PowerProfile `datastore:"-" firestore:"powerProfile,omitempty" json:"powerProfile,omitempty"`
}
