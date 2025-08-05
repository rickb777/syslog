#!/bin/bash -e
cd "$(dirname "$0")"

go test ./...
go vet ./...
gofmt -l -s -w *.go */*.go

for arch in amd64 arm64; do
  GOOS=linux GOARCH=$arch go build -o example_server/syslog.$arch ./example_server
done