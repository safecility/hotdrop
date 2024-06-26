package messages

import (
	"time"
)

// HotdropUnits for the slightly weird repetition in different units within the message - the naming of elements is
// also inconsistent so we prefer the milli units where named in the sensor reading
type HotdropUnits struct {
	Milli float64
	Nano  float64
	Base  float64
}

type HotdropReading struct {
	DeviceEUI                   string
	MaximumCurrent              HotdropUnits
	MinimumCurrent              HotdropUnits
	InstantaneousCurrent        HotdropUnits
	AverageCurrent              HotdropUnits
	AccumulatedCurrent          HotdropUnits
	SecondsAgoForMinimumCurrent float64
	SecondsAgoForMaximumCurrent float64
	SupplyVoltage               float64
	Temp                        float64
}

type HotdropDeviceReading struct {
	*PowerDevice `datastore:",omitempty"`
	HotdropReading
	Time time.Time
}
