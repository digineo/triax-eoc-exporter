package exporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"text/template"

	"github.com/digineo/triax-eoc-exporter/triax"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (cfg *Config) Start(listenAddress, version string) {
	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		tmpl.Execute(w, &indexVariables{
			Controllers: cfg.Controllers,
			Version:     version,
		})
	})

	router.GET("/controllers", cfg.listControllersHandler)
	router.GET("/controllers/:target/metrics", cfg.targetMiddleware(cfg.metricsHandler))
	router.GET("/controllers/:target/api/*path", cfg.targetMiddleware(cfg.apiHandler))
	router.PUT("/controllers/:target/nodes/:mac", cfg.targetMiddleware(cfg.updateNodeHandler))

	log.Printf("Starting exporter on http://%s/", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, router))
}

type targetHandler func(*triax.Client, http.ResponseWriter, *http.Request, httprouter.Params)

func (cfg *Config) targetMiddleware(next targetHandler) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		target := params.ByName("target")
		client, err := cfg.getClient(target)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		if client == nil {
			http.Error(w, "configuration not found", http.StatusNotFound)
			return
		}

		next(client, w, r, params)
	})
}

func (cfg *Config) listControllersHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.Body.Close()
	result := make([]string, len(cfg.Controllers))

	for i := range cfg.Controllers {
		result[i] = cfg.Controllers[i].Alias
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&result)
}

func (cfg *Config) metricsHandler(client *triax.Client, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(&triaxCollector{
		client: client,
		ctx:    r.Context(),
	})
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

// handler for updating nodes
func (cfg *Config) updateNodeHandler(client *triax.Client, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	defer r.Body.Close()

	// parse MAC address parameter
	mac, err := net.ParseMAC(params.ByName("mac"))
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid MAC address: %s", err), http.StatusBadRequest)
		return
	}

	// decode request body
	req := triax.UpdateRequest{}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// execute update
	err = client.UpdateNode(r.Context(), mac, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// proxy handler for API GET requests.
func (cfg *Config) apiHandler(client *triax.Client, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	defer r.Body.Close()

	msg := json.RawMessage{}
	err := client.Get(r.Context(), params.ByName("path"), &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	io.Copy(w, bytes.NewReader(msg))
}

type indexVariables struct {
	Controllers []Controller
	Version     string
}

var tmpl = template.Must(template.New("index").Option("missingkey=error").Parse(`<!doctype html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Triax EoC Exporter (Version {{.Version}})</title>
</head>
<body>
	<h1>Triax EoC Exporter</h1>
	<p>Version: {{.Version}}</p>

	<h2>Controllers</h2>
	<p><a href="/controllers">List as JSON</a></p>
	<dl>
	{{range .Controllers}}
		<dt>{{.Alias}}</dt>
		<dd>
			<a href="/controllers/{{.Alias}}/metrics">Metrics</a>,
			<a href="/controllers/{{.Alias}}/api/node/status/">Status</a>
		</dd>
	{{end}}
	</dl>

</body>
</html>
`))
