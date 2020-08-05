package smartpiapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/nDenerserve/SmartPi/src/smartpi"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginReturn struct {
	UserName string   `json:"username"`
	UserRole []string `json:"userrole"`
	Token    string   `json:"token"`
}

var User smartpi.User
var Config smartpi.Config

// Login is our handler to take a username and password and,
// if it's valid, return a token used for future requests.
func Login(w http.ResponseWriter, r *http.Request) {

	var ret loginReturn

	w.Header().Add("Content-Type", "application/json")
	// w.Header().Add("Access-Control-Allow-Origin", "*")
	r.ParseForm()

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println(err)
	}

	var cred credentials
	err = json.Unmarshal(b, &cred)
	if err != nil {
		log.Println(err)
	}

	if (cred == credentials{}) {
		cred.Username = r.Form.Get("username")
		cred.Password = r.Form.Get("password")
	}

	User.ReadUser(cred.Username, cred.Password)

	if !User.Exist {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":"invalid_credentials"}`)
		return
	}

	// We are happy with the credentials, so build a token. We've given it
	// an expiry of 1 hour.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user":     User.Name,
		"password": User.Password,
		"role":     User.Role,
		"exp":      time.Now().Add(time.Hour * time.Duration(5)).Unix(),
		"iat":      time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(Config.AppKey))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"error":"token_generation_failed"}`)
		return
	}

	ret = loginReturn{UserName: User.Name, UserRole: User.Role, Token: tokenString}

	if err := json.NewEncoder(w).Encode(ret); err != nil {
		log.Println(err)
	}

	// io.WriteString(w, `{"token":"`+tokenString+`"}`)
	return
}
