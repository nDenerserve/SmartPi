package models

// WifiNetwork represents a WiFi network with relevant details
type WifiNetwork struct {
	SSID     string
	Mode     string
	Channel  string
	Rate     string
	Signal   string
	Bars     string
	Security string
}
