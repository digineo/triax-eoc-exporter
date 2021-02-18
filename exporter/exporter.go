package exporter

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/digineo/triax-eoc-exporter/triax"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (cfg *Config) Start(listenAddress string) {
	http.Handle("/metrics", cfg.targetMiddleware(cfg.metricsHandler))
	http.Handle("/api/", cfg.targetMiddleware(cfg.apiHandler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != "/" {
			http.NotFound(w, r)
			return
		}

		tmpl.Execute(w, &indexVariables{
			Controllers: cfg.Controllers,
		})
	})

	log.Printf("Starting exporter on http://%s/", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

type targetHandler func(*triax.Client, http.ResponseWriter, *http.Request)

func (cfg *Config) targetMiddleware(next targetHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")

		if target == "" {
			http.Error(w, "target parameter missing", http.StatusBadRequest)
			return
		}

		client, err := cfg.getClient(target)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		if client == nil {
			http.Error(w, "configuration not found", http.StatusNotFound)
			return
		}

		next(client, w, r)
	})
}

func (cfg *Config) metricsHandler(client *triax.Client, w http.ResponseWriter, r *http.Request) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(&triaxCollector{
		client: client,
		ctx:    r.Context(),
	})
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

// proxy handler for API GET requests.
func (cfg *Config) apiHandler(client *triax.Client, w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(r.RequestURI, "/", 3)

	if len(parts) != 3 {
		http.Error(w, "path missing", http.StatusNotFound)
		return
	}

	defer r.Body.Close()

	msg := json.RawMessage{}
	err := client.Get(r.Context(), parts[2], &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	io.Copy(w, bytes.NewReader(msg))
}

type indexVariables struct {
	Controllers []Controller
}

var tmpl = template.Must(template.New("index").Option("missingkey=error").Parse(`<!doctype html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Triax EoC Exporter (Version %s)</title>
</head>
<body>
	<h1>Triax EoC Exporter</h1>

	<h2>Metrics</h2>
	<ol>
	{{range .Controllers}}
		<li><a href="/metrics?target={{.Alias }}">{{.Alias}}</a></li>
	{{end}}
	</ol>

	<h2>Node Status</h2>
	<ol>
	{{range .Controllers}}
		<li><a href="/api/node/status/?target={{.Alias }}">{{.Alias}}</a></li>
	{{end}}
	</ol>

</body>
</html>
`))
