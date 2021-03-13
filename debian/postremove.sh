#!/bin/sh

case "$1" in
    remove)
        systemctl daemon-reload
        userdel  triax-eoc-exporter || true
        groupdel triax-eoc-exporter 2>/dev/null || true
    ;;
esac
