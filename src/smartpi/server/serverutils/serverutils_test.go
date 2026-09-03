package serverutils

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	"github.com/nDenerserve/SmartPi/smartpi/devicetoken"
)

func testConfig(t *testing.T) *config.SmartPiConfig {
	t.Helper()
	return &config.SmartPiConfig{AppKey: "test-secret-not-the-shipped-default"}
}

func testStore(t *testing.T) *devicetoken.Store {
	t.Helper()
	s, err := devicetoken.NewStore(filepath.Join(t.TempDir(), "tokens.json"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func req(bearer string) *http.Request {
	r, _ := http.NewRequest("GET", "/", nil)
	if bearer != "" {
		r.Header.Set("Authorization", "Bearer "+bearer)
	}
	return r
}

func sessionToken(t *testing.T, conf *config.SmartPiConfig, username string) string {
	t.Helper()
	tok, err := GenerateToken(models.User{Name: username, Role: []string{"smartpi"}}, conf)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

// A session token is unaffected by any of this: it still authenticates as
// the human it names, same as before device tokens existed.
func TestDecryptUserdataFromToken_SessionToken(t *testing.T) {
	conf := testConfig(t)
	tok := sessionToken(t, conf, "jens")

	user, err := DecryptUserdataFromToken(req(tok), conf)
	if err != nil {
		t.Fatalf("valid session token rejected: %v", err)
	}
	if user.Name != "jens" {
		t.Fatalf("got user %+v", user)
	}
}

func TestDecryptUserdataFromToken_NoHeader(t *testing.T) {
	conf := testConfig(t)
	if _, err := DecryptUserdataFromToken(req(""), conf); err == nil {
		t.Fatal("missing Authorization header was accepted")
	}
}

func TestIsDeviceToken(t *testing.T) {
	conf := testConfig(t)
	store := testStore(t)
	tok, _ := store.Create("node-red", []string{devicetoken.ScopeDigitalOut}, "jens")
	session := sessionToken(t, conf, "jens")

	if !IsDeviceToken(req(tok.Secret)) {
		t.Fatal("a spat_ bearer value was not recognized as a device token")
	}
	if IsDeviceToken(req(session)) {
		t.Fatal("a session JWT was misclassified as a device token")
	}
	if IsDeviceToken(req("")) {
		t.Fatal("a missing header was misclassified as a device token")
	}
}

func serveWith(mw http.HandlerFunc, bearer string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req(bearer))
	return rec
}

func TestTokenVerifyMiddleWare_SessionTokenAlwaysSatisfiesScope(t *testing.T) {
	conf := testConfig(t)
	store := testStore(t)
	session := sessionToken(t, conf, "jens")

	called := false
	next := func(http.ResponseWriter, *http.Request) { called = true }

	// A session token was never asked for a digitalout scope and has none -
	// it must still pass, exactly as it did before device tokens existed.
	mw := TokenVerifyMiddleWare(next, conf, store, devicetoken.ScopeConfigWrite)
	rec := serveWith(mw, session)
	if !called || rec.Code != http.StatusOK {
		t.Fatalf("session token rejected by scoped middleware: called=%v code=%d", called, rec.Code)
	}
}

func TestTokenVerifyMiddleWare_DeviceTokenNeedsMatchingScope(t *testing.T) {
	conf := testConfig(t)
	store := testStore(t)
	tok, _ := store.Create("node-red", []string{devicetoken.ScopeDigitalOut}, "jens")

	called := false
	next := func(http.ResponseWriter, *http.Request) { called = true }

	// Right scope: passes.
	mw := TokenVerifyMiddleWare(next, conf, store, devicetoken.ScopeDigitalOut)
	rec := serveWith(mw, tok.Secret)
	if !called || rec.Code != http.StatusOK {
		t.Fatalf("correctly scoped device token rejected: called=%v code=%d", called, rec.Code)
	}

	// Wrong scope: the whole point of this design - a leaked digitalout
	// token must not open config:write.
	called = false
	mw = TokenVerifyMiddleWare(next, conf, store, devicetoken.ScopeConfigWrite)
	rec = serveWith(mw, tok.Secret)
	if called || rec.Code != http.StatusUnauthorized {
		t.Fatalf("device token without the required scope was let through: called=%v code=%d", called, rec.Code)
	}
}

func TestTokenVerifyMiddleWare_DeletedDeviceTokenIsRejected(t *testing.T) {
	conf := testConfig(t)
	store := testStore(t)
	tok, _ := store.Create("node-red", []string{devicetoken.ScopeDigitalOut}, "jens")
	if err := store.Delete(tok.ID); err != nil {
		t.Fatal(err)
	}

	called := false
	next := func(http.ResponseWriter, *http.Request) { called = true }
	mw := TokenVerifyMiddleWare(next, conf, store, devicetoken.ScopeDigitalOut)
	rec := serveWith(mw, tok.Secret)

	// This is the actual feature: revoking in the store must take effect on
	// the very next request, with no window where a deleted token still works.
	if called || rec.Code != http.StatusUnauthorized {
		t.Fatalf("deleted device token still authenticated: called=%v code=%d", called, rec.Code)
	}
}

func TestTokenVerifyMiddleWare_UnknownDeviceTokenIsRejected(t *testing.T) {
	conf := testConfig(t)
	store := testStore(t)

	called := false
	next := func(http.ResponseWriter, *http.Request) { called = true }
	mw := TokenVerifyMiddleWare(next, conf, store, devicetoken.ScopeDigitalOut)
	rec := serveWith(mw, "spat_never-issued-by-this-store")

	if called || rec.Code != http.StatusUnauthorized {
		t.Fatalf("unknown device token authenticated: called=%v code=%d", called, rec.Code)
	}
}

func TestTokenVerifyMiddleWare_ForeignAppKeyIsRejected(t *testing.T) {
	conf := testConfig(t)
	store := testStore(t)

	foreign, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "mallory",
	}).SignedString([]byte("a-different-secret"))

	called := false
	next := func(http.ResponseWriter, *http.Request) { called = true }
	mw := TokenVerifyMiddleWare(next, conf, store, devicetoken.ScopeDigitalOut)
	rec := serveWith(mw, foreign)

	if called || rec.Code != http.StatusUnauthorized {
		t.Fatalf("session token signed with a foreign appkey was accepted: called=%v code=%d", called, rec.Code)
	}
}

// The token-management endpoints are the one place a device token must never
// be accepted, no matter its scope - see the doc comment on
// RequireSessionToken for why.
func TestRequireSessionToken_RejectsDeviceToken(t *testing.T) {
	conf := testConfig(t)
	store := testStore(t)
	tok, _ := store.Create("node-red", devicetoken.Scopes, "jens") // even a token with every scope

	called := false
	next := func(http.ResponseWriter, *http.Request) { called = true }
	mw := RequireSessionToken(next, conf)
	rec := serveWith(mw, tok.Secret)

	if called || rec.Code != http.StatusUnauthorized {
		t.Fatalf("device token was accepted by RequireSessionToken: called=%v code=%d", called, rec.Code)
	}
}

func TestRequireSessionToken_AcceptsSessionToken(t *testing.T) {
	conf := testConfig(t)
	session := sessionToken(t, conf, "jens")

	called := false
	next := func(http.ResponseWriter, *http.Request) { called = true }
	mw := RequireSessionToken(next, conf)
	rec := serveWith(mw, session)

	if !called || rec.Code != http.StatusOK {
		t.Fatalf("session token rejected by RequireSessionToken: called=%v code=%d", called, rec.Code)
	}
}
