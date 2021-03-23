package triax

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (c *Client) Metrics(ctx context.Context) (*Metrics, error) {
	sys := sysinfoResponse{}      // uptime, memory
	eoc := syseocResponse{}       // EoC port names
	ghn := ghnStatusResponse{}    // G.HN port status
	nodes := nodeStatusResponse{} // data for each AP

	if err := c.Get(ctx, sysinfoPath, &eoc); err != nil {
		return nil, err
	}

	if err := c.Get(ctx, syseocPath, &eoc); err != nil {
		return nil, err
	}

	if err := c.Get(ctx, ghnStatusPath, &ghn); err != nil {
		return nil, err
	}

	if err := c.Get(ctx, nodeStatusPath, &nodes); err != nil {
		return nil, err
	}

	m := &Metrics{}
	m.Uptime = sys.Uptime
	m.Load = sys.Load
	m.Memory.Total = sys.Memory.Total
	m.Memory.Free = sys.Memory.Free
	m.Memory.Shared = sys.Memory.Shared
	m.Memory.Buffered = sys.Memory.Buffered

	m.GhnPorts = make(map[string]*GhnPort)
	for _, port := range ghn {
		m.GhnPorts[strings.ToLower(port.Mac)] = &GhnPort{
			Number:              -1, // determined in next step
			EndpointsOnline:     port.Connected,
			EndpointsRegistered: port.Registered,
		}
	}

	for mac := range m.GhnPorts {
		if i := eoc.MacAddr.Index(mac); i >= 0 {
			m.GhnPorts[mac].Number = i + 1 // yep.
		}
	}

	m.Endpoints = make([]EndpointMetrics, len(nodes))
	i := 0
	for _, node := range nodes {
		ep := &m.Endpoints[i]
		ep.Name = node.Name
		ep.MAC = node.Mac
		ep.Status = node.Statusid
		ep.StatusText = node.Status
		ep.Uptime = node.Sysinfo.Uptime
		ep.Load = node.Sysinfo.Load
		ep.GhnPortNumber = -1
		ep.GhnStats = node.GhnStats
		ep.Statistics = node.Statistics

		if node.RegTimestamp != "" {
			val, err := strconv.Atoi(node.RegTimestamp)
			if err != nil {
				return nil, fmt.Errorf("unable to parse regts value '%v': %w", node.RegTimestamp, err)
			} else {
				ep.OfflineSince = time.Unix(int64(val), 0)
			}
		}

		if mac := node.GhnMaster; mac != "" {
			ep.GhnPortMac = mac
			ep.GhnPortNumber = eoc.MacAddr.Index(mac)
		}

		i++
	}

	return m, nil
}
