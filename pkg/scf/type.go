package scf

// DefineEvent 请求结构
type DefineEvent struct {
	URL     string `json:"url"`     // 目标的 URL, eg: http://cip.cc/
	Content string `json:"content"` // 最原始的 HTTP 报文, base64
}

// RespEvent 响应结构
type RespEvent struct {
	Status bool   `json:"status"` // 请求是否正常
	Error  string `json:"error"`  // 错误信息
	Data   string `json:"data"`   // HTTP 响应原始报文, base64
}
