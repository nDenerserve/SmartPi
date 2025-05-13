package models

type ADE7878Readout struct {
	Current           Readings
	Voltage           Readings
	ActiveWatts       Readings
	CosPhi            Readings
	Frequency         Readings
	ApparentPower     Readings
	ReactivePower     Readings
	PowerFactor       Readings
	ActiveEnergy      Readings
	Energyconsumption Readings
	Energyproduction  Readings
}
