package exporter

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"text/template"

	"github.com/digineo/triax-eoc-exporter/client"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/digineo/triax-eoc-exporter/backend/v3"
)

func (cfg *Config) Start(listenAddress, version, date string) {
	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		tmpl.Execute(w, &indexVariables{
			Controllers: cfg.Controllers,
			Version:     version,
			Date:        date,
		})
	})

	router.GET("/controllers", cfg.listControllersHandler)
	router.GET("/controllers/:target/metrics", cfg.targetMiddleware(cfg.metricsHandler))
	router.GET("/controllers/:target/config", cfg.targetMiddleware(cfg.getConfigHandler))
	router.POST("/controllers/:target/config", cfg.targetMiddleware(cfg.updateConfigHandler))

	slog.Info("Starting exporter", "listenAddress", listenAddress, "version", version, "builtDate", date)
	slog.Info("Server stopped", "reason", http.ListenAndServe(listenAddress, router))
}

type targetHandler func(*client.Client, http.ResponseWriter, *http.Request, httprouter.Params)

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

func (cfg *Config) metricsHandler(client *client.Client, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(&triaxCollector{
		client: client,
		ctx:    r.Context(),
	})
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

// handler for updating configs
func (cfg *Config) updateConfigHandler(client *client.Client, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	defer r.Body.Close()

	jsonBody := json.RawMessage{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&jsonBody)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = client.SetConfig(r.Context(), jsonBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handler for getting configs
func (cfg *Config) getConfigHandler(client *client.Client, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	config, err := client.GetConfig(r.Context())

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	io.Copy(w, bytes.NewReader(config))
}

type indexVariables struct {
	Controllers []Controller
	Version     string
	Date        string
}

var tmpl = template.Must(template.New("index").Option("missingkey=error").Parse(`<!doctype html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Triax EoC Exporter (Version {{.Version}})</title>
</head>
<body>
	<h1>Triax EoC Exporter</h1>
	<p>
		Version: {{.Version}}<br>
		Built at: {{.Date}}
	</p>

	<h2>Controllers</h2>
	<p><a href="/controllers">List as JSON</a></p>
	<dl>
	{{range .Controllers}}
		<dt>{{.Alias}}</dt>
		<dd>
			<a href="/controllers/{{.Alias}}/metrics">Metrics</a>,
			<a href="/controllers/{{.Alias}}/config">Config</a>
		</dd>
	{{end}}
	</dl>

</body>
</html>
`))
