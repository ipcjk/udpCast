#!/usr/bin/env bash

env GOOS=linux GOARCH=amd64 go build --ldflags="-s -w" -o udpCast.linux *.go

upx --force udpCast.linux

