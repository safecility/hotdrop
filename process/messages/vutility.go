package messages

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"strings"
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

type VuSensorMessage struct {
	Data []struct {
		DevEui                    string    `json:"devEui"`
		ApiReceivedAt             time.Time `json:"apiReceivedAt"`
		ExternalNetworkType       string    `json:"externalNetworkType"`
		ExternalNetworkName       string    `json:"externalNetworkName"`
		ExternalNetworkReceivedAt time.Time `json:"externalNetworkReceivedAt"`
		Rssi                      float64   `json:"rssi"`
		Snr                       float64   `json:"snr"`
		FrameCount                int       `json:"frameCount"`
		Latitude                  float64   `json:"latitude"`
		Longitude                 float64   `json:"longitude"`
		Altitude                  float64   `json:"altitude"`
		SensorMeasurements        []struct {
			Type  string  `json:"type"`
			Value float64 `json:"value"`
		} `json:"sensorMeasurements"`
	} `json:"data"`
}

func (m *VuSensorMessage) GetHotdropDeviceReadings() (data []HotdropDeviceReading) {

	dataMap := make(map[string]int, len(m.Data))

	for i, d := range m.Data {

		measurements := len(d.SensorMeasurements)

		if measurements == 0 {
			log.Warn().Interface("data", d).Msg("no sensor measurements")
			continue
		}

		key := fmt.Sprintf("%s:%d", d.DevEui, d.ExternalNetworkReceivedAt.Nanosecond())
		_, e := dataMap[key]
		if e {
			log.Warn().Int("count", i).
				Interface("vutility", m).Msg("vutility message contains duplicates")
			continue
		}
		eui := strings.ToLower(d.DevEui)

		hdr := HotdropDeviceReading{
			Time: d.ExternalNetworkReceivedAt,
			HotdropReading: HotdropReading{
				DeviceEUI: eui,
			},
		}
		dataMap[key] = 1

		for _, hdm := range d.SensorMeasurements {
			switch hdm.Type {
			case "temperature_Celsius":
				hdr.Temp = hdm.Value
				break
			case "instantaneousCurrent_MilliAmpere":
				hdr.InstantaneousCurrent.Milli = hdm.Value
				break
			case "maximumCurrent_MilliAmpere":
				hdr.MaximumCurrent.Milli = hdm.Value
				break
			case "maximumCurrent_Ampere":
				hdr.MinimumCurrent.Base = hdm.Value
				break
			case "minimumCurrent_MilliAmpere":
				hdr.MinimumCurrent.Milli = hdm.Value
				break
			case "minimumCurrent_Ampere":
				hdr.MinimumCurrent.Base = hdm.Value
				break
			case "secondsAgoForMaximumCurrent":
				hdr.SecondsAgoForMinimumCurrent = hdm.Value
				break
			case "secondsAgoForMinimumCurrent":
				hdr.SecondsAgoForMinimumCurrent = hdm.Value
				break
			case "accumulatedCurrent_NanoAmpereHour":
				hdr.AccumulatedCurrent.Nano = hdm.Value
				break
			case "accumulatedCurrent_AmpereHour":
				hdr.AccumulatedCurrent.Milli = hdm.Value
				break
			case "supplyVoltage_Volt":
				hdr.SupplyVoltage = hdm.Value
				break
			case "averageCurrent_Ampere":
				hdr.AverageCurrent.Nano = hdm.Value
				break
			case "averageCurrent_AmpereNano":
				hdr.AverageCurrent.Nano = hdm.Value
				break
			case "resistorIndex":
				break
			default:
				log.Warn().Str("type", hdm.Type).Msg("unknown measurement")
			}
		}

		data = append(data, hdr)
	}

	return data
}
