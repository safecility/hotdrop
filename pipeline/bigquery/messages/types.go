package messages

import "github.com/safecility/go/lib"

type PowerDevice struct {
	*lib.Device
	PowerFactor float64 `datastore:",omitempty"`
	Voltage     float64 `datastore:",omitempty"`
}
