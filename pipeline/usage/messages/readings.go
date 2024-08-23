package messages

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/safecility/go/lib"
	"time"
)

type PowerProfile struct {
	PowerFactor float64 `firestore:",omitempty" json:"powerFactor,omitempty"`
	Voltage     float64 `firestore:",omitempty" json:"voltage,omitempty"`
}

type PowerDevice struct {
	lib.Device
	PowerProfile *PowerProfile `datastore:"-" firestore:",omitempty"`
}

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

type MeterReading struct {
	lib.Device
	ReadingKWH float64
	Time       time.Time
}

func (mc HotdropDeviceReading) Usage() (*MeterReading, error) {
	if mc.PowerDevice == nil {
		return nil, fmt.Errorf("PowerDevice is required")
	}
	if mc.AccumulatedCurrent.Milli == 0 {
		log.Info().Str("reading", fmt.Sprintf("%+v", mc)).Msg("zero usage - check device is new")
	}
	if mc.PowerProfile == nil {
		log.Warn().Str("UID", mc.DeviceUID).Msg("PowerProfile is missing, using defaults")
		mc.PowerProfile = &PowerProfile{
			PowerFactor: 1,
			Voltage:     230,
		}
	}
	kWh := mc.AccumulatedCurrent.Milli * mc.PowerProfile.PowerFactor * mc.PowerProfile.Voltage / 1000.0
	mr := &MeterReading{
		Device:     mc.Device,
		ReadingKWH: kWh,
		Time:       mc.Time,
	}

	return mr, nil
}
