package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/ini.v1"
)

func TestEnsureAppKey(t *testing.T) {
	for _, old := range []string{"", LegacyDefaultAppKey} {
		p := &SmartPiConfig{AppKey: old}
		replaced, err := p.ensureAppKey()
		if err != nil || !replaced || p.AppKey == old || len(p.AppKey) < 40 {
			t.Errorf("key %q: replaced %v, new %q, %v", old, replaced, p.AppKey, err)
		}
	}
	p := &SmartPiConfig{AppKey: "own-random-key"}
	if replaced, _ := p.ensureAppKey(); replaced || p.AppKey != "own-random-key" {
		t.Error("an own key must stay")
	}
	a, b := &SmartPiConfig{}, &SmartPiConfig{}
	a.ensureAppKey()
	b.ensureAppKey()
	if a.AppKey == b.AppKey {
		t.Error("keys must be random")
	}
}

func TestWriteConfigFileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smartpi")
	// an existing file readable by everyone is tightened
	if err := os.WriteFile(path, nil, 0644); err != nil {
		t.Fatal(err)
	}
	f := ini.Empty()
	f.Section("webserver").NewKey("appkey", "secret")
	if err := writeConfigFile(f, path); err != nil {
		t.Fatal(err)
	}
	fi, _ := os.Stat(path)
	if fi.Mode().Perm() != 0640 {
		t.Errorf("mode %v, want 0640", fi.Mode().Perm())
	}
	got, _ := ini.Load(path)
	if got.Section("webserver").Key("appkey").String() != "secret" {
		t.Error("content not written")
	}
}
