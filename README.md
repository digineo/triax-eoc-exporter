Triax EoC Exporter
==================

This is a [Prometheus](https://prometheus.io/) exporter for
[Triax EoC controllers](https://www.triax.com/products/ethernet-over-coax).

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
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: 127.0.0.1:9809 # The exporter's real hostname:port
    static_configs:
      - targets: # list the configured aliases below
        - my-controller
        - another-controller
```
