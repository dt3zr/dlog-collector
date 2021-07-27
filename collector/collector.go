package collector

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	HDB_MQTT_BROKER = "tcp://hdb.aztech.com:1883"
	X_TOPIC         = "D/4262086F0D00/#"
	// X_TOPIC = "D/4262086F0D00/4845BD7CCCCC"
)

const (
	CURRENT = 100 + iota
	VOLTAGE
	POWER_FACTOR
	FREQUENCY
	ACTIVE_POWER
	REACTIVE_POWER
	APPARENT_POWER
	ACTIVE_ENERGY
	HARMONICS
	REPORT_INTERVAL
)

var parameterString = map[int]string{
	CURRENT:        "Current",
	VOLTAGE:        "Voltage",
	POWER_FACTOR:   "PowerFactor",
	FREQUENCY:      "Frequency",
	ACTIVE_POWER:   "ActivePower",
	REACTIVE_POWER: "ReactivePower",
	APPARENT_POWER: "ApparentPower",
	ACTIVE_ENERGY:  "ActiveEnergy",
	HARMONICS:      "Harmonics",
}

type Packet struct {
	MacId   string          `json:"macId"`
	DevType string          `json:"devType"`
	Opcode  int             `json:"opcode"`
	Load    json.RawMessage `json:"load"`
}

type ParameterLog struct {
	MacId string
	Name  string
	Value float64
}

func getConnectedMqttClient() (mqtt.Client, error) {
	options := mqtt.NewClientOptions()
	options.AddBroker(HDB_MQTT_BROKER)
	options.SetTLSConfig(&tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert})
	options.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println(fmt.Sprintf("Connected to %s", HDB_MQTT_BROKER))
	})
	client := mqtt.NewClient(options)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}

func New(done <-chan struct{}) <-chan ParameterLog {
	result := make(chan ParameterLog)
	go func() {
		defer close(result)
		// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
		client, err := getConnectedMqttClient()
		if err != nil {
			log.Panicf("%v\n", err)
		}
		defer client.Disconnect(0)

		client.Subscribe(X_TOPIC, 0, func(client mqtt.Client, message mqtt.Message) {
			_, _, opc, err := splitTopic(message.Topic())
			opcode, parseErr := strconv.Atoi(opc)
			if err == nil && parseErr == nil {
				switch opcode {
				case CURRENT, VOLTAGE, POWER_FACTOR, FREQUENCY, ACTIVE_POWER,
					REACTIVE_POWER, APPARENT_POWER, ACTIVE_ENERGY, HARMONICS:
					packet := &Packet{}
					json.Unmarshal(message.Payload(), packet)
					load := make(map[string]string)
					json.Unmarshal(packet.Load, &load)
					if paramName, ok := parameterString[packet.Opcode]; ok {
						if paramValue, ok := load[paramName]; ok {
							if value, err := strconv.ParseFloat(paramValue, 64); err == nil {
								result <- ParameterLog{packet.MacId, paramName, value}
							}
						}
						// log.Printf("[%s] %s = %v (%s)", packet.MacId, p, load[p], load["Unit"])
					}
				}
			}

		})
		// block until done
		<-done
	}()
	return result
}
