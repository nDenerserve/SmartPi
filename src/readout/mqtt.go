// MQTT Exporter

package main

import (
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/nDenerserve/SmartPi/src/smartpi"
)

func newMQTTClient(c *smartpi.Config) (mqttclient MQTT.Client) {
	log.Debugf("Connecting to MQTT broker at %s", (c.MQTTbroker + ":" + c.MQTTbrokerport))
	//create a MQTTClientOptions struct setting the broker address, clientid, user and password
	opts := MQTT.NewClientOptions().AddBroker("tcp://" + c.MQTTbroker + ":" + c.MQTTbrokerport)
	opts.SetClientID("SmartPi-" + c.Name)
	opts.SetUsername(c.MQTTuser)
	opts.SetPassword(c.MQTTpass)
	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(3 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetKeepAlive(1 * time.Second)
	opts.SetMaxReconnectInterval(3 * time.Second)
	//create and start a client using the above ClientOptions
	mqttclient = MQTT.NewClient(opts)
	if mqtttoken := mqttclient.Connect(); mqtttoken.Wait() && mqtttoken.Error() != nil {
		//panic(mqtttoken.Error())
		log.Debugf("Connecting to MQTT broker failed. %q", mqtttoken.Error())
	}
	return mqttclient
}

func publishMQTT(m MQTT.Client, status *bool, t string, v float64) bool {
	if *status {
		log.Debugf("  -> ", t, ":", v)
		token := m.Publish(t, 1, false, strconv.FormatFloat(v, 'f', 2, 32))

		if !token.WaitTimeout(2 * time.Second) {
			log.Debugf("  MQTT Timeout. Stopping MQTT sequence.")
			return false
		} else if token.Error() != nil {
			log.Error(token.Error())
			return false
		}
		return true
	}
	return false
}

func publishMQTTReadouts(c *smartpi.Config, mqttclient MQTT.Client, values *smartpi.ADE7878Readout) {
	//[basetopic]/[node]/[keyname]
	// Let's try to (re-)connect if MQTT connection was lost.
	if !mqttclient.IsConnected() {
		if mqtttoken := mqttclient.Connect(); mqtttoken.Wait() && mqtttoken.Error() != nil {
			log.Debugf("Connecting to MQTT broker failed. %q", mqtttoken.Error())
		}
	}
	if mqttclient.IsConnected() {
		log.Debug("Publishing readoputs via MQTT...")

		// Status is used to stop MQTT publication sequence in case of first error.
		var status = true
		publishMQTT(mqttclient, &status, c.MQTTtopic+"/I4", values.Current[smartpi.PhaseN])
		for _, p := range smartpi.MainPhases {
			label := p.PhaseNumber()
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/I"+label, values.Current[p])
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/V"+label, values.Voltage[p])
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/P"+label, values.ActiveWatts[p])
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/COS"+label, values.CosPhi[p])
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/F"+label, values.Frequency[p])
		}

		log.Debug("MQTT done.")
	}
}
