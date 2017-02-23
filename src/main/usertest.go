package main


import (
	  "smartpi"
		"fmt"
)

func main() {

user := smartpi.NewUser()
user.ReadUserFromFile("root")
fmt.Println(user.Password)
fmt.Println(user.Exist)
for _, b := range user.Role {
	fmt.Println(b)
}
user.ReadUserFromFile("pi")
fmt.Println(user.Password)
fmt.Println(user.Exist)
for _, b := range user.Role {
	fmt.Println(b)
}
}
