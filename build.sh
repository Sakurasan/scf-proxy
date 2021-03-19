
#!/bin/bash
work_path=$(cd `dirname $0`/../; pwd)
echo $work_path

GOOS=linux GOARCH=amd64 go build -o main cmd/main.go
zip main.zip main