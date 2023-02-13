package models

type User struct {
	Name     string   `json:"username"`
	Password string   `json:"password"`
	Role     []string `json:"role"`
	Exist    bool     `json:"exist"`
}
