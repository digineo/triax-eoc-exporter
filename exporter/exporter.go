package exporter

import (
	"fmt"
	"log"
	"net/http"

	"git.digineo.de/digineo/triax_eoc_exporter/config"
	"git.digineo.de/digineo/triax_eoc_exporter/triax"
	"github.com/digineo/goldflags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Start(cfg *config.Config) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, indexHTML, goldflags.VersionString())
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		reg := prometheus.NewRegistry()
		// TODO: enable (noisy in development)
		// reg.MustRegister(prometheus.NewGoCollector())

		for _, ctrl := range cfg.Controllers {
			client, err := triax.NewClient(ctrl.Endpoint, ctrl.Insecure, ctrl.MAC)
			if err != nil {
				log.Printf("[skip] error constructing client for %q: %v", ctrl.Endpoint, err)
				continue
			}

			reg.MustRegister(&triaxCollector{
				client: client,
				ctx:    r.Context(),
			})
		}

		h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	})

	log.Printf("Starting exporter on http://%s/", cfg.Bind)
	log.Fatal(http.ListenAndServe(cfg.Bind, nil))
}

const indexHTML = `<!doctype html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Triax EoC Exporter (Version %s)</title>
</head>
<body>
	<h1>Triax EoC Exporter</h1>
	<p><a href="/metrics">Metrics</a></p>
</body>
</html>
`
