#!/bin/bash
set -e
set -u

(
	cd cmd/win-dns-api-go
        GOOS=windows go build
)

cp cmd/win-dns-api-go/win-dns-api-go.exe .
