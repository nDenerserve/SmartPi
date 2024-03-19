package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/nDenerserve/SmartPi/models"

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

func GenerateToken(user models.User) (string, error) {

	var err error
	secret := "secret"

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

func DecryptUserdataFromToken(r *http.Request) (models.User, error) {

	authHeader := r.Header.Get("Authorization")
	bearerToken := strings.Split(authHeader, " ")

	if len(bearerToken) == 2 {
		authToken := bearerToken[1]

		token, error := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(("There was an error"))
			}

			return []byte("secret"), nil
		})

		if error != nil {
			return models.User{}, error
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

			user := models.User{
				Name: claims["username"].(string),
				Role: strings.Split(strings.Replace(strings.Replace(fmt.Sprint(claims["role"]), "[", "", -1), "]", "", -1), " "), // remove the [ and ] and split into []string
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

func TokenVerifyMiddleWare(next http.HandlerFunc) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errorObject models.Error
		authHeader := r.Header.Get("Authorization")
		bearerToken := strings.Split(authHeader, " ")

		if len(bearerToken) == 2 {
			authToken := bearerToken[1]

			token, error := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf(("There was an error"))
				}

				return []byte("secret"), nil
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

func CheckAllowedUser(userRole []string, configRole []string) bool {

	userAllowed := false

	for _, uv := range userRole {
		for _, cv := range configRole {
			if uv == cv {
				userAllowed = true
			}
		}
	}

	return userAllowed
}

func Checklog(e error) {
	if e != nil {
		log.Println(e)
	}
}

func Checkpanic(e error) {
	if e != nil {
		panic(e)
	}
}

type bitString string

func (b bitString) AsByteSlice() []byte {
	var out []byte
	var str string

	for i := len(b); i > 0; i -= 8 {
		if i-8 < 0 {
			str = string(b[0:i])
		} else {
			str = string(b[i-8 : i])
		}
		v, err := strconv.ParseUint(str, 2, 8)
		if err != nil {
			panic(err)
		}
		out = append([]byte{byte(v)}, out...)
	}
	return out
}

func (b bitString) AsHexSlice() []string {
	var out []string
	byteSlice := b.AsByteSlice()
	for _, b := range byteSlice {
		out = append(out, "0x"+hex.EncodeToString([]byte{b}))
	}
	return out
}

// function, which takes a string as
// argument and return the reverse of string.
func Reverse(s string) string {
	rns := []rune(s) // convert to rune
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {

		// swap the letters of the string,
		// like first with last and so on.
		rns[i], rns[j] = rns[j], rns[i]
	}

	// return the reversed string.
	return string(rns)
}

func DiffTime(a, b time.Time) (int, int, int, int, int, int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year := int(y2 - y1)
	month := int(M2 - M1)
	day := int(d2 - d1)
	hour := int(h2 - h1)
	min := int(m2 - m1)
	sec := int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return year, month, day, hour, min, sec
}

func Monthchange(a, b time.Time) int {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, _ := a.Date()
	y2, M2, _ := b.Date()

	year := int(y2 - y1)
	month := int(M2 - M1)

	if month < 0 {
		month += (12 * year)
	}

	return month
}

func Int2StringSlice(intSlice []int) []string {

	stringSlice := []string{}

	for i := range intSlice {
		number := intSlice[i]
		text := strconv.Itoa(number)
		stringSlice = append(stringSlice, text)
	}
	return stringSlice
}
