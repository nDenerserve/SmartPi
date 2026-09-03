package serverutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/smartpi/config"
	"github.com/nDenerserve/SmartPi/smartpi/devicetoken"

	"github.com/golang-jwt/jwt/v5"
)

func CompareHashAndPassword(hashedPassword string, password []byte) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), password)

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func GenerateToken(user models.User, conf *config.SmartPiConfig) (string, error) {

	var err error
	secret := conf.AppKey

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Name,
		"role":     user.Role,
		"iss":      "enerserve",
	})

	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		log.Fatal(err)
	}

	// spew.Dump(token)

	return tokenString, nil
}

// bearerToken extracts the token value from an "Authorization: Bearer <value>"
// header. It reports false if the header is missing or not in that exact
// two-part shape.
func bearerToken(r *http.Request) (string, bool) {
	parts := strings.Split(r.Header.Get("Authorization"), " ")
	if len(parts) != 2 {
		return "", false
	}
	return parts[1], true
}

// parseSessionToken parses and verifies bearer as a session JWT signed with
// conf.AppKey. It never accepts a device token (see devicetoken.Prefix) -
// callers that also need to accept those check devicetoken.LooksLikeToken
// first and take a different path entirely.
func parseSessionToken(bearer string, conf *config.SmartPiConfig) (*jwt.Token, error) {
	return jwt.Parse(bearer, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return []byte(conf.AppKey), nil
	})
}

// DecryptUserdataFromToken resolves the human user behind a session token.
// It is only meaningful for session tokens - a device token has no OS user
// behind it, only the scopes it was granted at creation - so callers that
// also accept device tokens must check IsDeviceToken first and skip this for
// those requests rather than calling it and treating a resulting error as
// "unauthenticated".
func DecryptUserdataFromToken(r *http.Request, conf *config.SmartPiConfig) (models.User, error) {

	bearer, ok := bearerToken(r)
	if !ok {
		var errorObject models.Error
		errorObject.Message = "Invalid Token"
		return models.User{}, errorObject
	}

	token, err := parseSessionToken(bearer, conf)
	if err != nil {
		return models.User{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user := models.User{
			Name: claims["username"].(string),
		}
		return user, nil
	}

	log.Printf("Invalid JWT Token")
	return models.User{}, err
}

// IsDeviceToken reports whether the request's bearer token is a device token
// (see package devicetoken) rather than a session token. Handlers that gate
// on a per-user allowlist alongside TokenVerifyMiddleWare - AllowedDigitalOutUser
// and AllowedAnalogOut420mAUser - use this to skip that allowlist for device
// tokens: a device token's access is already scoped per-token at creation
// time in the web UI, by an operator who is themselves already subject to
// that allowlist, so re-checking an OS username that a device token never
// had in the first place would only lock every device token out.
func IsDeviceToken(r *http.Request) bool {
	bearer, ok := bearerToken(r)
	return ok && devicetoken.LooksLikeToken(bearer)
}

// TokenVerifyMiddleWare gates next behind a bearer token that carries scope.
//
// A session token (a JWT signed with conf.AppKey) satisfies every scope, the
// same unconditional access it has always had - session tokens represent a
// logged-in human, and scoping human access is a separate, independent
// concern from this change. A device token only passes if it exists in
// tokens and was explicitly granted scope when it was created; everything
// else about it, including its validity, is decided by that lookup alone -
// it is neither signed nor time-limited.
func TokenVerifyMiddleWare(next http.HandlerFunc, conf *config.SmartPiConfig, tokens *devicetoken.Store, scope string) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		bearer, ok := bearerToken(r)
		if !ok {
			errorObject.Message = "Invalid token."
			RespondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}

		if devicetoken.LooksLikeToken(bearer) {
			tok, found := tokens.Lookup(bearer)
			if !found || !tok.HasScope(scope) {
				errorObject.Message = "Invalid token."
				RespondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}
			tokens.TouchLastUsed(tok.ID)
			next.ServeHTTP(w, r)
			return
		}

		token, err := parseSessionToken(bearer, conf)
		if err != nil {
			errorObject.Message = err.Error()
			RespondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}
		if !token.Valid {
			errorObject.Message = "Invalid token."
			RespondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}
		next.ServeHTTP(w, r)
	})

}

// RequireSessionToken gates next behind a session token specifically - a
// device token is never accepted here, regardless of what scopes it carries.
// This is deliberately stricter than TokenVerifyMiddleWare and used only for
// the token-management endpoints themselves: if a device token could manage
// tokens, a leaked digitalout-scoped token could mint itself a
// config:write-scoped replacement, and per-token scoping would stop meaning
// anything.
func RequireSessionToken(next http.HandlerFunc, conf *config.SmartPiConfig) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error

		bearer, ok := bearerToken(r)
		if !ok || devicetoken.LooksLikeToken(bearer) {
			errorObject.Message = "Invalid token."
			RespondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}

		token, err := parseSessionToken(bearer, conf)
		if err != nil {
			errorObject.Message = err.Error()
			RespondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}
		if !token.Valid {
			errorObject.Message = "Invalid token."
			RespondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}
		next.ServeHTTP(w, r)
	})

}

func RespondWithError(w http.ResponseWriter, status int, error models.Error) {
	// EnableCors(&w)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(error)
}

func ResponseJSON(w http.ResponseWriter, data interface{}) {
	// EnableCors(&w)
	json.NewEncoder(w).Encode(data)
}

// func CheckConfigForPasswordMiddleWare(next http.HandlerFunc, c *config.SmartPiConfig) http.HandlerFunc {

// 	if c.SecureValues {
// 		return TokenVerifyMiddleWare(next)
// 	} else {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			next.ServeHTTP(w, r)
// 		})
// 	}

// }
