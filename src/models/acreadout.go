package models

type Readings map[SmartPiPhase]float64

type ReadoutAccumulator struct {
	Current           Readings
	Voltage           Readings
	ActiveWatts       Readings
	CosPhi            Readings
	Frequency         Readings
	WattHoursConsumed Readings
	WattHoursProduced Readings
}
