// Prometheus Metrics Exporter

package main

import (
	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/smartpi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const metricsNamespace = "smartpi"

var (
	currentMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "amps",
			Help:      "Line current",
		},
		[]string{"phase"},
	)
	voltageMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "volts",
			Help:      "Line voltage",
		},
		[]string{"phase"},
	)
	activePowerMetirc = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "active_watts",
			Help:      "Active Watts",
		},
		[]string{"phase"},
	)
	cosphiMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "phase_angle",
			Help:      "Line voltage phase angle",
		},
		[]string{"phase"},
	)
	frequencyMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "frequency_hertz",
			Help:      "Line frequency in hertz",
		},
		[]string{"phase"},
	)
	apparentPowerMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "apparent_power_volt_amps",
			Help:      "Line apparent power in volt amps",
		},
		[]string{"phase"},
	)
	reactivePowerMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "reactive_power_volt_amps",
			Help:      "Line reactive power in volt amps reactive",
		},
		[]string{"phase"},
	)
	powerFactorMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "power_factor_ratio",
			Help:      "Line power factor ratio",
		},
		[]string{"phase"},
	)
	wattHoursConsumedMetric = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "consumed_watt_hours_total",
			Help:      "Accumulated watt hours consumed",
		},
		[]string{"phase"},
	)
	wattHoursProducedMetric = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "produced_watt_hours_total",
			Help:      "Accumulated watt hours produced",
		},
		[]string{"phase"},
	)
)

func updatePrometheusMetrics(v *smartpi.ADE7878Readout, c *config.Config) {
	currentMetric.WithLabelValues("N").Set(v.Current[models.PhaseN])
	for _, p := range smartpi.MainPhases {
		// Skip updating metrics where the phase is not measured.
		if !c.MeasureCurrent[p] && !c.MeasureVoltage[p] {
			continue
		}

		label := p.String()
		currentMetric.WithLabelValues(label).Set(v.Current[p])
		voltageMetric.WithLabelValues(label).Set(v.Voltage[p])
		activePowerMetirc.WithLabelValues(label).Set(v.ActiveWatts[p])
		cosphiMetric.WithLabelValues(label).Set(v.CosPhi[p])
		frequencyMetric.WithLabelValues(label).Set(v.Frequency[p])
		apparentPowerMetric.WithLabelValues(label).Set(v.ApparentPower[p])
		reactivePowerMetric.WithLabelValues(label).Set(v.ReactivePower[p])
		powerFactorMetric.WithLabelValues(label).Set(v.PowerFactor[p])

		// Use the active watts and sample rate to calculate an estimate of watt hours.
		wattHours := v.ActiveWatts[p] / (3600.0 * float64(c.Samplerate))
		if wattHours >= 0 {
			wattHoursConsumedMetric.WithLabelValues(label).Add(wattHours)
		} else {
			wattHoursProducedMetric.WithLabelValues(label).Add(-wattHours)
		}
	}
}
