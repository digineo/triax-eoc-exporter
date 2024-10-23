Triax EoC Exporter
==================

[![Build](https://github.com/digineo/triax-eoc-exporter/actions/workflows/go.yml/badge.svg)](https://github.com/digineo/triax-eoc-exporter/actions/workflows/go.yml)
[![Codecov](http://codecov.io/github/digineo/triax-eoc-exporter/coverage.svg?branch=master)](http://codecov.io/github/digineo/triax-eoc-exporter?branch=master)

This is a [Prometheus](https://prometheus.io/) exporter for
[Triax EoC controllers](https://www.triax.com/products/ethernet-over-coax).
It has been tested with the [EoC controller software](https://www.triax.com/product/ethernet-over-coax-software-update/) version 3.4.7.

## Features

* Exporting metrics for prometheus
* HTTP-Proxy for reading/writing controller configurations

## Installation

Use one of the ready-to-use [releases](https://github.com/digineo/triax-eoc-exporter/releases), or compile and install it using the Go toolchain:

    go install github.com/digineo/triax-eoc-exporter@latest

## Configuration

### Exporter

List all your controllers in the config.toml file.

If you use the Debian package, just edit `/etc/triax-eoc-exporter/config.toml` and restart the exporter by running `systemctl restart triax-eoc-exporter`.
Modify the start parameters in `/etc/defaults/triax-eoc-exporter` if you want the controller to bind on other addresses than localhost.


After starting the controller, just visit http://localhost:9809/
You will see a list of all configured controllers and links to the corresponding metrics endpoints.

### Prometheus

Add a scrape config to your Prometheus configuration and reload Prometheus.

```yaml
scrape_configs:
  - job_name: triax-eoc
    relabel_configs:
      - source_labels: [__address__]
        regex:         (.+)
        target_label:  __metrics_path__
        replacement:   /controllers/$1/metrics
      - source_labels: [__address__]
        target_label: instance
      - target_label: __address__
        replacement: 127.0.0.1:9809 # The exporter's real hostname:port
    static_configs:
      - targets: # list the configured aliases below
        - my-controller
        - another-controller
```

## Endpoint Status

* 1 OK
* 2 configuring
* 4 updating
* 8 offline (responding)
* 9 offline (detected)
* 10 offline
