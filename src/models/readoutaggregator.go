package models

import "math"

// ReadoutAggregator condenses all samples of one publication window into a
// single ADE7878Readout. Instantaneous quantities (current, voltage, power,
// frequency) are averaged over the samples actually collected, energy
// quantities are summed up so that no energy gets lost when readouts are
// published less often than they are measured.
//
// The sample count is tracked instead of being derived from the configured
// samplerate, so the aggregation stays correct if samples are dropped or the
// samplerate is changed at runtime.
type ReadoutAggregator struct {
	// samples counts the readouts collected in the current window. Everything
	// below holds running sums over exactly those samples.
	samples int

	current       Readings
	voltage       Readings
	activeWatts   Readings
	cosPhi        Readings
	frequency     Readings
	apparentPower Readings
	reactivePower Readings
	powerFactor   Readings

	activeEnergy      Readings
	energyconsumption Readings
	energyproduction  Readings

	// wattHourBalanced is the signed energy balance over all phases, summed up
	// over the window. A negative value means the window produced more than it
	// consumed.
	wattHourBalanced float64
}

// NewReadoutAggregator returns an aggregator with an empty window.
func NewReadoutAggregator() *ReadoutAggregator {
	a := new(ReadoutAggregator)
	a.Reset()
	return a
}

// Reset starts a new publication window.
func (a *ReadoutAggregator) Reset() {
	a.samples = 0
	a.current = make(Readings)
	a.voltage = make(Readings)
	a.activeWatts = make(Readings)
	a.cosPhi = make(Readings)
	a.frequency = make(Readings)
	a.apparentPower = make(Readings)
	a.reactivePower = make(Readings)
	a.powerFactor = make(Readings)
	a.activeEnergy = make(Readings)
	a.energyconsumption = make(Readings)
	a.energyproduction = make(Readings)
	a.wattHourBalanced = 0.0
}

// Samples returns the number of readouts collected since the last reset.
func (a *ReadoutAggregator) Samples() int {
	return a.samples
}

// Add feeds one sample together with its balanced energy into the aggregator.
// Only the phases actually present in the readout are touched, so a phase that
// is not measured never contributes to the average of another one.
func (a *ReadoutAggregator) Add(v *ADE7878Readout, wattHourBalanced float64) {
	a.samples++
	addReadings(a.current, v.Current)
	addReadings(a.voltage, v.Voltage)
	addReadings(a.activeWatts, v.ActiveWatts)
	addReadings(a.cosPhi, v.CosPhi)
	addReadings(a.frequency, v.Frequency)
	addReadings(a.apparentPower, v.ApparentPower)
	addReadings(a.reactivePower, v.ReactivePower)
	addReadings(a.powerFactor, v.PowerFactor)
	addReadings(a.activeEnergy, v.ActiveEnergy)
	addReadings(a.energyconsumption, v.Energyconsumption)
	addReadings(a.energyproduction, v.Energyproduction)
	a.wattHourBalanced += wattHourBalanced
}

// Snapshot returns the aggregated readout of the current window together with
// the summed up balanced energy. It does not reset the aggregator.
func (a *ReadoutAggregator) Snapshot() (ADE7878Readout, float64) {
	r := ADE7878Readout{
		Current:           make(Readings),
		Voltage:           make(Readings),
		ActiveWatts:       make(Readings),
		CosPhi:            make(Readings),
		Frequency:         make(Readings),
		ApparentPower:     make(Readings),
		ReactivePower:     make(Readings),
		PowerFactor:       make(Readings),
		ActiveEnergy:      make(Readings),
		Energyconsumption: make(Readings),
		Energyproduction:  make(Readings),
	}
	// An empty window has no meaningful values and must not divide by zero.
	if a.samples == 0 {
		return r, 0.0
	}
	samples := float64(a.samples)

	meanReadings(r.Current, a.current, samples)
	meanReadings(r.Voltage, a.voltage, samples)
	meanReadings(r.ActiveWatts, a.activeWatts, samples)
	meanReadings(r.Frequency, a.frequency, samples)
	meanReadings(r.ApparentPower, a.apparentPower, samples)
	meanReadings(r.ReactivePower, a.reactivePower, samples)
	meanReadings(r.PowerFactor, a.powerFactor, samples)

	// Energy is summed up, not averaged.
	copyReadings(r.ActiveEnergy, a.activeEnergy)
	copyReadings(r.Energyconsumption, a.energyconsumption)
	copyReadings(r.Energyproduction, a.energyproduction)

	// cos phi is a displacement factor and must not be averaged arithmetically:
	// when the power direction changes within the window the individual values
	// cancel each other out. Averaging the power phasor instead keeps both the
	// magnitude and the sign of the aggregated value meaningful. Without any
	// power in the window the phasor is undefined, so the plain mean is kept.
	for p := range a.cosPhi {
		activeWatts := a.activeWatts[p] / samples
		reactivePower := a.reactivePower[p] / samples
		if apparent := math.Hypot(activeWatts, reactivePower); apparent > 0 {
			r.CosPhi[p] = activeWatts / apparent
		} else {
			r.CosPhi[p] = a.cosPhi[p] / samples
		}
	}

	return r, a.wattHourBalanced
}

// addReadings adds every reading of src to the matching entry in dst.
func addReadings(dst Readings, src Readings) {
	for p, value := range src {
		dst[p] += value
	}
}

// meanReadings writes the mean of every summed reading in src to dst.
func meanReadings(dst Readings, src Readings, samples float64) {
	for p, value := range src {
		dst[p] = value / samples
	}
}

// copyReadings transfers every reading of src to dst unchanged. It is used for
// the quantities that are summed up rather than averaged.
func copyReadings(dst Readings, src Readings) {
	for p, value := range src {
		dst[p] = value
	}
}
