#!/bin/bash

set -e
set -u

(
    cd cmd/win-dns-api-go
    GOOS=windows go build
)

(
    cd cmd/win-dns-to-bind
    go build
    GOOS=windows go build
)

cp cmd/win-dns-api-go/win-dns-api-go.exe .

cp cmd/win-dns-to-bind/win-dns-to-bind .
cp cmd/win-dns-to-bind/win-dns-to-bind.exe .
