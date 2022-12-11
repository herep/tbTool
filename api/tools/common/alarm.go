package common

type Response struct {
	Code    int32                  `json:"code"`
	Msg     string                 `json:"msg"`
	NowTime int64                  `json:"nowTime,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type ResponseInterface struct {
	Code    int32       `json:"code"`
	Msg     string      `json:"msg"`
	NowTime int64       `json:"nowTime,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}