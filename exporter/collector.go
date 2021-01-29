package exporter

import (
	"context"
	"log"
	"strconv"

	"git.digineo.de/digineo/triax_eoc_exporter/triax"
	"github.com/prometheus/client_golang/prometheus"
)

type triaxCollector struct {
	client *triax.Client
	ctx    context.Context
}

var _ prometheus.Collector = (*triaxCollector)(nil)

var (
	ctrlUp               = ctrlDesc("up", "indicator whether controller is reachable")
	ctrlUptime           = ctrlDesc("uptime", "uptime of controller in seconds")
	ctrlLoad             = ctrlDesc("load", "current system load of controller")
	ctrlMemoryTotal      = ctrlDesc("mem_total", "total system memory of controller in bytes")
	ctrlMemoryFree       = ctrlDesc("mem_free", "free system memory of controller in bytes")
	ctrlMemoryBuffered   = ctrlDesc("mem_buffered", "buffered system memory of controller in bytes")
	ctrlMemoryShared     = ctrlDesc("mem_shared", "shared system memory of controller in bytes")
	ctrlGhnNumOnline     = ctrlDesc("ghn_endpoints_online", "number of endponts online for a G.HN port", "port")
	ctrlGhnNumRegistered = ctrlDesc("ghn_endpoints_registered", "number of endponts registered for a G.HN port", "port")

	nodeLabel   = []string{"name"}
	nodeStatus  = nodeDesc("status", "current AP status")
	nodeUptime  = nodeDesc("uptime", "uptime of AP in seconds")
	nodeLoad    = nodeDesc("load", "current system load of AP")
	nodeGhnPort = nodeDesc("ghn_port", "G.HN port number", "ghn_mac")
	nodeClients = nodeDesc("clients", "number of connected WLAN clients", "radio")

	ghnAbort  = nodeDesc("ghn_abort", "Total count of Abort")
	ghnError  = nodeDesc("ghn_error", "Total count of Error")
	ghnFrames = nodeDesc("ghn_frames", "Total count of Frames")
	ghnLpdus  = nodeDesc("ghn_lpdus", "Total count of Lpdus")
	ghnRxbps  = nodeDesc("ghn_rxbps", "Total count of Rxbps")
	ghnTxbps  = nodeDesc("ghn_txbps", "Total count of Txbps")

	// TODO: apGhnStats - welche Felder und was ist deren Bedeutung?
)

func (t *triaxCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- ctrlUp
	ch <- ctrlUptime
	ch <- ctrlLoad
	ch <- ctrlMemoryTotal
	ch <- ctrlMemoryFree
	ch <- ctrlMemoryBuffered
	ch <- ctrlMemoryShared
	ch <- ctrlGhnNumOnline
	ch <- ctrlGhnNumRegistered

	ch <- nodeStatus
	ch <- nodeUptime
	ch <- nodeLoad
	ch <- nodeGhnPort
	ch <- nodeClients
}

func (t *triaxCollector) Collect(ch chan<- prometheus.Metric) {
	const C, G = prometheus.CounterValue, prometheus.GaugeValue

	metric := func(desc *prometheus.Desc, typ prometheus.ValueType, v float64, label ...string) {
		ch <- prometheus.MustNewConstMetric(desc, typ, v, label...)
	}

	m, err := t.client.Metrics(t.ctx)
	metric(ctrlUp, G, boolToFloat(err == nil))
	if err != nil {
		log.Println("fetching failed:", err)
		return
	}

	metric(ctrlUptime, C, float64(m.Uptime))
	metric(ctrlLoad, G, float64(m.Load))

	metric(ctrlMemoryTotal, G, float64(m.Memory.Total))
	metric(ctrlMemoryFree, G, float64(m.Memory.Free))
	metric(ctrlMemoryBuffered, G, float64(m.Memory.Buffered))
	metric(ctrlMemoryShared, G, float64(m.Memory.Shared))

	if ports := m.GhnPorts; ports != nil {
		for _, port := range ports {
			number := strconv.Itoa(port.Number)
			metric(ctrlGhnNumRegistered, G, float64(port.EndpointsRegistered), number)
			metric(ctrlGhnNumOnline, G, float64(port.EndpointsOnline), number)
		}
	}

	if nodes := m.Endpoints; nodes != nil {
		for _, node := range nodes {
			metric(nodeStatus, G, float64(node.Status), node.Name)
			metric(nodeUptime, C, float64(node.Uptime), node.Name)
			metric(nodeClients, G, float64(node.Clients.Radio24), node.Name, "2.4 Ghz")
			metric(nodeClients, G, float64(node.Clients.Radio5), node.Name, "5 Ghz")

			if stats := node.GhnStats; stats != nil {
				// metric(ghnAbort, C, float64(stats.Abort), node.Name)
				// metric(ghnError, C, float64(stats.Error), node.Name)
				// metric(ghnFrames, C, float64(stats.Frames), node.Name)
				// metric(ghnLpdus, C, float64(stats.Lpdus), node.Name)
				metric(ghnRxbps, G, float64(stats.Rxbps), node.Name)
				metric(ghnTxbps, G, float64(stats.Txbps), node.Name)
			}
		}
	}
}

func boolToFloat(val bool) float64 {
	if val {
		return 1
	}

	return 0
}

func ctrlDesc(name, help string, extraLabel ...string) *prometheus.Desc {
	fqdn := prometheus.BuildFQName("triax", "eoc_controller", name)
	return prometheus.NewDesc(fqdn, help, extraLabel, nil)
}

func nodeDesc(name, help string, extraLabel ...string) *prometheus.Desc {
	fqdn := prometheus.BuildFQName("triax", "eoc_endpoint", name)
	return prometheus.NewDesc(fqdn, help, append(nodeLabel, extraLabel...), nil)
}
