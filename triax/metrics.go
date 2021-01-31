package triax

type EndpointMetrics struct {
	Name          string
	MAC           string
	Status        int
	StatusText    string
	Uptime        uint
	Load          float64
	GhnPortNumber int
	GhnPortMac    string
	GhnStats      *GhnStats
	Statistics    Statistics
}

type GhnPort struct {
	Number              int
	EndpointsOnline     uint
	EndpointsRegistered uint
}

type Metrics struct {
	Uptime uint
	Load   float64
	Memory struct {
		Total    uint
		Free     uint
		Buffered uint
		Shared   uint
	}
	GhnPorts  map[string]*GhnPort
	Endpoints []EndpointMetrics
}
