/*
    Copyright (C) Jens Ramhorst
	  This file is part of SmartPi.
    SmartPi is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    SmartPi is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.
    You should have received a copy of the GNU General Public License
    along with SmartPi.  If not, see <http://www.gnu.org/licenses/>.
    Diese Datei ist Teil von SmartPi.
    SmartPi ist Freie Software: Sie können es unter den Bedingungen
    der GNU General Public License, wie von der Free Software Foundation,
    Version 3 der Lizenz oder (nach Ihrer Wahl) jeder späteren
    veröffentlichten Version, weiterverbreiten und/oder modifizieren.
    SmartPi wird in der Hoffnung, dass es nützlich sein wird, aber
    OHNE JEDE GEWÄHRLEISTUNG, bereitgestellt; sogar ohne die implizite
    Gewährleistung der MARKTFÄHIGKEIT oder EIGNUNG FÜR EINEN BESTIMMTEN ZWECK.
    Siehe die GNU General Public License für weitere Details.
    Sie sollten eine Kopie der GNU General Public License zusammen mit diesem
    Programm erhalten haben. Wenn nicht, siehe <http://www.gnu.org/licenses/>.
*/

// Package devicetoken implements revocable API tokens for machine clients
// (Node-RED and similar), as opposed to the JWT session tokens issued by
// /api/v1/login for logged-in humans.
//
// A device token is a random value, not a JWT: it carries no signature and
// no claims, so it does not depend on and is not invalidated by a change of
// the session token signing secret (config.SmartPiConfig.AppKey). Its
// validity instead comes entirely from being present in the Store - created
// through the web UI, and revoked by deleting it there. This is what makes
// "generate a token, paste it into Node-RED, delete it later to revoke"
// actually work: revocation only has an effect if every request is checked
// against the store, which is the trade-off made here in exchange for never
// needing a token to expire.
package devicetoken

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Prefix marks a bearer value as a device token rather than a session JWT, so
// the auth middleware can dispatch on it without attempting to parse it as a
// JWT first (a session token always starts with "eyJ", the base64 of the
// {"alg":...} header all JWTs share). It also makes a leaked token
// recognizable to secret scanners, the way GitHub's ghp_ or Stripe's sk_
// prefixes do for their own tokens.
const Prefix = "spat_"

// DefaultPath is where the token store is persisted. It lives under
// /var/smartpi rather than next to /etc/smartpi because /etc/smartpi is
// itself a plain ini file, not a directory - and rather than under
// /var/run/smartpi because that directory is recommended to be mounted as
// tmpfs (see the "For secure 24/7 operation" section of the readme) and
// would lose every issued token on reboot.
const DefaultPath = "/var/smartpi/tokens.json"

// touchInterval bounds how often a token's LastUsedAt is written to disk. A
// Node-RED flow polling once a second must not rewrite the whole store on
// every single request; the in-memory value is always current, only the
// persisted one is allowed to lag by up to this much.
const touchInterval = time.Minute

// Known scopes. A device token is only ever as powerful as the scopes it was
// granted at creation, checked by TokenVerifyMiddleWare against the scope
// required by the route it is presented to.
const (
	ScopeDigitalOut  = "digitalout"
	ScopeAnalogOut   = "analogout"
	ScopeAnalogIn    = "analogin"
	ScopeConfigRead  = "config:read"
	ScopeConfigWrite = "config:write"
	ScopeNetwork     = "network"
)

// Scopes lists every scope a token can be granted, in the order they should
// be offered when creating one.
var Scopes = []string{ScopeDigitalOut, ScopeAnalogOut, ScopeAnalogIn, ScopeConfigRead, ScopeConfigWrite, ScopeNetwork}

// ValidScope reports whether scope is one of the known Scopes.
func ValidScope(scope string) bool {
	for _, s := range Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// Token is one issued device token.
//
// Secret is stored and returned in plain text, deliberately not hashed: it
// has to be shown again in the web UI on demand (so a lost copy does not
// force revoking and re-issuing), and the web UI is already behind a session
// login - a hash would not be reversible for that. Confidentiality of the
// store therefore rests on tokens.json's file permissions, the same way the
// session signing secret in /etc/smartpi already does.
type Token struct {
	ID         string     `json:"id"`
	Secret     string     `json:"secret"`
	Label      string     `json:"label"`
	Scopes     []string   `json:"scopes"`
	CreatedAt  time.Time  `json:"createdAt"`
	CreatedBy  string     `json:"createdBy"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
}

// HasScope reports whether the token was granted scope.
func (t *Token) HasScope(scope string) bool {
	for _, s := range t.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// LooksLikeToken reports whether bearer has the shape of a device token
// rather than a session JWT. It is a plain prefix check, not a store lookup,
// so it can be used to route a request before - or without - touching the
// store at all.
func LooksLikeToken(bearer string) bool {
	return strings.HasPrefix(bearer, Prefix)
}

// Store is a small persistent set of device tokens, safe for concurrent use
// by the HTTP server's goroutines.
type Store struct {
	path string

	mu       sync.RWMutex
	bySecret map[string]*Token
	byID     map[string]*Token
}

// NewStore loads the token store from path, or returns an empty one if the
// file (or any of its parent directories) does not exist yet - the normal
// case on a device that has never issued a token.
func NewStore(path string) (*Store, error) {
	s := &Store{
		path:     path,
		bySecret: map[string]*Token{},
		byID:     map[string]*Token{},
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading device token store: %w", err)
	}

	var tokens []*Token
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("parsing device token store: %w", err)
	}
	for _, t := range tokens {
		s.bySecret[t.Secret] = t
		s.byID[t.ID] = t
	}
	return s, nil
}

// Create generates a new token with the given label and scopes, persists the
// store and returns the new token, secret included.
func (s *Store) Create(label string, scopes []string, createdBy string) (*Token, error) {
	id, err := randomHex(8)
	if err != nil {
		return nil, err
	}
	secret, err := randomSecret()
	if err != nil {
		return nil, err
	}

	t := &Token{
		ID:        id,
		Secret:    secret,
		Label:     label,
		Scopes:    scopes,
		CreatedAt: time.Now().UTC(),
		CreatedBy: createdBy,
	}

	s.mu.Lock()
	s.bySecret[t.Secret] = t
	s.byID[t.ID] = t
	s.mu.Unlock()

	if err := s.save(); err != nil {
		return nil, err
	}
	return t, nil
}

// List returns every token, oldest first, in the shape the web UI displays.
func (s *Store) List() []*Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sortedLocked()
}

// Delete removes a token by id. Deleting an id that does not (or no longer)
// exist is not an error, so a double click on "revoke" in the web UI stays
// harmless.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	t, ok := s.byID[id]
	if ok {
		delete(s.byID, id)
		delete(s.bySecret, t.Secret)
	}
	s.mu.Unlock()

	if !ok {
		return nil
	}
	return s.save()
}

// Lookup finds a token by the secret presented in an Authorization: Bearer
// header. The returned Token is a copy; callers must go through
// TouchLastUsed to record usage.
func (s *Store) Lookup(secret string) (*Token, bool) {
	s.mu.RLock()
	t, ok := s.bySecret[secret]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	cp := *t
	return &cp, true
}

// TouchLastUsed records that the token identified by id was just used to
// authenticate a request. The write to disk is throttled to touchInterval,
// so a device token used on every sample of a fast Node-RED poll does not
// rewrite tokens.json on every request - only the in-memory value is always
// current.
func (s *Store) TouchLastUsed(id string) {
	now := time.Now().UTC()

	s.mu.Lock()
	t, ok := s.byID[id]
	due := false
	if ok {
		due = t.LastUsedAt == nil || now.Sub(*t.LastUsedAt) >= touchInterval
		if due {
			t.LastUsedAt = &now
		}
	}
	s.mu.Unlock()

	if due {
		// Best-effort: a failed write here must not fail the request that
		// triggered it, so the error is only worth knowing to the caller
		// if it wants to log it.
		_ = s.save()
	}
}

// sortedLocked returns every token as a copy, oldest first. Callers must
// hold s.mu.
func (s *Store) sortedLocked() []*Token {
	list := make([]*Token, 0, len(s.byID))
	for _, t := range s.byID {
		cp := *t
		list = append(list, &cp)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.Before(list[j].CreatedAt) })
	return list
}

// save writes the store atomically: to a temp file in the same directory,
// then renamed into place, so a crash or power loss mid-write never leaves
// tokens.json truncated. Mirrors the tempfile-then-rename pattern
// SaveParameterToFile already uses for /etc/smartpi itself.
func (s *Store) save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.sortedLocked(), "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("encoding device token store: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".tokens-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return fmt.Errorf("setting permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("replacing %s: %w", s.path, err)
	}
	return nil
}

func randomSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating token secret: %w", err)
	}
	return Prefix + base64.RawURLEncoding.EncodeToString(b), nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating token id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
