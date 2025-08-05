#!/bin/bash -e

HOSTS=$*
HOSTS="$HOSTS $(uname -n) unknown"

case `uname -m` in
  aarch64|arm64) ARCH=arm64;;
  x86_64|amd64)  ARCH=amd64;;
  *) echo unknown ARCH
     uname -m
     exit 1
esac

sudo setcap 'cap_net_bind_service=+ep' syslog.$ARCH
sudo cp -vf syslog.$ARCH /usr/local/bin/syslog-lite
sudo cp -vf syslog-lite.service /etc/systemd/system/
sudo cp -vf syslog-lite.conf /etc/default/

for d in $HOSTS; do
  echo mkdir -p /var/log/$d
  sudo mkdir -p /var/log/$d
  sudo rm -f /var/log/$d/*
  sudo chown sys:adm /var/log/$d
done

sudo systemctl daemon-reload
