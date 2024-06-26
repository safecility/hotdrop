package messages

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

// VuSensorMessage A general purpose json provided by Vutility for transferring an array of sensor data
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
			Type  string `json:"type"`
			Value any    `json:"value"`
		} `json:"sensorMeasurements"`
	} `json:"data"`
}

func DecodeVutilityJson(data []byte) (*VuSensorMessage, error) {
	var vuMessage VuSensorMessage

	err := json.Unmarshal(data, &vuMessage)
	if err != nil {
		log.Debug().Str("data", fmt.Sprintf("%s", data)).Msg("raw data")
		return nil, err
	}

	if len(vuMessage.Data) > 0 {
		m := vuMessage.Data[0]
		log.Info().Str("DevEui", m.DevEui).Time("received", m.ApiReceivedAt).Msg("vutility")
	}

	for i, m := range vuMessage.Data {
		for j, dv := range m.SensorMeasurements {
			switch dv.Value.(type) {
			case float64:
			default:
				vuMessage.Data[i].SensorMeasurements[j].Value = 0.0
			}
		}
	}

	return &vuMessage, nil
}
