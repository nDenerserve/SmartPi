package config

import "testing"

func TestPublicMasksSecrets(t *testing.T) {
	p := &SmartPiConfig{AppKey: "signing-key", FTPpass: "ftp", MQTTpass: "", InfluxAPIToken: "token", FTPuser: "bob"}
	c := p.Public()
	if c.AppKey != "" {
		t.Error("the signing key must not be returned")
	}
	if c.FTPpass != SecretMask || c.InfluxAPIToken != SecretMask {
		t.Errorf("secrets not masked: %q %q", c.FTPpass, c.InfluxAPIToken)
	}
	if c.MQTTpass != "" {
		t.Error("an empty secret stays empty, so the client sees that none is set")
	}
	if c.FTPuser != "bob" {
		t.Error("other fields are returned as they are")
	}
	if p.FTPpass != "ftp" || p.AppKey != "signing-key" {
		t.Error("Public must not change the configuration itself")
	}
}

func TestFilterWrite(t *testing.T) {
	msg := map[string]interface{}{
		"AppKey":   "attacker-key",
		"FTPpass":  SecretMask,
		"MQTTpass": "new",
		"FTPuser":  "bob",
	}
	FilterWrite(msg)
	if _, ok := msg["AppKey"]; ok {
		t.Error("the signing key must not be writable")
	}
	if _, ok := msg["FTPpass"]; ok {
		t.Error("a masked secret sent back must keep the stored value")
	}
	if msg["MQTTpass"] != "new" || msg["FTPuser"] != "bob" {
		t.Errorf("changed values must stay: %v", msg)
	}
}
