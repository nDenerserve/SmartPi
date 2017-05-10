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
			Name:      "phase_frequency_hertz",
			Help:      "Line frequency in hertz",
		},
		[]string{"phase"},
	)
)

func updatePrometheusMetrics(v [25]float32) {
	currentMetric.WithLabelValues("A").Set(float64(v[0]))
	currentMetric.WithLabelValues("B").Set(float64(v[1]))
	currentMetric.WithLabelValues("C").Set(float64(v[2]))
	currentMetric.WithLabelValues("N").Set(float64(v[3]))
	voltageMetric.WithLabelValues("A").Set(float64(v[4]))
	voltageMetric.WithLabelValues("B").Set(float64(v[5]))
	voltageMetric.WithLabelValues("C").Set(float64(v[6]))
	activePowerMetirc.WithLabelValues("A").Set(float64(v[7]))
	activePowerMetirc.WithLabelValues("B").Set(float64(v[8]))
	activePowerMetirc.WithLabelValues("C").Set(float64(v[9]))
	cosphiMetric.WithLabelValues("A").Set(float64(v[10]))
	cosphiMetric.WithLabelValues("B").Set(float64(v[11]))
	cosphiMetric.WithLabelValues("C").Set(float64(v[12]))
	frequencyMetric.WithLabelValues("A").Set(float64(v[13]))
	frequencyMetric.WithLabelValues("B").Set(float64(v[14]))
	frequencyMetric.WithLabelValues("C").Set(float64(v[15]))
}
