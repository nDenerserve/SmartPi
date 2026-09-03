package devicetoken

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestCreateLookupDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "tokens.json")

	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore on a missing file/dir: %v", err)
	}

	tok, err := s.Create("Node-RED · Heizung", []string{ScopeDigitalOut}, "jens")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !strings.HasPrefix(tok.Secret, Prefix) {
		t.Fatalf("secret %q does not carry prefix %q", tok.Secret, Prefix)
	}
	if !LooksLikeToken(tok.Secret) {
		t.Fatal("LooksLikeToken false for a freshly generated secret")
	}
	if !tok.HasScope(ScopeDigitalOut) || tok.HasScope(ScopeConfigWrite) {
		t.Fatalf("scope check wrong: %+v", tok.Scopes)
	}

	got, ok := s.Lookup(tok.Secret)
	if !ok || got.ID != tok.ID {
		t.Fatalf("Lookup after Create: ok=%v got=%+v", ok, got)
	}

	// A second store instance reading the same path sees the persisted token -
	// this is what makes revocation and the "show it to me again" web UI
	// requirement work across a server restart.
	s2, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore reloading persisted file: %v", err)
	}
	got2, ok := s2.Lookup(tok.Secret)
	if !ok || got2.Secret != tok.Secret || got2.Label != tok.Label {
		t.Fatalf("reloaded store did not round-trip the token: ok=%v got=%+v", ok, got2)
	}

	if err := s.Delete(tok.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok := s.Lookup(tok.Secret); ok {
		t.Fatal("token still found after Delete")
	}
	// Deleting an id that no longer exists must stay harmless (the web UI's
	// "revoke" button can be double-clicked).
	if err := s.Delete(tok.ID); err != nil {
		t.Fatalf("Delete of an already-deleted id: %v", err)
	}

	s3, _ := NewStore(path)
	if _, ok := s3.Lookup(tok.Secret); ok {
		t.Fatal("deleted token reappeared after reload")
	}
}

func TestLookupUnknownSecret(t *testing.T) {
	s, _ := NewStore(filepath.Join(t.TempDir(), "tokens.json"))
	if _, ok := s.Lookup("spat_does-not-exist"); ok {
		t.Fatal("Lookup succeeded for a secret that was never created")
	}
}

func TestValidScope(t *testing.T) {
	if !ValidScope(ScopeDigitalOut) {
		t.Fatal("ScopeDigitalOut should be valid")
	}
	if ValidScope("sudo") {
		t.Fatal("an unknown scope must not validate")
	}
}

func TestTwoTokensAreDistinct(t *testing.T) {
	s, _ := NewStore(filepath.Join(t.TempDir(), "tokens.json"))
	a, err := s.Create("a", nil, "jens")
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Create("b", nil, "jens")
	if err != nil {
		t.Fatal(err)
	}
	if a.ID == b.ID || a.Secret == b.Secret {
		t.Fatalf("two created tokens collided: %+v / %+v", a, b)
	}
}

func TestTouchLastUsedIsThrottled(t *testing.T) {
	s, _ := NewStore(filepath.Join(t.TempDir(), "tokens.json"))
	tok, _ := s.Create("a", nil, "jens")

	s.TouchLastUsed(tok.ID)
	first, _ := s.Lookup(tok.Secret)
	if first.LastUsedAt == nil {
		t.Fatal("LastUsedAt not set after first touch")
	}
	firstStamp := *first.LastUsedAt

	// A touch immediately afterwards must not move LastUsedAt forward - that
	// is the whole point of throttling, otherwise a fast poller rewrites the
	// store on every request.
	s.TouchLastUsed(tok.ID)
	second, _ := s.Lookup(tok.Secret)
	if !second.LastUsedAt.Equal(firstStamp) {
		t.Fatalf("LastUsedAt moved despite touchInterval: %v -> %v", firstStamp, *second.LastUsedAt)
	}
}

func TestConcurrentAccess(t *testing.T) {
	s, _ := NewStore(filepath.Join(t.TempDir(), "tokens.json"))
	tok, _ := s.Create("a", []string{ScopeDigitalOut}, "jens")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Lookup(tok.Secret)
			s.TouchLastUsed(tok.ID)
			s.List()
		}()
	}
	wg.Wait()
}

func TestLooksLikeTokenRejectsJWT(t *testing.T) {
	jwtLike := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJlbmVyc2VydmUifQ.sig"
	if LooksLikeToken(jwtLike) {
		t.Fatal("a JWT-shaped bearer value must not be classified as a device token")
	}
}
