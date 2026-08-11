package models

type AnalogOut420mAStatus struct {
	Moduleaddress string  `json:"moduleaddress"`
	CurrentValue  float64 `json:"currentvalue"`
	SetValue      float64 `json:"setvalue"`
	Error         string  `json:"error,omitempty"`
}
