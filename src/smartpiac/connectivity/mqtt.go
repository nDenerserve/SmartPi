// MQTT Exporter

package smartpiacConnectivity

import (
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	log "github.com/sirupsen/logrus"

	smartpiacDevice "github.com/nDenerserve/SmartPi/smartpiac/device"
)

// NewMQTTClient connects to the local MQTT broker. A failed connection is only
// logged, because the client reconnects on its own and the publish functions
// retry before every publication.
func NewMQTTClient(c *config.SmartPiConfig) (mqttclient mqtt.Client) {
	log.Debugf("Connecting to MQTT broker at %s", (c.MQTTbroker + ":" + c.MQTTbrokerport))
	//create a MQTTClientOptions struct setting the broker address, clientid, user and password
	opts := mqtt.NewClientOptions().AddBroker(c.MQTTbrokerscheme + c.MQTTbroker + ":" + c.MQTTbrokerport)
	opts.SetClientID("SmartPi-" + c.Name)
	opts.SetUsername(c.MQTTuser)
	opts.SetPassword(c.MQTTpass)
	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(3 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetKeepAlive(2 * time.Second) // Min value has to be 2
	opts.SetMaxReconnectInterval(3 * time.Second)
	//create and start a client using the above ClientOptions
	mqttclient = mqtt.NewClient(opts)
	if mqtttoken := mqttclient.Connect(); mqtttoken.Wait() && mqtttoken.Error() != nil {
		//panic(mqtttoken.Error())
		log.Debugf("Connecting to MQTT broker failed. %q", mqtttoken.Error())
	}
	return mqttclient
}

// publishMQTT publishes a single value to one topic. status is shared across a
// publication sequence: once a publication has failed it stays false and the
// remaining values of that sequence are skipped, so that a broken connection
// does not block the measurement loop for every single topic.
func publishMQTT(m mqtt.Client, qos uint8, status *bool, t string, v float64) bool {
	if *status {
		log.Debug("  -> ", t, ":", v)
		token := m.Publish(t, byte(qos), false, strconv.FormatFloat(v, 'f', 6, 64))
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

// PublishMQTTReadouts publishes one set of readouts to the local broker. The
// values are either a single sample or, if a publication interval is configured,
// the aggregated values of the last interval - the topics are the same in both
// cases.
func PublishMQTTReadouts(c *config.SmartPiConfig, mqttclient mqtt.Client, values *models.ADE7878Readout, wattHourBalanced float64) {
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
		publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/I4", values.Current[models.PhaseN])
		for _, p := range smartpiacDevice.MainPhases {
			label := p.PhaseNumber()
			publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/I"+label, values.Current[p])
			publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/V"+label, values.Voltage[p])
			publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/P"+label, values.ActiveWatts[p])
			publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/COS"+label, values.CosPhi[p])
			publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/F"+label, values.Frequency[p])
			publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/Ec"+label, values.Energyconsumption[p])
			publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/Ep"+label, values.Energyproduction[p])
		}
		pTotalBalanced = 0.0
		for _, p := range smartpiacDevice.MainPhases {
			pTotalBalanced = pTotalBalanced + values.ActiveWatts[p]
		}
		publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/Ptot", pTotalBalanced)
		publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/Ebal", wattHourBalanced)
		log.Debug("MQTT done.")
	}
}

// PublishMQTTCalculations publishes the calculated minute values to the local
// broker: the consumed and produced energy of the last minute (ec1m, ep1m) and
// the totals from the persistent counter files (cc, pc). This always runs once
// per minute, independent of the readout publication interval.
func PublishMQTTCalculations(c *config.SmartPiConfig, mqttclient mqtt.Client, ec1m float64, ep1m float64, cc float64, pc float64) {

	//[basetopic]/[node]/[keyname]
	// Let's try to (re-)connect if MQTT connection was lost.
	if !mqttclient.IsConnected() {
		if mqtttoken := mqttclient.Connect(); mqtttoken.Wait() && mqtttoken.Error() != nil {
			log.Errorf("Connecting to MQTT broker failed. %q", mqtttoken.Error())
		}
	}
	if mqttclient.IsConnected() {
		log.Debug("Publishing calculations via MQTT...")

		// Status is used to stop MQTT publication sequence in case of first error.
		var status = true
		publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/Ec1m", ec1m)
		publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/Ep1m", ep1m)
		publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/EcTot", cc)
		publishMQTT(mqttclient, c.MQTTQoS, &status, c.MQTTtopic+"/EpTot", pc)

		log.Debug("MQTT calculations done.")
	}
}

// NewSmartPicloudMQTTClient connects to the SmartPicloud broker. It behaves
// exactly like NewMQTTClient but uses the cloud credentials.
func NewSmartPicloudMQTTClient(c *config.SmartPiConfig) (mqttclient mqtt.Client) {
	log.Debugf("Connecting to SmartPiMQTT broker at %s", (c.SmartpicloudMQTTbroker + ":" + c.SmartpicloudMQTTbrokerport))
	//create a MQTTClientOptions struct setting the broker address, clientid, user and password
	opts := mqtt.NewClientOptions().AddBroker(c.SmartpicloudMQTTbrokerscheme + c.SmartpicloudMQTTbroker + ":" + c.SmartpicloudMQTTbrokerport)
	opts.SetClientID("SmartPi-" + c.SmartpicloudUsername)
	opts.SetUsername(c.SmartpicloudMQTTuser)
	opts.SetPassword(c.SmartpicloudMQTTpass)
	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(3 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetKeepAlive(2 * time.Second) // Min value has to be 2
	opts.SetMaxReconnectInterval(3 * time.Second)
	//create and start a client using the above ClientOptions
	mqttclient = mqtt.NewClient(opts)
	if mqtttoken := mqttclient.Connect(); mqtttoken.Wait() && mqtttoken.Error() != nil {
		//panic(mqtttoken.Error())
		log.Debugf("Connecting to SmartPIMQTT broker failed. %q", mqtttoken.Error())
	}
	return mqttclient
}

// PublishSmartPicloudMQTTReadouts publishes one set of readouts to SmartPicloud.
// It mirrors PublishMQTTReadouts but uses the cloud topic and QoS, and is
// throttled by its own publication interval.
func PublishSmartPicloudMQTTReadouts(c *config.SmartPiConfig, mqttclient mqtt.Client, values *models.ADE7878Readout, wattHourBalanced float64) {
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
		publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/I4", values.Current[models.PhaseN])
		for _, p := range smartpiacDevice.MainPhases {
			label := p.PhaseNumber()
			publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/I"+label, values.Current[p])
			publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/V"+label, values.Voltage[p])
			publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/P"+label, values.ActiveWatts[p])
			publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/COS"+label, values.CosPhi[p])
			publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/F"+label, values.Frequency[p])
			publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/Ec"+label, values.Energyconsumption[p])
			publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/Ep"+label, values.Energyproduction[p])
		}
		pTotalBalanced = 0.0
		for _, p := range smartpiacDevice.MainPhases {
			pTotalBalanced = pTotalBalanced + values.ActiveWatts[p]
		}
		publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/Ptot", pTotalBalanced)
		publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/Ebal", wattHourBalanced)
		log.Debug("MQTT done.")
	}
}

// PublishSmartPicloudMQTTCalculations publishes the calculated minute values to
// SmartPicloud, see PublishMQTTCalculations.
func PublishSmartPicloudMQTTCalculations(c *config.SmartPiConfig, mqttclient mqtt.Client, ec1m float64, ep1m float64, cc float64, pc float64) {

	//[basetopic]/[node]/[keyname]
	// Let's try to (re-)connect if MQTT connection was lost.
	if !mqttclient.IsConnected() {
		if mqtttoken := mqttclient.Connect(); mqtttoken.Wait() && mqtttoken.Error() != nil {
			log.Errorf("Connecting to MQTT broker failed. %q", mqtttoken.Error())
		}
	}
	if mqttclient.IsConnected() {
		log.Debug("Publishing calculations via MQTT...")

		// Status is used to stop MQTT publication sequence in case of first error.
		var status = true
		publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/Ec1m", ec1m)
		publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/Ep1m", ep1m)
		publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/EcTot", cc)
		publishMQTT(mqttclient, c.SmartpicloudMQTTQoS, &status, c.SmartpicloudMQTTtopic+"/EpTot", pc)

		log.Debug("MQTT calculations done.")
	}
}
