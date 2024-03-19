package smartpidc

import (
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nDenerserve/SmartPi/repository/config"
	log "github.com/sirupsen/logrus"
)

func NewMQTTClient(c *config.DCconfig) (mqttclient mqtt.Client) {
	log.Debugf("Connecting to MQTT broker at %s", (c.MQTTbroker + ":" + c.MQTTbrokerport))
	//create a MQTTClientOptions struct setting the broker address, clientid, user and password
	opts := mqtt.NewClientOptions().AddBroker(c.MQTTbrokerscheme + c.MQTTbroker + ":" + c.MQTTbrokerport)
	opts.SetClientID("SmartPiDC-" + c.Name)
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

func publishMQTT(m mqtt.Client, status *bool, t string, v float64) bool {
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

func PublishMQTTReadouts(c *config.DCconfig, mqttclient mqtt.Client, inputconfig []int, values []float64, power []float64, energyConsumed []float64, energyProduced []float64) {
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

		for i := 0; i < len(values); i++ {
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/"+c.InputName[i], values[i])
		}

		for i := 0; i < len(power); i++ {
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/"+c.PowerName[i], power[i])
		}

		for i := 0; i < len(energyConsumed); i++ {
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/"+c.EnergyConsumptionName[i], energyConsumed[i])
		}

		for i := 0; i < len(energyProduced); i++ {
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/"+c.EnergyProductionName[i], energyProduced[i])
		}

		log.Debug("MQTT done.")
	}
}

func PublishMQTTCalculations(c *config.DCconfig, mqttclient mqtt.Client, energyBalanced []float64) {
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

		for i := 0; i < len(energyBalanced); i++ {
			publishMQTT(mqttclient, &status, c.MQTTtopic+"/"+c.EnergyBalancedName[i], energyBalanced[i])
		}

		log.Debug("MQTT done.")
	}
}
