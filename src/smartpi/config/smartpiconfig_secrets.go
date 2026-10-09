package config

// SecretMask replaces a stored secret in what the API returns. Secrets are
// never sent to a client: anyone with config:read (every logged-in user
// and device tokens with that scope) could otherwise read the passwords of
// FTP, MQTT and InfluxDB - and the signing key of the session tokens,
// which lets them create an admin session.
const SecretMask = "********"

// SecretFields are the fields of SmartPiConfig that hold passwords or
// tokens. They are masked when read via the API; a write that sends the
// mask back keeps the stored value.
var SecretFields = []string{"Influxpassword", "InfluxAPIToken", "FTPpass", "MQTTpass", "SmartpicloudMQTTpass"}

// ReadOnlyFields can never be changed via the API.
var ReadOnlyFields = []string{"AppKey"}

// Public returns a copy of the configuration that is safe to send to a
// client: secrets that are set are replaced by SecretMask, the signing key
// is left out.
func (p *SmartPiConfig) Public() SmartPiConfig {
	c := *p
	c.AppKey = ""
	for _, s := range []*string{&c.Influxpassword, &c.InfluxAPIToken, &c.FTPpass, &c.MQTTpass, &c.SmartpicloudMQTTpass} {
		if *s != "" {
			*s = SecretMask
		}
	}
	return c
}

// FilterWrite removes from a write request (field name -> value) the
// fields that must not be changed via the API and the secrets that were
// sent back masked, so the stored values stay.
func FilterWrite(msg map[string]interface{}) {
	for _, f := range ReadOnlyFields {
		delete(msg, f)
	}
	for _, f := range SecretFields {
		if v, ok := msg[f].(string); ok && v == SecretMask {
			delete(msg, f)
		}
	}
}
