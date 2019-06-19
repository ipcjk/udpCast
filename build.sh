#!/usr/bin/env bash

env GOOS=linux GOARCH=amd64 go build --ldflags="-s -w" -o udpCast.linux udpCast.go udpcast_const.go
env GOOS=linux GOARCH=amd64 go build -tags perf -ldflags="-s -w" -o udpCast_perf.linux udpCast_perf.go udpcast_const.go

upx --force udpCast.linux
upx --force udpCast_perf.linux

