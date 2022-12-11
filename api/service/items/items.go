package items

import (
	"bytes"
	"time"
)

const (
	AppKey      = "21593345"
	GatewayLink = "http://openapi.cdshoes.cn/OpenApi/Call/Dev13074885409"
)

type ItemService interface {
	GetTaoBaoItemsUrl(method, sign, session string, requestMap map[string]string) (itemUrl string)
}

type ItemServiceImpl struct{}

func NewItemServiceImpl() ItemService {
	return &ItemServiceImpl{}
}

//淘宝商品请求生成接口
func (is *ItemServiceImpl) GetTaoBaoItemsUrl(method, sign, session string ,requestMap map[string]string) (itemUrl string) {
	var buff bytes.Buffer
	buff.WriteString(GatewayLink)
	buff.WriteString("&app_key=")
	buff.WriteString(AppKey)
	buff.WriteString("&method=")
	buff.WriteString(method)
	buff.WriteString("&v=2.0")
	buff.WriteString("&sign=")
	buff.WriteString(sign)
	buff.WriteString("&timestamp=")
	buff.WriteString(time.Now().Format("2006-01-02 15:04:05"))
	buff.WriteString("&partner_id=top-apitools")
	buff.WriteString("&session=")
	buff.WriteString(session)
	buff.WriteString(requestObjectJson(requestMap))
	buff.WriteString("&format=json&sign_method=md5")


	return buff.String()
}

//参数拼接
func requestObjectJson(requestMap map[string]string) string {
	var buff bytes.Buffer

	buff.WriteString("&")
	for k,v :=range requestMap {
		buff.WriteString(k+"=")
		buff.WriteString(v)
		buff.WriteString("&")
	}

	return buff.String()
}
