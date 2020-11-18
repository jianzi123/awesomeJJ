#!/bin/bash

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

go build -o client-go-exec-plugin

./client-go-exec-plugin --addr=10.1.97.101

