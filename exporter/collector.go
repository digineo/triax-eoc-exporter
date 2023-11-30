package exporter

import (
	"context"
	"log/slog"

	"github.com/digineo/triax-eoc-exporter/client"
	"github.com/digineo/triax-eoc-exporter/types"
	"github.com/prometheus/client_golang/prometheus"
)

type triaxCollector struct {
	client *client.Client
	ctx    context.Context
}

var _ prometheus.Collector = (*triaxCollector)(nil)

func (t *triaxCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- types.CtrlUp
	ch <- types.CtrlUptime
	ch <- types.CtrlInfo
	ch <- types.CtrlLoad
	ch <- types.CtrlMemoryTotal
	ch <- types.CtrlMemoryFree
	ch <- types.CtrlMemoryBuffered
	ch <- types.CtrlMemoryShared
	ch <- types.CtrlGhnNumOnline
	ch <- types.CtrlGhnNumRegistered

	ch <- types.NodeStatus
	ch <- types.NodeUptime
	ch <- types.NodeLoad
	ch <- types.NodeGhnPort
	ch <- types.NodeClients
}

func (t *triaxCollector) Collect(ch chan<- prometheus.Metric) {
	err := t.client.Collect(t.ctx, ch)

	// Write up
	ch <- prometheus.MustNewConstMetric(types.CtrlUp, prometheus.GaugeValue, boolToFloat(err == nil))

	if err != nil {
		slog.Error("fetching failed", "error", err)
	}
}

func boolToFloat(val bool) float64 {
	if val {
		return 1
	}

	return 0
}
