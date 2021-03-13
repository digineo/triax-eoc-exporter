#!/bin/sh

case "$1" in
    remove)
        systemctl disable triax-eoc-exporter || true
        systemctl stop triax-eoc-exporter    || true
    ;;
esac
