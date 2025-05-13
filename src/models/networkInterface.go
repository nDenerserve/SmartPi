package models

type NetworkInterface struct {
	Name   string `json:"name"`
	Device string `json:"device"`
	Type   string `json:"type"`
	State  string `json:"state"`
}
