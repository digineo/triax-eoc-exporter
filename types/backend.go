package types

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

type Backend interface {
	Collect(context.Context, chan<- prometheus.Metric) error
}
