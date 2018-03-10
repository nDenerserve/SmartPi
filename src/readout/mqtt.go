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

func publishMQTT(m MQTT.Client, status bool, t string, v float64) bool {
	if status {
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
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"I1", values.Current[smartpi.PhaseA])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"I2", values.Current[smartpi.PhaseB])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"I3", values.Current[smartpi.PhaseC])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"I4", values.Current[smartpi.PhaseN])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"V1", values.Voltage[smartpi.PhaseA])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"V2", values.Voltage[smartpi.PhaseB])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"V3", values.Voltage[smartpi.PhaseC])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"P1", values.ActiveWatts[smartpi.PhaseA])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"P2", values.ActiveWatts[smartpi.PhaseB])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"P3", values.ActiveWatts[smartpi.PhaseC])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"COS1", values.CosPhi[smartpi.PhaseA])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"COS2", values.CosPhi[smartpi.PhaseB])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"COS3", values.CosPhi[smartpi.PhaseC])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"F1", values.Frequency[smartpi.PhaseA])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"F2", values.Frequency[smartpi.PhaseB])
		publishMQTT(mqttclient, status, c.MQTTtopic+"/"+"F3", values.Frequency[smartpi.PhaseC])

		log.Debug("MQTT done.")
	}
}
