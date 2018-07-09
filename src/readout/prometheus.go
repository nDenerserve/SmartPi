// Prometheus Metrics Exporter

package main

import (
	"github.com/Nitroman605/SmartPi/src/smartpi"

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
	currentMetric.WithLabelValues("N").Set(v.Current[smartpi.PhaseN])
	for _, p := range smartpi.MainPhases {
		label := p.String()
		currentMetric.WithLabelValues(label).Set(v.Current[p])
		voltageMetric.WithLabelValues(label).Set(v.Voltage[p])
		activePowerMetirc.WithLabelValues(label).Set(v.ActiveWatts[p])
		cosphiMetric.WithLabelValues(label).Set(v.CosPhi[p])
		frequencyMetric.WithLabelValues(label).Set(v.Frequency[p])
		apparentPowerMetric.WithLabelValues(label).Set(v.ApparentPower[p])
		reactivePowerMetric.WithLabelValues(label).Set(v.ReactivePower[p])
		powerFactorMetric.WithLabelValues(label).Set(v.PowerFactor[p])
	}
}
