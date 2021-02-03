package exporter

import (
	"log"
	"net/http"
	"text/template"

	"git.digineo.de/digineo/triax-eoc-exporter/triax"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (cfg *Config) Start(listenAddress string) {
	http.Handle("/metrics", cfg.targetMiddleware(cfg.metricsHandler))
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

</body>
</html>
`))
