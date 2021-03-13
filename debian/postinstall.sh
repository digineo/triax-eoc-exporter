#!/bin/sh

groupadd --system triax-eoc-exporter || true
useradd --system -d /nonexistent -s /usr/sbin/nologin -g triax-eoc-exporter triax-eoc-exporter || true

chown triax-eoc-exporter /etc/triax-eoc-exporter/*.toml

systemctl daemon-reload
systemctl enable triax-eoc-exporter
systemctl restart triax-eoc-exporter
