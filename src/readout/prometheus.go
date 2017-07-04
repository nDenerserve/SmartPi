// Prometheus Metrics Exporter

package main

import (
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

func updatePrometheusMetrics(v [25]float64) {
	currentMetric.WithLabelValues("A").Set(v[0])
	currentMetric.WithLabelValues("B").Set(v[1])
	currentMetric.WithLabelValues("C").Set(v[2])
	currentMetric.WithLabelValues("N").Set(v[3])
	voltageMetric.WithLabelValues("A").Set(v[4])
	voltageMetric.WithLabelValues("B").Set(v[5])
	voltageMetric.WithLabelValues("C").Set(v[6])
	activePowerMetirc.WithLabelValues("A").Set(v[7])
	activePowerMetirc.WithLabelValues("B").Set(v[8])
	activePowerMetirc.WithLabelValues("C").Set(v[9])
	cosphiMetric.WithLabelValues("A").Set(v[10])
	cosphiMetric.WithLabelValues("B").Set(v[11])
	cosphiMetric.WithLabelValues("C").Set(v[12])
	frequencyMetric.WithLabelValues("A").Set(v[13])
	frequencyMetric.WithLabelValues("B").Set(v[14])
	frequencyMetric.WithLabelValues("C").Set(v[15])
	apparentPowerMetric.WithLabelValues("A").Set(v[16])
	apparentPowerMetric.WithLabelValues("B").Set(v[17])
	apparentPowerMetric.WithLabelValues("C").Set(v[18])
	reactivePowerMetric.WithLabelValues("A").Set(v[19])
	reactivePowerMetric.WithLabelValues("B").Set(v[20])
	reactivePowerMetric.WithLabelValues("C").Set(v[21])
	powerFactorMetric.WithLabelValues("A").Set(v[22])
	powerFactorMetric.WithLabelValues("B").Set(v[23])
	powerFactorMetric.WithLabelValues("C").Set(v[24])
}
