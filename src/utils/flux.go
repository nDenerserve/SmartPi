package utils

import (
	"fmt"
	"regexp"
)

// Values from a request that end up in an InfluxDB Flux query are checked
// here before the query is built: the data endpoints need no login, and a
// quote or bracket in a value would otherwise change the query itself.

// fluxDurationRe is a Flux duration literal such as "300s", "1h" or "1h30m".
var fluxDurationRe = regexp.MustCompile(`^([0-9]{1,9}(ns|us|µs|ms|s|m|h|d|w|mo|y))+$`)

// fluxFieldRe is a measurement field name such as "P1" or "CosPhi3".
var fluxFieldRe = regexp.MustCompile(`^[A-Za-z0-9_]{1,64}$`)

// fluxFuncs are the aggregate functions the API offers.
var fluxFuncs = map[string]bool{"mean": true, "median": true, "sum": true, "min": true, "max": true,
	"first": true, "last": true, "count": true, "spread": true}

// FluxDuration checks a duration for aggregateWindow(every: …).
func FluxDuration(s string) error {
	if !fluxDurationRe.MatchString(s) {
		return fmt.Errorf("invalid aggregate %q (e.g. 300s, 15m, 1h)", s)
	}
	return nil
}

// FluxField checks a field name used in a filter.
func FluxField(s string) error {
	if !fluxFieldRe.MatchString(s) {
		return fmt.Errorf("invalid value %q", s)
	}
	return nil
}

// FluxFunc checks the name of an aggregate function.
func FluxFunc(s string) error {
	if !fluxFuncs[s] {
		return fmt.Errorf("invalid function %q", s)
	}
	return nil
}
