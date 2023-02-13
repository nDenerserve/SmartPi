package models

type DigitalOutStatus struct {
	Moduleaddress string    `json:"moduleaddress"`
	PortStatus    [4]string `json:"portstatus"`
}
