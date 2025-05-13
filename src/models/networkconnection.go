package models

// NetworkConnection represents a network connection with relevant details
type NetworkConnection struct {
	Name                string
	Uuid                string
	Type                string
	Device              string
	ConnectionAddresses []NetworkConnectionAddress
	State               string
}

type NetworkConnectionAddress struct {
	Ipv4Address string
	IpMethod    string
	CidrSuffix  uint8
}
