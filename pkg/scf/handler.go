package scf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

var (
	ScfApiProxyUrl string
)

func HandlerHttp(w http.ResponseWriter, r *http.Request) {
	dumpReq, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event := &DefineEvent{
		URL:     r.URL.String(),
		Content: base64.StdEncoding.EncodeToString(dumpReq),
	}
	bytejson, err := json.Marshal(event)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	req, err := http.NewRequest("POST", ScfApiProxyUrl, bytes.NewReader(bytejson))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("client.Do()", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	bytersp, _ := ioutil.ReadAll(resp.Body)

	var respevent RespEvent
	if err := json.Unmarshal(bytersp, &respevent); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	if resp.StatusCode > 0 && resp.StatusCode != 200 {
		log.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	retByte, err := base64.StdEncoding.DecodeString(respevent.Data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	resp.Body.Close()

	w.Write(retByte)
	return
}
