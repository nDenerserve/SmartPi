package models

type Error struct {
	Message string `json:"message"`
}

// Error implements error
func (Error) Error() string {
	panic("unimplemented")
}
