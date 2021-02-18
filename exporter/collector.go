package exporter

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/digineo/triax-eoc-exporter/triax"
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
	nodeStatus  = nodeDesc("status", "current endpoint status")
	nodeUptime  = nodeDesc("uptime", "uptime of endpoint in seconds")
	nodeLoad    = nodeDesc("load", "current system load of endpoint")
	nodeGhnPort = nodeDesc("ghn_port", "G.HN port number", "ghn_mac")
	nodeClients = nodeDesc("clients", "number of connected WLAN clients", "band")

	counterLabel   = []string{"interface", "direction"}
	counterBytes   = nodeDesc("interface_bytes", "total bytes transmitted or received", counterLabel...)
	counterPackets = nodeDesc("interface_packets", "total packets transmitted or received", counterLabel...)
	counterErrors  = nodeDesc("interface_errors", "total number of errors", counterLabel...)

	ghnRxbps = nodeDesc("ghn_rxbps", "negotiated RX rate in bps")
	ghnTxbps = nodeDesc("ghn_txbps", "negotiated TX rate in bps")
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

	counterMetric := func(counters *triax.Counters, node, ifname string) {
		metric(counterBytes, C, float64(counters.RxByte), node, ifname, "rx")
		metric(counterBytes, C, float64(counters.TxByte), node, ifname, "tx")
		metric(counterPackets, C, float64(counters.RxPacket), node, ifname, "rx")
		metric(counterPackets, C, float64(counters.TxPacket), node, ifname, "tx")
		metric(counterErrors, C, float64(counters.RxErr), node, ifname, "rx")
		metric(counterErrors, C, float64(counters.TxErr), node, ifname, "tx")
	}

	m, err := t.client.Metrics(t.ctx)
	metric(ctrlUp, G, boolToFloat(err == nil))
	if err != nil {
		log.Println("fetching failed:", err)
		return
	}

	metric(ctrlUptime, C, float64(m.Uptime))
	metric(ctrlLoad, G, m.Load)

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

			// ethernet statistics
			for _, stats := range node.Statistics.Ethernet {
				if stats.Link {
					counterMetric(&stats.Counters, node.Name, fmt.Sprintf("eth%d", stats.Port))
				}
			}

			// wireless statistics
			for _, stats := range node.Statistics.Wireless {
				metric(nodeClients, G, float64(stats.Clients), node.Name, strconv.Itoa(stats.Band))
				counterMetric(&stats.Counters, node.Name, fmt.Sprintf("wifi%d", stats.Band))
			}

			// ghn statistics
			if stats := node.GhnStats; stats != nil {
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
