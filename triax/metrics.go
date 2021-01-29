package triax

type EndpointMetrics struct {
	Name          string
	MAC           string
	Status        int
	StatusText    string
	Uptime        int
	Load          float64
	GhnPortNumber int
	GhnPortMac    string
	Clients       map[string]int
}

type GhnPort struct {
	Number              int
	EndpointsOnline     int
	EndpointsRegistered int
}

type Metrics struct {
	Up int

	Uptime int
	Load   float64
	Memory struct {
		Total    int
		Free     int
		Buffered int
		Shared   int
	}
	GhnPorts  map[string]*GhnPort
	Endpoints []*EndpointMetrics
}
