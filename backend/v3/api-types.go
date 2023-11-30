package v3

import "time"

const loginPath = "cgi.lua/login"

// request for /cgi.lua/login.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// response from /cgi.lua/login.
type loginResponse struct {
	Level     int    `json:"level"`
	Status    bool   `json:"status"`
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

const capabilitiesPath = "cgi.lua/capabilities"

type capabilitiesResponse struct {
	Product struct {
		BoardNo   int    `json:"board_no"`
		Type      string `json:"type"`
		BoardRev  string `json:"board_rev"`
		Macs      map[string]string
		BoardWeek int    `json:"board_week"`
		Name      string `json:"name"`
		Model     string `json:"model"`
		Mac       string `json:"mac"`
		BoardID   string `json:"board_id"`
		Serial    string `json:"serial"`
		BoardYear int    `json:"board_year"`
	} `json:"product"`
	Features struct {
		GhnMux bool `json:"ghn-mux"`
		Wifi   bool `json:"wifi"`
	} `json:"features"`
}

type Node struct {
	Mac            string  `json:"mac"`
	RxErrorPercent float64 `json:"rxErrorPercent"`
	Modem          string  `json:"modem"`
	TxRate         float64 `json:"txRate"`
	RxRate         float64 `json:"rxRate"`
	DeviceID       int     `json:"deviceId"`
	RxAbortPercent float64 `json:"rxAbortPercent"`
	Power          struct {
		Min float64 `json:"min"`
		Agc int     `json:"agc"`
		Avg float64 `json:"avg"`
		Max float64 `json:"max"`
	} `json:"power"`
	RxFrames int `json:"rxFrames"`
	Snr      struct {
		Min float64 `json:"min"`
		Avg float64 `json:"avg"`
		Max float64 `json:"max"`
	} `json:"snr"`
	WireLength int `json:"wireLength"`
	RxLPDUs    int `json:"rxLPDUs"`
}

const matricsPath = "cgi.lua/status?type=system,ghn,ethernet,remote"

type metricsResponse struct {
	System System `json:"system"`
	Ghn    struct {
		Timestamp int              `json:"timestamp"`
		Nodes     map[string]Node  `json:"nodes"`
		Modems    map[string]Modem `json:"modems"`
	} `json:"ghn"`
	Ethernet struct {
		Ts    int                     `json:"ts"`
		Ports map[string]EthernetPort `json:"ports"`
	} `json:"ethernet"`
	Remote map[string]Remote `json:"remote"`
}

type Counters struct {
	RxBcast  int `json:"rx_bcast"`
	RxByte   int `json:"rx_byte"`
	RxDrop   int `json:"rx_drop"`
	RxErr    int `json:"rx_err"`
	RxMcast  int `json:"rx_mcast"`
	RxPacket int `json:"rx_packet"`
	RxUcast  int `json:"rx_ucast"`
	TxBcast  int `json:"tx_bcast"`
	TxByte   int `json:"tx_byte"`
	TxDrop   int `json:"tx_drop"`
	TxErr    int `json:"tx_err"`
	TxMcast  int `json:"tx_mcast"`
	TxPacket int `json:"tx_packet"`
	TxUcast  int `json:"tx_ucast"`
}

type EthernetPort struct {
	Access      string `json:"access"`
	Autoneg     bool   `json:"autoneg"`
	Index       int    `json:"index"`
	Link        bool   `json:"link"`
	RxBroadcast int    `json:"rxBroadcast"`
	RxBytes     int    `json:"rxBytes"`
	RxDrops     int    `json:"rxDrops"`
	RxErrors    int    `json:"rxErrors"`
	RxMulticast int    `json:"rxMulticast"`
	RxPackets   int    `json:"rxPackets"`
	RxRate      int    `json:"rxRate"`
	RxUnicast   int    `json:"rxUnicast"`
	Switch      string `json:"switch"`
	TxBroadcast int    `json:"txBroadcast"`
	TxBytes     int    `json:"txBytes"`
	TxDrops     int    `json:"txDrops"`
	TxErrors    int    `json:"txErrors"`
	TxMulticast int    `json:"txMulticast"`
	TxPackets   int    `json:"txPackets"`
	TxRate      int    `json:"txRate"`
	TxUnicast   int    `json:"txUnicast"`
}

type Modem struct {
	TxUnicast      int     `json:"txUnicast"`
	CPUUsage       int     `json:"cpuUsage"`
	RxUnicast      int     `json:"rxUnicast"`
	MemUsage       int     `json:"memUsage"`
	CombiningGroup string  `json:"combiningGroup"`
	Speed          float64 `json:"speed"`
	OperationTime  int     `json:"operationTime"`
	RxErrors       int     `json:"rxErrors"`
	Noise          struct {
		Min float64 `json:"min"`
		Agc int     `json:"agc"`
		Avg float64 `json:"avg"`
		Max float64 `json:"max"`
	} `json:"noise"`
	OperationMode      string  `json:"operationMode"`
	LinkLost           int     `json:"linkLost"`
	RxMulticast        int     `json:"rxMulticast"`
	Name               string  `json:"name"`
	EndpointRegistered int     `json:"endpointRegistered"`
	TxMulticast        int     `json:"txMulticast"`
	Index              int     `json:"index"`
	RetxPercent        float64 `json:"retx_percent"`
	Port               string  `json:"port"`
	Switch             string  `json:"switch"`
	TxRate             int     `json:"txRate"`
	DomainID           string  `json:"domainId"`
	RxRate             int     `json:"rxRate"`
	RxBroadcast        int     `json:"rxBroadcast"`
	DomainName         string  `json:"domainName"`
	TxPackets          int     `json:"txPackets"`
	RxBytes            int64   `json:"rxBytes"`
	RxPackets          int     `json:"rxPackets"`
	ResetCause         string  `json:"resetCause"`
	FecPercent         float32 `json:"fec_percent"`
	TxBytes            int64   `json:"txBytes"`
	Firmware           string  `json:"firmware"`
	RxBlocksError      int     `json:"rxBlocksError"`
	Ipv4               string  `json:"ipv4"`
	OperationSlot      int     `json:"operationSlot"`
	Master             string  `json:"master"`
	EndpointCount      int     `json:"endpointCount"`
	Mac                string  `json:"mac"`
	RxBlocks           int     `json:"rxBlocks"`
	TxBroadcast        int     `json:"txBroadcast"`
	TxBlocks           int     `json:"txBlocks"`
	TxDrops            int     `json:"txDrops"`
	DomainMode         string  `json:"domainMode"`
	TxBlocksResent     int     `json:"txBlocksResent"`
	TxErrors           int     `json:"txErrors"`
	Uptime             int     `json:"uptime"`
	ResetMarker        int     `json:"resetMarker"`
	RxDrops            int     `json:"rxDrops"`
}

type System struct {
	Description string `json:"description"`
	Board       string `json:"board"`
	ImagesValid bool   `json:"images_valid"`
	Uptime      int    `json:"uptime"`
	Version     string `json:"version"`
	Contact     string `json:"contact"`
	Memory      struct {
		Used  int     `json:"used"`
		Usage float64 `json:"usage"`
		Total int     `json:"total"`
		Free  int     `json:"free"`
	} `json:"memory"`
	Name      string `json:"name"`
	Clock     string `json:"clock"`
	Location  string `json:"location"`
	Timestamp int    `json:"timestamp"`
	Processes int    `json:"processes"`
	CPU       struct {
		Usage5Min  float64 `json:"usage5min"`
		Usage      float64 `json:"usage"`
		Usage15Min float64 `json:"usage15min"`
		Usage1Min  float64 `json:"usage1min"`
	} `json:"cpu"`
}

type Remote struct {
	EthernetClients []any `json:"ethernet_clients"`
	WirelessClients []struct {
		Mac      string `json:"mac"`
		Per      int    `json:"per"`
		Ssid     string `json:"ssid"`
		Ipaddr   string `json:"ipaddr"`
		Protocol string `json:"protocol"`
		Radio    string `json:"radio"`
		Bitrate  struct {
			Tx int `json:"tx"`
			Rx int `json:"rx"`
		} `json:"bitrate"`
		Band     int    `json:"band"`
		Uptime   int    `json:"uptime"`
		Hostname string `json:"hostname"`
		Packets  struct {
			Tx      int `json:"tx"`
			RxError int `json:"rx_error"`
			Rx      int `json:"rx"`
			TxError int `json:"tx_error"`
		} `json:"packets"`
		Throughput struct {
			Tx int `json:"tx"`
			Rx int `json:"rx"`
		} `json:"throughput"`
		Network string `json:"network"`
		Signal  int    `json:"signal"`
	} `json:"wireless_clients"`
	Group    string `json:"group"`
	Loadtime int    `json:"loadtime"`
	Wireless []struct {
		Band         int      `json:"band"`
		Clients      int      `json:"clients"`
		Label        string   `json:"label"`
		Mac          string   `json:"mac"`
		Bitrate      int      `json:"bitrate"`
		Txpower      int      `json:"txpower"`
		ChannelWidth string   `json:"channel_width"`
		Channel      int      `json:"channel"`
		Counters     Counters `json:"counters"`
		Radio        string   `json:"radio"`
		Enabled      bool     `json:"enabled"`
		Frequency    int      `json:"frequency"`
	} `json:"wireless"`
	Status     string `json:"status"`
	ConfigHash string `json:"config_hash"`
	Mac        string `json:"mac"`
	Ghn        []struct {
		RetxPercent float32 `json:"retx_percent"`
		Clients     int     `json:"clients"`
		Label       string  `json:"label"`
		Snr         struct {
			Min float64 `json:"min"`
			Avg float64 `json:"avg"`
			Max float64 `json:"max"`
		} `json:"snr"`
		Port       int     `json:"port"`
		Status     bool    `json:"status"`
		FecPercent float32 `json:"fec_percent"`
		Ipv4       string  `json:"ipv4"`
		Enabled    bool    `json:"enabled"`
		Noise      struct {
			Min float64 `json:"min"`
			Agc int     `json:"agc"`
			Avg float64 `json:"avg"`
			Max float64 `json:"max"`
		} `json:"noise"`
		Power struct {
			Min float64 `json:"min"`
			Agc int     `json:"agc"`
			Avg float64 `json:"avg"`
			Max float64 `json:"max"`
		} `json:"power"`
		Bitrate struct {
			Tx int `json:"tx"`
			Rx int `json:"rx"`
		} `json:"bitrate"`
		ResetCause string   `json:"reset_cause"`
		Master     string   `json:"master"`
		Uptime     int      `json:"uptime"`
		MemUsage   int      `json:"mem_usage"`
		Counters   Counters `json:"counters"`
		Mac        string   `json:"mac"`
		Speed      float64  `json:"speed"`
		CPUUsage   int      `json:"cpu_usage"`
	} `json:"ghn"`
	System struct {
		SerialNo    string `json:"serial_no"`
		Description string `json:"description"`
		HostMac     string `json:"host_mac"`
		MemStat     struct {
			Total int `json:"total"`
			Free  int `json:"free"`
		} `json:"mem_stat"`
		Hostname   string    `json:"hostname"`
		MacAddress string    `json:"mac_address"`
		SwVersion  string    `json:"sw_version"`
		Datetime   time.Time `json:"datetime"`
		Model      string    `json:"model"`
		Uptime     *uint     `json:"uptime"`
		MemUsage   float64   `json:"mem_usage"`
		Name       string    `json:"name"`
		Timestamp  int       `json:"timestamp"`
		GhnVersion string    `json:"ghn_version"`
		CPUUsage   float64   `json:"cpu_usage"`
	} `json:"system"`
	Message  string `json:"message"`
	PortName string `json:"port_name"`
	State    int    `json:"state"`
	Ethernet []struct {
		Enabled  bool     `json:"enabled"`
		Port     int      `json:"port"`
		Duplex   bool     `json:"duplex"`
		Label    string   `json:"label"`
		Counters Counters `json:"counters"`
		Autoneg  bool     `json:"autoneg"`
		Link     bool     `json:"link"`
		Mac      string   `json:"mac"`
	} `json:"ethernet"`
	Seen      int    `json:"seen"`
	Timestamp int    `json:"timestamp"`
	Serial    string `json:"serial"`
	Network   struct {
		Areas []struct {
			Network string `json:"network"`
			Ipv4    struct {
				Prefix  int    `json:"prefix"`
				Address string `json:"address"`
			} `json:"ipv4,omitempty"`
		} `json:"areas"`
		Ipv6 struct {
			Prefix  int    `json:"prefix"`
			Address string `json:"address"`
		} `json:"ipv6"`
		Ipv4 struct {
			Prefix  int    `json:"prefix"`
			Address string `json:"address"`
		} `json:"ipv4"`
	} `json:"network"`
}
