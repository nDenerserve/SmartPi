package smartpi

import (
	"strconv"
	"time"

	//import the Paho Go MQTT library
	log "github.com/Sirupsen/logrus"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type MQTTExporter struct {
	//TODO: store a local pointer to the configuration
	//config *Config
	//TODO: mode the mqttclient as member into this struct.
}

var mqttclient MQTT.Client

/*
func NewMQTTExporter(c *Config) *MQTTExporter {
	t := new(MQTTExporter)
	t.config = c
	return t
}
*/

func NewMQTTExporter() *MQTTExporter {
	t := new(MQTTExporter)
	return t
}

func (exporter MQTTExporter) Connect(config *Config) {
	//exporter.config = config
	log.Info("MQTTExporter: Connecting to MQTT broker at ", (config.MQTTbroker + ":" + config.MQTTbrokerport))
	//create a MQTTClientOptions struct setting the broker address, clientid, user and password
	opts := MQTT.NewClientOptions().AddBroker("tcp://" + config.MQTTbroker + ":" + config.MQTTbrokerport)
	opts.SetClientID("SmartPi")
	opts.SetUsername(config.MQTTuser)
	opts.SetPassword(config.MQTTpass)
	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(3 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetKeepAlive(1 * time.Second)
	opts.SetMaxReconnectInterval(3 * time.Second)
	//create and start a client using the above ClientOptions
	mqttclient = MQTT.NewClient(opts)
	if mqtttoken := mqttclient.Connect(); mqtttoken.Wait() && mqtttoken.Error() != nil {
		//panic(mqtttoken.Error())
		if config.DebugLevel > 0 {
			log.Warn("MQTTExporter: Connecting to MQTT broker failed.\n   ", mqtttoken.Error())
		}
	}
}

func (exporter MQTTExporter) PublishReadouts(config *Config, values []float64, readouts []string) {
	//TODO: Check length readouts >= values. If not stop with a warning.
	//TODO: Consider definition of slice of struct for data
	//					- value
	//          - unit
	//          - time series vs accumulated data

	//[basetopic]/[node]/[keyname]
	if config.MQTTenabled {
		// Let's try to (re-)connect if MQTT connection was lost.
		log.Debug("MQTTExporter: Checking connection to MQTT broker.")
		if !mqttclient.IsConnected() {
			log.Debug("MQTTExporter: Reconnectiong")
			if mqtttoken := mqttclient.Connect(); mqtttoken.Wait() && mqtttoken.Error() != nil {
				if config.DebugLevel > 0 {
					log.Warn("MQTTExporter: Re-connecting to MQTT broker failed.\n   ", mqtttoken.Error())
				}
			}
		} else {
			log.Debug("MQTTExporter: Still connected to MQTT broker.")
		}
		if mqttclient.IsConnected() {
			if config.DebugLevel > 0 {
				log.Info("MQTTExporter: Publishing readoputs via MQTT...")
			}

			// Status is used to stop MQTT publication sequence in case of first error.
			var status = true

			for i := 0; i < len(readouts); i++ {
				topic := config.MQTTtopic + "/" + readouts[i]

				if status {
					log.Debug("MQTTExporter:   -> ", topic, ":", values[i])
					token := mqttclient.Publish(topic, 1, false, strconv.FormatFloat(float64(values[i]), 'f', 2, 32))

					if !token.WaitTimeout(2 * time.Second) {
						log.Warn("MQTTExporter: MQTT Timeout. Stopping MQTT sequence.")
						status = false
					} else if token.Error() != nil {
						log.Warn(token.Error())
						status = false
					}
				}
			}
			log.Debug("MQTTExporter: Finished MQTT publication successfully.")
		}
	} else {
		log.Warn("MQTTExporter: Why the hell PublishReadouts is called, when MQTT is not enabled?")
	}
}
