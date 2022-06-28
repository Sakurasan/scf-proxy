package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
	"github.com/tencentyun/scf-go-lib/events"
	"scf-proxy/pkg/scf"
)

func main() {
	// log.SetOutput(os.Stdout)
	cloudfunction.Start(server)
}
func init() {
	fmt.Println("scf-proxy-server")
	log.SetFormatter(new(Formatter))
}
func server(ctx context.Context, req events.APIGatewayRequest) (resp events.APIGatewayResponse, err error) {
	bytedata, _ := json.Marshal(req)
	log.Printf("[请求接收] \n => ctx \n%v \n => req \n%s \n========== \n\n", ctx, string(bytedata))

	return scf.Handler(ctx, req), nil
}

type Formatter struct{}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var buffer *bytes.Buffer
	if entry.Buffer != nil {
		buffer = entry.Buffer
	} else {
		buffer = &bytes.Buffer{}
	}

	buffer.WriteString("[")
	buffer.WriteString(entry.Time.Format("2006-01-02 15:04:05.999999999"))
	buffer.WriteString("]")

	buffer.WriteString("[")
	buffer.WriteString(entry.Level.String())
	buffer.WriteString("]")

	if entry.HasCaller() {
		buffer.Write([]byte("["))
		buffer.Write([]byte(fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)))
		buffer.Write([]byte("] "))

		// buffer.Write([]byte("[fn] "))
		// buffer.Write([]byte(entry.Caller.Func.Name()))
		// buffer.WriteString(" ")
	}
	buffer.WriteString(entry.Message)
	buffer.WriteString("\n")
	return buffer.Bytes(), nil
}
