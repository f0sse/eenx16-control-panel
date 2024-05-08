/*
 * heater.go
 * lucas@pamorana.net (2024)
 *
 * MQTT control of a Shelly Plug.
 *
 * This program is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more
 * details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

/* -------------------------------------------------------------------------- *\
|*                                HEATER LOGIC                                *|
\* -------------------------------------------------------------------------- */

package control

import (
	/* standard library */
	"fmt"
	"os"
	"time"
	"errors"
	"strconv"
	"crypto/tls"
	"crypto/x509"

	/* remote modules */
	"github.com/gorilla/websocket"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

/* -------------------------------------------------------------------------- *\
|*                             MQTT CONTROL LOGIC                             *|
\* -------------------------------------------------------------------------- */

/*
 * TOPICS
 */

// reports "on" "off" "overpowered"
const mqtt_topic_status = "shellies/shellyplug-kandidat/relay/0"

// reports instantaneous power consumption in [W]
const mqtt_topic_power = "shellies/shellyplug-kandidat/relay/0/power"

// reports consumed energy in [Wm]
const mqtt_topic_energy = "shellies/shellyplug-kandidat/relay/0/energy"

// "on", "off", "toggle"
const mqtt_topic_command = "shellies/shellyplug-kandidat/relay/0/command"

// reports input events:
//   {"event": ("S", "L", ""),"event_cnt": int}
//
//   S=short,
//   L=long,
//   ''=invalid
const mqtt_topic_event = "shellies/shellyplug-kandidat/input_event/0"

/*
 * CONNECTION DETAILS
 */

const mqtt_cid   = "9013906080"

/* self-signed certificate for the mosquitto (mqtt) server */
const mqtt_cert  = `-----BEGIN CERTIFICATE-----
<REDACTED>
-----END CERTIFICATE-----`

var mqttClient MQTT.Client = nil

func mqttNewTLSConfig() (*tls.Config, error) {
	certpool := x509.NewCertPool()

	if (!certpool.AppendCertsFromPEM([]byte(mqtt_cert))) {
		return nil, errors.New("Certificate could not be parsed")
	}

	tlsconfig := &tls.Config{
		RootCAs: certpool,
		ClientAuth: tls.NoClientCert,
		ClientCAs: nil,
	}

	return tlsconfig, nil
}

func MQTTConnect() (error) {
	if (mqttClient != nil) {
		return errors.New("MQTT already connected once")
	}

	tlsconfig, err := mqttNewTLSConfig()

	if (err != nil) {
		return err
	}

	opts := MQTT.NewClientOptions()

	opts.AddBroker(mqtt_url)
	opts.SetClientID(mqtt_cid)
	opts.SetCleanSession(false)
	opts.SetTLSConfig(tlsconfig)
	opts.SetUsername(os.Getenv("MQTT_USERNAME"))
	opts.SetPassword(os.Getenv("MQTT_PASSWORD"))
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetKeepAlive(21 * time.Second)
	opts.SetOrderMatters(true)

	opts.SetDefaultPublishHandler(mqttHandlePublish)
	opts.SetConnectionLostHandler(mqttOnConnectionLost)
	opts.SetOnConnectHandler(mqttOnConnected)
	opts.SetReconnectingHandler(mqttOnReconnecting)

	mqttClient = MQTT.NewClient(opts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		mqttClient.Disconnect(250)
		token.Error()
	}

	mqttClient.Subscribe(mqtt_topic_status, 0, mqttHandleRelayState)
	mqttClient.Subscribe(mqtt_topic_power,  0, mqttHandleRelayPower)
	mqttClient.Subscribe(mqtt_topic_energy, 0, mqttHandleRelayEnergy)

	return nil
}

func MQTTDisconnect(code uint) {
	if (mqttClient != nil) {
		mqttClient.Disconnect(code)
	}
}

func MQTTHeaterPlugCommand(cmd string) {
	if (mqttClient != nil) {
		mqttClient.Publish(mqtt_topic_command, 0, false, cmd)
	}
}


/* -------------------------------------------------------------------------- *\
|*                               EVENT HANDLERS                               *|
\* -------------------------------------------------------------------------- */

func mqttOnConnectionLost(c MQTT.Client, e error) {
	;
}

func mqttOnConnected(c MQTT.Client) {
	;
}

func mqttOnReconnecting(c MQTT.Client, opts *MQTT.ClientOptions) {
	fmt.Printf("Reconnecting ...\n")
}


/* -------------------------------------------------------------------------- *\
|*                               TOPIC HANDLERS                               *|
\* -------------------------------------------------------------------------- */

var publishDelay *time.Timer = nil

var heaterState HeaterStatus = HeaterStatus{
	Time: time.Now(),
	State: false,
	Power: 0.0,
	Energy: 0.0,
	Temperature: 0.0,
}

type HeaterStatus struct {
	Time        time.Time `json:"time"`        //
	State       bool      `json:"state"`       // "on" "off"
	Power       float64   `json:"power"`       // [W]
	Energy      float64   `json:"energy"`      // [Wh]
	Temperature float64   `json:"temperature"` // [°C]
}

// shelly plug updates are published at the same time every 30 seconds.
// debounce, to avoid 3 websocket publishes at the same time.
func heaterRearmPublish() {
	const duration = time.Millisecond * 100
	if publishDelay == nil {
		publishDelay = time.AfterFunc(duration, func() { PushPlugState(true) })
	} else {
		publishDelay.Reset(duration)
	}
}

func mqttHandlePublish(c MQTT.Client, msg MQTT.Message) {
	// msg.Topic()
	// msg.Payload()
	heaterRearmPublish()
}

func mqttHandleRelayState(c MQTT.Client, msg MQTT.Message) {
	state := msg.Payload()
	heaterState.State = string(state) == "on"
	heaterRearmPublish()
}

func mqttHandleRelayPower(c MQTT.Client, msg MQTT.Message) {
	power := msg.Payload()
	if f, err := strconv.ParseFloat(string(power), 64); err == nil {
		heaterState.Power = f
	}
	heaterRearmPublish()
}

func mqttHandleRelayEnergy(c MQTT.Client, msg MQTT.Message) {
	energy := msg.Payload()
	if f, err := strconv.ParseFloat(string(energy), 64); err == nil {
		// Wm to Wh
		heaterState.Energy = f / 60.0
	}
	heaterRearmPublish()
}


/* -------------------------------------------------------------------------- *\
|*                              WEBSOCKET STREAM                              *|
\* -------------------------------------------------------------------------- */

/* websocket connections */
var HeaterClients = make(map[*websocket.Conn]bool)

func heaterWebSocketConnection(conn *websocket.Conn) {
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			delete(HeaterClients, conn)
			break
		}
	}
}

func PushPlugState(metrics bool) {
	heaterState.Time = time.Now()
	for client := range HeaterClients {
		err := client.WriteJSON(heaterState)
		if err != nil {
			client.Close()
			delete(HeaterClients, client)
		}
	}
	if metrics {
		InfluxWriteHeaterState(heaterState)
	}
}
