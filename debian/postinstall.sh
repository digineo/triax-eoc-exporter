#!/bin/sh

systemctl daemon-reload
systemctl enable triax-eoc-exporter
systemctl restart triax-eoc-exporter
