package service

import (
	"context"
	"net/http"
)

type HealthAggregation struct {
	Health     HealthResult                 `yaml:"health"`
	Components []HealthAggregationComponent `yaml:"components"`
}

type HealthAggregationComponent struct {
	Name   string       `yaml:"name"`
	Health HealthResult `yaml:"health"`
}

type HealthResult string

func (r HealthResult) ToHTTPStatusCode() int {
	var status int

	switch r {
	case HealthUp:
		status = http.StatusOK
	case HealthDown:
		status = http.StatusServiceUnavailable
	}

	return status
}

const (
	HealthUp   HealthResult = "Up"
	HealthDown HealthResult = "Down"
)

//go:generate mockgen -source health.go -package mock -destination mock/health.go Health
type Health interface {
	Check(ctx context.Context) HealthAggregation
}

type HealthCheck func(ctx context.Context) (string, bool)

type healthAggregator struct {
	healthChecks []HealthCheck
}

func (h *healthAggregator) Check(ctx context.Context) HealthAggregation {
	finalOK := true
	checks := make(chan HealthAggregationComponent, len(h.healthChecks))
	components := make([]HealthAggregationComponent, len(h.healthChecks))

	for i := range h.healthChecks {
		i := i

		healthCheck := func() {
			name, healthOk := h.healthChecks[i](ctx)
			checks <- HealthAggregationComponent{
				Name:   name,
				Health: HealthResultFromBool(healthOk),
			}

			if !healthOk && finalOK {
				finalOK = false
			}
		}

		go healthCheck()
	}

	for i := range h.healthChecks {
		components[i] = <-checks
	}

	return HealthAggregation{Health: HealthResultFromBool(finalOK), Components: components}
}

func HealthAggregator(healthChecks []HealthCheck) Health {
	return &healthAggregator{healthChecks}
}

func HealthResultFromBool(healthUp bool) HealthResult {
	if healthUp {
		return HealthUp
	}

	return HealthDown
}
