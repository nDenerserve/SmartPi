package models

type Error struct {
	Message string `json:"message"`
}

// Error implements error
func (e Error) Error() string {
	return e.Message
}
