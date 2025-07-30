#!/bin/bash -e
cd "$(dirname "$0")"

go test ./...
go vet ./...
gofmt -l -s -w *.go */*.go
go build -o syslog ./example_server