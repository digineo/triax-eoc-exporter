package v2

import (
	"context"
	"fmt"
	"strconv"

	"github.com/digineo/triax-eoc-exporter/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (b *backend) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	const C, G = prometheus.CounterValue, prometheus.GaugeValue

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

	board, err := b.Board(ctx)
	if err != nil {
		return err
	}
	metric(types.CtrlInfo, C, 1, board.Serial, board.EthMac, board.Release.Revision)

	m, err := b.Metrics(ctx)
	if err != nil {
		return err
	}

	metric(types.CtrlUptime, C, float64(m.Uptime))
	metric(types.CtrlLoad, G, m.Load)

	metric(types.CtrlMemoryTotal, G, float64(m.Memory.Total))
	metric(types.CtrlMemoryFree, G, float64(m.Memory.Free))
	metric(types.CtrlMemoryBuffered, G, float64(m.Memory.Buffered))
	metric(types.CtrlMemoryShared, G, float64(m.Memory.Shared))

	if ports := m.GhnPorts; ports != nil {
		for _, port := range ports {
			number := strconv.Itoa(port.Number)
			metric(types.CtrlGhnNumRegistered, G, float64(port.EndpointsRegistered), number)
			metric(types.CtrlGhnNumOnline, G, float64(port.EndpointsOnline), number)
		}
	}

	if nodes := m.Endpoints; nodes != nil {
		for _, node := range nodes {
			metric(types.NodeStatus, G, float64(node.Status), node.Name)

			if node.Uptime != nil {
				metric(types.NodeUptime, C, float64(*node.Uptime), node.Name)
			} else {
				metric(types.NodeOffline, C, float64(node.OfflineSince.Unix()), node.Name)
			}

			// ethernet statistics
			for _, stats := range node.Statistics.Ethernet {
				if stats.Link {
					counterMetric(&stats.Counters, node.Name, fmt.Sprintf("eth%d", stats.Port))
				}
			}

			// wireless statistics
			for _, stats := range node.Statistics.Wireless {
				metric(types.NodeClients, G, float64(stats.Clients), node.Name, strconv.Itoa(stats.Band))
				counterMetric(&stats.Counters, node.Name, fmt.Sprintf("wifi%d", stats.Band))
			}

			// ghn statistics
			if stats := node.GhnStats; stats != nil {
				metric(types.GhnRxbps, G, float64(stats.Rxbps), node.Name)
				metric(types.GhnTxbps, G, float64(stats.Txbps), node.Name)
			}
		}
	}

	return nil
}
