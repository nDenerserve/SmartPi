package models

// I2CDeviceEntry is one address found occupied by an i2c scan.
type I2CDeviceEntry struct {
	// Address is the 7-bit bus address, formatted as "0xNN".
	Address string `json:"address"`
	// Status is "detected" for an address that acknowledged the probe, or
	// "in_use" for one already claimed by a kernel driver (i2cdetect's "UU")
	// and therefore not probed at all.
	Status string `json:"status"`
	// Hint is a best-effort guess at which SmartPi-supported chip sits at
	// Address, based purely on its address range - never confirmed by
	// reading any register, since none of MCP23017/ADE7878/MCP4725/MCP3424
	// has a WHO_AM_I register to check against. Empty if the range isn't
	// recognized.
	Hint string `json:"hint,omitempty"`
}

// I2CScanStatus is the result of scanning one I2C bus for occupied addresses.
type I2CScanStatus struct {
	Bus     string           `json:"bus"`
	Devices []I2CDeviceEntry `json:"devices"`
}
