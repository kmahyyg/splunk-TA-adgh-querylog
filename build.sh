#!/bin/bash

rm -ri bin/*

export GOOS=linux
export GOARCH=amd64
go build -ldflags='-s -w' -trimpath -o bin/adgh_qlog_preproc.elf ./cmd/main.go

export GOOS=darwin
export GOARCH=arm64
go build -ldflags='-s -w' -trimpath -o bin/adgh_qlog_preproc.mach-o ./cmd/main.go
