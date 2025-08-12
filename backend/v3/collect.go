package v3

import (
	"context"
	"fmt"
	"strconv"

	"github.com/digineo/triax-eoc-exporter/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (b *backend) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	const C, G = prometheus.CounterValue, prometheus.GaugeValue
	response := metricsResponse{}
	capabilities := capabilitiesResponse{}

	metric := func(desc *prometheus.Desc, typ prometheus.ValueType, v float64, label ...string) {
		ch <- prometheus.MustNewConstMetric(desc, typ, v, label...)
	}
	counterMetric := func(counters *Counters, node, ifname string) {
		metric(types.CounterBytes, C, float64(counters.RxByte), node, ifname, "rx")
		metric(types.CounterBytes, C, float64(counters.TxByte), node, ifname, "tx")
		metric(types.CounterPackets, C, float64(counters.RxPacket), node, ifname, "rx")
		metric(types.CounterPackets, C, float64(counters.TxPacket), node, ifname, "tx")
		metric(types.CounterErrors, C, float64(counters.RxErr), node, ifname, "rx")
		metric(types.CounterErrors, C, float64(counters.TxErr), node, ifname, "tx")
	}

	if err := b.Get(ctx, capabilitiesPath, &capabilities); err != nil {
		return err
	}

	if err := b.Get(ctx, matricsPath, &response); err != nil {
		return err
	}

	metric(types.CtrlInfo, C, 1, capabilities.Product.Serial, capabilities.Product.Mac, response.System.Version)
	metric(types.CtrlUptime, C, float64(response.System.Uptime))
	metric(types.CtrlMemoryTotal, G, float64(response.System.Memory.Total))
	metric(types.CtrlMemoryFree, G, float64(response.System.Memory.Total-response.System.Memory.Used))

	for _, modem := range response.Ghn.Modems {
		number := strconv.Itoa(modem.Index + 1)
		metric(types.CtrlGhnNumRegistered, G, float64(modem.EndpointRegistered), number)
		metric(types.CtrlGhnNumOnline, G, float64(modem.EndpointCount), number)
	}

	// mapping from MAC addresses to names
	macToName := make(map[string]string)

	// Endpoint side
	for mac, node := range response.Remote {
		name := node.System.Name

		// store name in mappings
		macToName[mac] = name

		metric(types.NodeInfo, G, 1, name, node.Serial, node.Mac, node.System.Model)
		metric(types.NodeStatus, G, float64(node.State), name)

		if uptime := node.System.Uptime; uptime != nil {
			metric(types.NodeUptime, G, float64(*uptime), name)
		}

		// ethernet statistics
		for _, stats := range node.Ethernet {
			if stats.Link {
				counterMetric(&stats.Counters, name, fmt.Sprintf("eth%d", stats.Port))
			}
		}

		// wireless statistics
		for _, stats := range node.Wireless {
			metric(types.NodeClients, G, float64(stats.Clients), name, strconv.Itoa(stats.Band))
			counterMetric(&stats.Counters, name, fmt.Sprintf("wifi%d", stats.Band))
		}

		// G.hn statistics
		if len(node.Ghn) > 0 && node.Ghn[0].Status != nil {
			ghn := node.Ghn[0]
			if ghn.Bitrate != nil {
				metric(types.GhnRxbps, G, float64(ghn.Bitrate.Rx), name)
				metric(types.GhnTxbps, G, float64(ghn.Bitrate.Tx), name)
			}
			if ghn.Snr != nil {
				metric(types.GhnSnrMin, G, float64(ghn.Snr.Min), name, types.SIDE_ENDPOINT)
				metric(types.GhnSnrAvg, G, float64(ghn.Snr.Avg), name, types.SIDE_ENDPOINT)
				metric(types.GhnSnrMax, G, float64(ghn.Snr.Max), name, types.SIDE_ENDPOINT)

			}
		}
	}

	// Controller Side
	for mac, node := range response.Ghn.Nodes {
		name := macToName[mac]
		if name == "" {
			name = mac
		}

		metric(types.GhnWireLength, G, float64(node.WireLength), name)
		metric(types.GhnSnrMin, G, float64(node.Snr.Min), name, types.SIDE_CONTROLLER)
		metric(types.GhnSnrAvg, G, float64(node.Snr.Avg), name, types.SIDE_CONTROLLER)
		metric(types.GhnSnrMax, G, float64(node.Snr.Max), name, types.SIDE_CONTROLLER)
	}

	return nil
}
