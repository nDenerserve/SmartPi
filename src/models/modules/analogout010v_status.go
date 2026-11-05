package models

// AnalogOut010VChannel is the state of one output of the MCP4728-based
// 0-10V analog output module, as read back from the DAC itself.
type AnalogOut010VChannel struct {
	// Channel is 1-4, the output as labelled on the board, left to right.
	Channel int `json:"channel"`
	// DACChannel is the MCP4728 channel (A-D) behind this output - the
	// board routes them swapped in pairs, see analogOut010VOutputToDAC.
	DACChannel string `json:"dacchannel"`
	// Voltage is the output voltage the DAC's live register currently
	// drives, converted to the 0-10V scale at the terminal.
	Voltage float64 `json:"voltage"`
	// DACValue is the raw 12-bit code in the live register.
	DACValue uint16 `json:"dacvalue"`
	// EEPROMVoltage/EEPROMDACValue are the values the DAC loads on
	// power-up - only changed by a set request with persist=true.
	EEPROMVoltage  float64 `json:"eepromvoltage"`
	EEPROMDACValue uint16  `json:"eepromdacvalue"`
	// Vref is "internal" (2.048V bandgap) or "vdd" (3.3V supply).
	Vref string `json:"vref"`
	// Gain is the DAC's internal output gain, 1 or 2.
	Gain int `json:"gain"`
	// PowerDown is the MCP4728 power-down mode, 0 = normal operation.
	PowerDown int `json:"powerdown"`
	// Calibrated reports whether a calibration from /etc/smartpiModules
	// ([analogout010v] calibration_<address>) is applied to this output -
	// Voltage/EEPROMVoltage then include it.
	Calibrated bool `json:"calibrated"`
}

// AnalogOut010VStatus is the state of one 4-channel 0-10V output module.
// Channels is indexed 0-3 for outputs 1-4 as labelled on the board.
type AnalogOut010VStatus struct {
	Moduleaddress string                  `json:"moduleaddress"`
	Channels      [4]AnalogOut010VChannel `json:"channels"`
	Error         string                  `json:"error,omitempty"`
}
