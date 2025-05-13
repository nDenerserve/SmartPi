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

	"github.com/dgrijalva/jwt-go"
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

func DecryptUserdataFromToken(r *http.Request, conf *config.SmartPiConfig) (models.User, error) {

	authHeader := r.Header.Get("Authorization")
	bearerToken := strings.Split(authHeader, " ")

	if len(bearerToken) == 2 {
		authToken := bearerToken[1]

		token, error := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error")
			}

			return []byte(conf.AppKey), nil
		})

		if error != nil {
			return models.User{}, error
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			user := models.User{
				Name: claims["username"].(string),
			}
			return user, nil

		} else {
			log.Printf("Invalid JWT Token")
			return models.User{}, error
		}

	} else {
		var errorObject models.Error
		errorObject.Message = "Invalid Token"
		return models.User{}, errorObject
	}

}

func TokenVerifyMiddleWare(next http.HandlerFunc, conf *config.SmartPiConfig) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error
		authHeader := r.Header.Get("Authorization")
		bearerToken := strings.Split(authHeader, " ")

		log.Debug(bearerToken)

		if len(bearerToken) == 2 {
			authToken := bearerToken[1]

			token, error := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf(("There was an error"))
				}

				return []byte(conf.AppKey), nil
			})

			if error != nil {
				errorObject.Message = error.Error()
				RespondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}

			if token.Valid {
				next.ServeHTTP(w, r)
			} else {
				errorObject.Message = error.Error()
				RespondWithError(w, http.StatusUnauthorized, errorObject)
				return
			}
		} else {
			errorObject.Message = "Invalid token."
			RespondWithError(w, http.StatusUnauthorized, errorObject)
			return
		}

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
