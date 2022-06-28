package scf

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/tencentyun/scf-go-lib/events"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func Handler(ctx context.Context, reqOrigin events.APIGatewayRequest) (resp events.APIGatewayResponse) {
	var reqEvent = new(DefineEvent)
	//b64dreqOrigin, _ := base64.StdEncoding.DecodeString(reqOrigin.Body)
	if err := json.Unmarshal([]byte(reqOrigin.Body), reqEvent); err != nil {
		return handlerErr(reqOrigin, err.Error())
	}
	proxyresp, err := forworld(reqEvent)
	if err != nil {
		return handlerErr(reqOrigin, err.Error())
	}
	body, err := json.Marshal(proxyresp)
	if err != nil {
		return handlerErr(reqOrigin, err.Error())
	}
	resp = events.APIGatewayResponse{
		IsBase64Encoded: false,
		StatusCode:      200,
		Headers:         map[string]string{"ContentType": "application/json"},
		Body:            string(body),
	}
	return
}

func forworld(reqevent *DefineEvent) (*RespEvent, error) {
	var (
		respvent = new(RespEvent)
	)
	rawreq, err := base64.StdEncoding.DecodeString(reqevent.Content)
	if err != nil {
		return nil, err
	}
	originreq, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(rawreq)))
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest(originreq.Method, originreq.RequestURI, originreq.Body)
	for k, v := range originreq.Header {
		req.Header.Set(k, v[0])
	}
	// EOF
	req.Close = true
	tr := &http.Transport{
		Dial: (&net.Dialer{
			//LocalAddr: localAddr,
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		//DialContext: (&net.Dialer{
		//	Timeout:   30 * time.Second,
		//	KeepAlive: 30 * time.Second,
		//}).DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: tr, Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("client.Do()", err)
		return nil, err
	}
	defer resp.Body.Close()

	byteresp, _ := ioutil.ReadAll(resp.Body)
	byterespHeaders, _ := json.Marshal(resp.Header)
	respHeaders := base64.StdEncoding.EncodeToString(byterespHeaders)
	//响应headers
	//{"Accept-Ranges":["bytes"],
	//"Content-Length":["573"],
	//"Content-Type":["text/html; charset=utf-8"],
	//"Date":["Mon, 27 Jun 2022 02:56:52 GMT"],
	//"Etag":["\"61419ca8-23d\""],
	//"Last-Modified":["Wed, 15 Sep 2021 07:11:36 GMT"],
	//"Server":["nginx/1.20.1"]}
	respBoby := base64.StdEncoding.EncodeToString(byteresp)
	respvent.Data = respHeaders + "^" + respBoby
	//respvent.Data = base64.StdEncoding.EncodeToString(byteresp)
	respvent.Status = true

	return respvent, nil
}

//handleErr 处理错误
func handlerErr(reqOrigin events.APIGatewayRequest, errString string) (resp events.APIGatewayResponse) {
	// log
	log.Printf("[出现错误] \n//err %v \n//req %v \n========== \n", errString, reqOrigin)

	// handle error
	j, _ := json.Marshal(errString)
	resp = events.APIGatewayResponse{
		IsBase64Encoded: false,
		StatusCode:      500,
		Headers:         map[string]string{"ContentType": "application/json"},
		Body:            string(j),
	}
	return
}
