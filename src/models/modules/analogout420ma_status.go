package models

type AnalogOut420mAStatus struct {
	Moduleaddress string  `json:"moduleaddress"`
	CurrentValue  float64 `json:"currentvalue"`
	SetValue      float64 `json:"setvalue"`
	// DACValue is the raw 12-bit code in the DAC register.
	DACValue uint16 `json:"dacvalue"`
	// Calibrated tells whether a calibration stored on the SmartPi was
	// applied to CurrentValue/SetValue and to the code written.
	Calibrated bool   `json:"calibrated"`
	Error      string `json:"error,omitempty"`
}
