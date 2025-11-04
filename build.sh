#!/bin/bash -ex
cd "$(dirname "$0")"
go install tool
mage build coverage
cat report.out

for arch in amd64 arm64; do
  GOOS=linux GOARCH=$arch go build -o example_server/syslog.$arch ./example_server
done