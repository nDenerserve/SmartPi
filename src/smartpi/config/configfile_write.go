package config

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

// configFileMode is the mode of the configuration files: they hold
// passwords (FTP, MQTT, InfluxDB) and the signing key of the session
// tokens, so only the owner (smartpi) and its group may read them.
const configFileMode = 0640

// writeConfigFile writes an INI configuration directly to path and
// restricts it to configFileMode. It replaces writing a fixed file in /tmp
// first, which another local user could create beforehand (blocking the
// save or reading the secrets) and which was readable by everyone.
func writeConfigFile(f *ini.File, path string) error {
	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, configFileMode)
	if err != nil {
		return err
	}
	if _, err := file.Write(buf.Bytes()); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	// existing files keep their mode on open; tighten it (possible for the
	// owner only - an error here is not fatal, the content is saved)
	if err := os.Chmod(path, configFileMode); err != nil {
		log.Warnf("%s: cannot restrict the permissions: %v", path, err)
	}
	return nil
}

// LegacyDefaultAppKey is the signing key that was shipped with every
// SmartPi (and is in the public source code). A token signed with it can be
// created by anyone, so it is never used.
const LegacyDefaultAppKey = "ew980723j35h97fqw4!234490#t33465"

// EnsureAppKey gives the device its own random signing key if none is set
// or the legacy default is still in use, and saves it. It returns true if
// the key was replaced: tokens signed with the old key are no longer
// accepted, users log in again.
func (p *SmartPiConfig) EnsureAppKey() (bool, error) {
	replaced, err := p.ensureAppKey()
	if replaced {
		p.SaveParameterToFile()
	}
	return replaced, err
}

func (p *SmartPiConfig) ensureAppKey() (bool, error) {
	if p.AppKey != "" && p.AppKey != LegacyDefaultAppKey {
		return false, nil
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return false, err
	}
	p.AppKey = base64.RawURLEncoding.EncodeToString(b)
	return true, nil
}
