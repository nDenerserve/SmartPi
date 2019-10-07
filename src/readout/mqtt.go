// MQTT Exporter

package main

import (
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/nDenerserve/SmartPi/src/smartpi"
	log "github.com/sirupsen/logrus"
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
	opts.SetKeepAlive(2 * time.Second) // Min value has to be 2
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
		token := m.Publish(t, 0, false, strconv.FormatFloat(v, 'f', 2, 32))

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

func publishMQTTReadouts(c *smartpi.Config, mqttclient MQTT.Client, values *smartpi.ADE7878Readout, accuvalues *smartpi.ReadoutAccumulator) {
	var pTotalBalanced float64
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
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/Ec"+label, accuvalues.WattHoursConsumed[p])
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/Ep"+label, accuvalues.WattHoursProduced[p])
		}
		pTotalBalanced = 0.0
		for _, p := range smartpi.MainPhases {
			pTotalBalanced = pTotalBalanced + values.ActiveWatts[p]
		}
		publishMQTT(mqttclient, &status, c.MQTTtopic+"/Ptot", pTotalBalanced)
		log.Debug("MQTT done.")
	}
}

func publishMQTTCalculations(c *smartpi.Config, mqttclient MQTT.Client, ec1m float64, ep1m float64, cc float64, pc float64) {

	//[basetopic]/[node]/[keyname]
	// Let's try to (re-)connect if MQTT connection was lost.
	if !mqttclient.IsConnected() {
		if mqtttoken := mqttclient.Connect(); mqtttoken.Wait() && mqtttoken.Error() != nil {
			log.Debugf("Connecting to MQTT broker failed. %q", mqtttoken.Error())
		}
	}
	if mqttclient.IsConnected() {
		log.Debug("Publishing calculations via MQTT...")

		// Status is used to stop MQTT publication sequence in case of first error.
		var status = true
		publishMQTT(mqttclient, &status, c.MQTTtopic+"/Ec1m", ec1m)
		publishMQTT(mqttclient, &status, c.MQTTtopic+"/Ep1m", ep1m)
		publishMQTT(mqttclient, &status, c.MQTTtopic+"/EcTot", cc)
		publishMQTT(mqttclient, &status, c.MQTTtopic+"/EpTot", pc)

		log.Debug("MQTT calculations done.")
	}
}
