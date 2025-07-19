#!/bin/bash
work_path=$(cd `dirname $0`/../; pwd)
echo $work_path

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -ldflags '-s -w' -gcflags="all=-trimpath=${PWD}" -asmflags="all=-trimpath=${PWD}" -o main cmd/main.go
zip main.zip main