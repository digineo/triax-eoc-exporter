package exporter

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"git.digineo.de/digineo/triax_eoc_exporter/config"
	"git.digineo.de/digineo/triax_eoc_exporter/triax"
	"github.com/digineo/goldflags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Start(cfg *config.Config) {
	triax.Verbose = true

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, indexHTML, goldflags.VersionString())
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		reg := prometheus.NewRegistry()
		// TODO: enable (noisy in development)
		// reg.MustRegister(prometheus.NewGoCollector())

		target := r.URL.Query().Get("target")

		var err error
		var client *triax.Client

		for _, ctrl := range cfg.Controllers {
			if target == ctrl.Alias || target == ctrl.Host {
				client, err = triax.NewClient(&url.URL{
					Scheme: "https",
					User:   url.UserPassword("admin", ctrl.Password),
					Host:   ctrl.Host,
					Path:   "/",
				})
				if err != nil {
					log.Printf("error constructing client for %q: %v", ctrl.Host, err)
				}
				break
			}
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		if client == nil {
			http.Error(w, "configuration not found", http.StatusNotFound)
			return
		}

		reg.MustRegister(&triaxCollector{
			client: client,
			ctx:    r.Context(),
		})
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
