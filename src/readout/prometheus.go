// Prometheus Metrics Exporter

package main

import (
	"github.com/nDenerserve/SmartPi/src/smartpi"
	"github.com/prometheus/client_golang/prometheus"
)

const metricsNamespace = "smartpi"

var (
	currentMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "amps",
			Help:      "Line current",
		},
		[]string{"phase"},
	)
	voltageMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "volts",
			Help:      "Line voltage",
		},
		[]string{"phase"},
	)
	activePowerMetirc = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "active_watts",
			Help:      "Active Watts",
		},
		[]string{"phase"},
	)
	cosphiMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "phase_angle",
			Help:      "Line voltage phase angle",
		},
		[]string{"phase"},
	)
	frequencyMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "frequency_hertz",
			Help:      "Line frequency in hertz",
		},
		[]string{"phase"},
	)
	apparentPowerMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "apparent_power_volt_amps",
			Help:      "Line apparent power in volt amps",
		},
		[]string{"phase"},
	)
	reactivePowerMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "reactive_power_volt_amps",
			Help:      "Line reactive power in volt amps reactive",
		},
		[]string{"phase"},
	)
	powerFactorMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "power_factor_ratio",
			Help:      "Line power factor ratio",
		},
		[]string{"phase"},
	)
)

func updatePrometheusMetrics(v *smartpi.ADE7878Readout) {
	currentMetric.WithLabelValues("A").Set(v.Current[smartpi.PhaseA])
	currentMetric.WithLabelValues("B").Set(v.Current[smartpi.PhaseB])
	currentMetric.WithLabelValues("C").Set(v.Current[smartpi.PhaseC])
	currentMetric.WithLabelValues("N").Set(v.Current[smartpi.PhaseN])
	voltageMetric.WithLabelValues("A").Set(v.Voltage[smartpi.PhaseA])
	voltageMetric.WithLabelValues("B").Set(v.Voltage[smartpi.PhaseB])
	voltageMetric.WithLabelValues("C").Set(v.Voltage[smartpi.PhaseC])
	activePowerMetirc.WithLabelValues("A").Set(v.ActiveWatts[smartpi.PhaseA])
	activePowerMetirc.WithLabelValues("B").Set(v.ActiveWatts[smartpi.PhaseB])
	activePowerMetirc.WithLabelValues("C").Set(v.ActiveWatts[smartpi.PhaseC])
	cosphiMetric.WithLabelValues("A").Set(v.CosPhi[smartpi.PhaseA])
	cosphiMetric.WithLabelValues("B").Set(v.CosPhi[smartpi.PhaseB])
	cosphiMetric.WithLabelValues("C").Set(v.CosPhi[smartpi.PhaseC])
	frequencyMetric.WithLabelValues("A").Set(v.Frequency[smartpi.PhaseA])
	frequencyMetric.WithLabelValues("B").Set(v.Frequency[smartpi.PhaseB])
	frequencyMetric.WithLabelValues("C").Set(v.Frequency[smartpi.PhaseC])
	apparentPowerMetric.WithLabelValues("A").Set(v.ApparentPower[smartpi.PhaseA])
	apparentPowerMetric.WithLabelValues("B").Set(v.ApparentPower[smartpi.PhaseB])
	apparentPowerMetric.WithLabelValues("C").Set(v.ApparentPower[smartpi.PhaseC])
	reactivePowerMetric.WithLabelValues("A").Set(v.ReactivePower[smartpi.PhaseA])
	reactivePowerMetric.WithLabelValues("B").Set(v.ReactivePower[smartpi.PhaseB])
	reactivePowerMetric.WithLabelValues("C").Set(v.ReactivePower[smartpi.PhaseC])
	powerFactorMetric.WithLabelValues("A").Set(v.PowerFactor[smartpi.PhaseA])
	powerFactorMetric.WithLabelValues("B").Set(v.PowerFactor[smartpi.PhaseB])
	powerFactorMetric.WithLabelValues("C").Set(v.PowerFactor[smartpi.PhaseC])
}
