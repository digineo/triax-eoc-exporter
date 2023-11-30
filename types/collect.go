package types

import "github.com/prometheus/client_golang/prometheus"

var (
	CtrlUp               = CtrlDesc("up", "indicator whether controller is reachable")
	CtrlUptime           = CtrlDesc("uptime", "uptime of controller in seconds")
	CtrlInfo             = CtrlDesc("info", "controller infos about the installed software", "serial", "eth_mac", "version")
	CtrlLoad             = CtrlDesc("load", "current system load of controller")
	CtrlMemoryTotal      = CtrlDesc("mem_total", "total system memory of controller in bytes")
	CtrlMemoryFree       = CtrlDesc("mem_free", "free system memory of controller in bytes")
	CtrlMemoryBuffered   = CtrlDesc("mem_buffered", "buffered system memory of controller in bytes")
	CtrlMemoryShared     = CtrlDesc("mem_shared", "shared system memory of controller in bytes")
	CtrlGhnNumOnline     = CtrlDesc("ghn_endpoints_online", "number of endponts online for a G.HN port", "port")
	CtrlGhnNumRegistered = CtrlDesc("ghn_endpoints_registered", "number of endponts registered for a G.HN port", "port")

	NodeLabel   = []string{"name"}
	NodeStatus  = NodeDesc("status", "current endpoint status")
	NodeUptime  = NodeDesc("uptime", "uptime of endpoint in seconds")
	NodeOffline = NodeDesc("offline_since", "offline since unix timestamp")
	NodeLoad    = NodeDesc("load", "current system load of endpoint")
	NodeGhnPort = NodeDesc("ghn_port", "G.HN port number", "ghn_mac")
	NodeClients = NodeDesc("clients", "number of connected WLAN clients", "band")

	LounterLabel   = []string{"interface", "direction"}
	CounterBytes   = NodeDesc("interface_bytes", "total bytes transmitted or received", LounterLabel...)
	CounterPackets = NodeDesc("interface_packets", "total packets transmitted or received", LounterLabel...)
	CounterErrors  = NodeDesc("interface_errors", "total number of errors", LounterLabel...)

	GhnRxbps = NodeDesc("ghn_rxbps", "negotiated RX rate in bps")
	GhnTxbps = NodeDesc("ghn_txbps", "negotiated TX rate in bps")
)

func CtrlDesc(name, help string, extraLabel ...string) *prometheus.Desc {
	fqdn := prometheus.BuildFQName("triax", "eoc_controller", name)
	return prometheus.NewDesc(fqdn, help, extraLabel, nil)
}

func NodeDesc(name, help string, extraLabel ...string) *prometheus.Desc {
	fqdn := prometheus.BuildFQName("triax", "eoc_endpoint", name)
	return prometheus.NewDesc(fqdn, help, append(NodeLabel, extraLabel...), nil)
}
