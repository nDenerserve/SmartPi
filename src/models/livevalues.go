package models

import "time"

type Livedata struct {
	Time   time.Time   `json:"time"`
	Values []Livevalue `json:"values"`
}

type Livevalue struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}
