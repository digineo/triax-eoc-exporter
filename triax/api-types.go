package triax

import (
	"bytes"
	"encoding/json"
	"strings"
)

const loginPath = "login/"

// request for /api/login.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// response from /api/login.
type loginResponse struct {
	Cookie  string `json:"cookie"`
	Message string `json:"message"`
}

const ghnStatusPath = "ghn/status/"

// response from /api/ghn/status.
type ghnStatusResponse []struct {
	Connected uint   `json:"connected"`
	Mac       string `json:"mac"`
	// Name       string `json:"name"`
	Registered uint `json:"registered"`
}

const sysinfoPath = "system/info/"

// response from /api/system/info.
type sysinfoResponse struct {
	Memory struct {
		Buffered uint `json:"buffered"`
		Free     uint `json:"free"`
		Shared   uint `json:"shared"`
		Total    uint `json:"total"`
	} `json:"memory"`
	Uptime uint    `json:"uptime"`
	Load   float64 `json:"load"`
}

const syseocPath = "config/system/eoc/"

// response from /config/system/eoc.
type syseocResponse struct {
	MacAddr MacAddrList `json:"macaddr"`
}

type MacAddrList []string

func (mal *MacAddrList) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err // nolint go-lint
	}
	addrs := strings.Fields(s)
	for i := range addrs {
		addrs[i] = strings.ToLower(addrs[i])
	}
	*mal = addrs
	return nil
}

func (mal MacAddrList) Index(s string) int {
	s = strings.ToLower(s)
	for i, mac := range mal {
		if mac == s {
			return i
		}
	}
	return -1
}

const nodeStatusPath = "node/status/"

// response from /api/node/status/. The key is a mangeled form
// of the AP's MAC address, and should be ignored.
type nodeStatusResponse map[string]struct {
	Clients      []Clients  `json:"clients"`
	GhnMaster    string     `json:"ghn_master"`
	GhnStats     *GhnStats  `json:"ghn_stats"`
	Mac          string     `json:"mac"`
	Name         string     `json:"name"`
	RegTimestamp string     `json:"regts"`
	Serial       string     `json:"serial"`
	Statistics   Statistics `json:"statistics"`
	Status       string     `json:"status"`
	Statusid     int        `json:"statusid"`
	Sysinfo      struct {
		Load   float64 `json:"load"`
		Uptime uint    `json:"uptime"`
	} `json:"sysinfo"`
}

type Clients struct {
	Band    string `json:"band"`    // "5 GHz"
	Channel string `json:"channel"` // 120
}

type GhnStats struct {
	Abort  string `json:"abort"`
	Error  string `json:"error"`
	Frames string `json:"frames"`
	Lpdus  string `json:"lpdus"`
	Rxbps  uint   `json:"rxbps"`
	Txbps  uint   `json:"txbps"`
}

type Counters struct {
	RxBcast  uint    `json:"rx_bcast"`
	RxByte   uint64  `json:"rx_byte"`
	RxDrop   uint    `json:"rx_drop"`
	RxErr    uint    `json:"rx_err"`
	RxMcast  float64 `json:"rx_mcast"`
	RxPacket uint    `json:"rx_packet"`
	RxUcast  uint    `json:"rx_ucast"`
	TxBcast  uint    `json:"tx_bcast"`
	TxByte   uint64  `json:"tx_byte"`
	TxDrop   uint    `json:"tx_drop"`
	TxErr    uint    `json:"tx_err"`
	TxMcast  float64 `json:"tx_mcast"`
	TxPacket uint    `json:"tx_packet"`
	TxUcast  uint    `json:"tx_ucast"`
}

type Ethernet struct {
	Autoneg  bool     `json:"autoneg"`
	Duplex   bool     `json:"duplex"`
	Enabled  bool     `json:"enabled"`
	Link     bool     `json:"link"`
	Port     int      `json:"port"`
	Speed    int      `json:"speed"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Counters Counters `json:"counters"`
}

type Ghn struct {
	Clients  int      `json:"clients"`
	Counters Counters `json:"counters"`
	Label    string   `json:"label"`
	Mac      string   `json:"mac"`
	Port     int      `json:"port"`
	Enabled  bool     `json:"enabled"`
	Status   bool     `json:"status"`
}

type System struct {
	CPUUsage   float64 `json:"cpu_usage"`
	MacAddress string  `json:"mac_address"`
	MemStat    struct {
		Free  int `json:"free"`
		Total int `json:"total"`
	} `json:"mem_stat"`
	Uptime int `json:"uptime"`
}

type Statistics struct {
	Ethernet  []Ethernet `json:"ethernet"`
	Ghn       []Ghn      `json:"ghn"`
	System    System     `json:"system"`
	Timestamp int        `json:"timestamp"`
	Wireless  []struct {
		Band     int      `json:"band"`
		Clients  int      `json:"clients"`
		Counters Counters `json:"counters"`
		Enabled  bool     `json:"enabled"`
		Mac      string   `json:"mac"`
		Radio    string   `json:"radio"`
	} `json:"wireless"`
}
