package exporter

import (
	"context"
	"log"

	"git.digineo.de/digineo/triax_eoc_exporter/triax"
	"github.com/prometheus/client_golang/prometheus"
)

type triaxCollector struct {
	client *triax.Client
	ctx    context.Context
}

var _ prometheus.Collector = (*triaxCollector)(nil)

var (
	ctrlLabel            = []string{"mac"}
	ctrlUp               = ctrlDesc("up", "indicator whether controller is reachable")
	ctrlUptime           = ctrlDesc("uptime", "uptime of controller in seconds")
	ctrlLoad             = ctrlDesc("load", "current system load of controller")
	ctrlMemoryTotal      = ctrlDesc("mem_total", "total system memory of controller in bytes")
	ctrlMemoryFree       = ctrlDesc("mem_free", "free system memory of controller in bytes")
	ctrlMemoryBuffered   = ctrlDesc("mem_buffered", "buffered system memory of controller in bytes")
	ctrlMemoryShared     = ctrlDesc("mem_shared", "shared system memory of controller in bytes")
	ctrlGhnNumOnline     = ctrlDesc("ghn_endpoints_online", "number of endponts online for a G.HN port", "ghn_port", "ghn_mac")
	ctrlGhnNumRegistered = ctrlDesc("ghn_endpoints_registered", "number of endponts registered for a G.HN port", "ghn_port", "ghn_mac")

	apLabel   = []string{"controller", "mac", "name"}
	apStatus  = apDesc("status", "current AP status", "status_text")
	apUptime  = apDesc("uptime", "uptime of AP in seconds")
	apLoad    = apDesc("load", "current system load of AP")
	apGhnPort = apDesc("ghn_port", "G.HN port number", "ghn_mac")
	apClients = apDesc("clients", "number of connected WLAN clients", "radio")

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

	ch <- apStatus
	ch <- apUptime
	ch <- apLoad
	ch <- apGhnPort
	ch <- apClients
}

func (t *triaxCollector) Collect(ch chan<- prometheus.Metric) {
	const C, G = prometheus.CounterValue, prometheus.GaugeValue
	metric := func(desc *prometheus.Desc, typ prometheus.ValueType, v float64, label ...string) {
		ch <- prometheus.MustNewConstMetric(desc, typ, v, label...)
	}

	log.Printf("[%s] logging into controller", t.client.MAC)
	if err := t.client.Login(t.ctx); err != nil {
		log.Printf("[%s] error logging into controller: %v", t.client.MAC, err)
		metric(ctrlUp, G, 0, t.client.MAC)
		return
	}

	log.Printf("[%s] fetching metrics", t.client.MAC)
	m, err := t.client.FetchData(t.ctx)
	log.Printf("%v\n%#v", err, m)
}

func ctrlDesc(name, help string, extraLabel ...string) *prometheus.Desc {
	fqdn := prometheus.BuildFQName("triax", "eoc_controller", name)
	return prometheus.NewDesc(fqdn, help, append(ctrlLabel, extraLabel...), nil)
}

func apDesc(name, help string, extraLabel ...string) *prometheus.Desc {
	fqdn := prometheus.BuildFQName("triax", "eoc_endpoint", name)
	return prometheus.NewDesc(fqdn, help, append(apLabel, extraLabel...), nil)
}
