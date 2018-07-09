// MQTT Exporter

package main

import (
	"encoding/json"
	"time"

	"github.com/Nitroman605/SmartPi/src/smartpi"

	log "github.com/Sirupsen/logrus"
	MQTT "github.com/eclipse/paho.mqtt.golang"
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

func publishMQTT(m MQTT.Client, status *bool, t string, v string) bool {
	if *status {
		log.Debugf("  -> ", t, ":", v)
		token := m.Publish(t, 1, false, v)

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
		jsonReadout, _ := json.Marshal(values)
		publishMQTT(mqttclient, &status, c.MQTTtopic, string(jsonReadout))

		log.Debug("MQTT done.")
	}
}
