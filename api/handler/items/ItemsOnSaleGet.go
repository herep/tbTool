package items

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"tbTool/api/service/items"
	"tbTool/api/tools/common"
	"tbTool/pkg"
	"tbTool/pkg/request"
	"time"
)

const (
	TestSession = "6102205c5833518108ZZd5e2cc696ba487a790d5509e42c2211416907585"
	MethodName  = "taobao.items.onsale.get"
)

type ItemOnSaleGet struct {
	ItemsOnSaleGetResponse struct {
		Items struct {
			Item []struct {
				NumIid int64 `json:"num_iid"`
			} `json:"item"`
		} `json:"items"`
		TotalResults int    `json:"total_results"`
		RequestID    string `json:"request_id"`
	} `json:"items_onsale_get_response"`

	ErrorResponse struct {
		Code      int    `json:"code"`
		Msg       string `json:"msg"`
		RequestID string `json:"request_id"`
	} `json:"error_response"`
}
type ItemOnSaleGetHandler struct {
	is items.ItemService
}

func NewItemsOnSaleGetHandler(is items.ItemService) *ItemOnSaleGetHandler {
	return &ItemOnSaleGetHandler{
		is: is,
	}
}

//获取当前会话用户出售中的商品列表
func (ih *ItemOnSaleGetHandler) TaoBaoItemsOnSaleGet(c *gin.Context) pkg.Render{
	var info ItemOnSaleGet

	sign := c.MustGet("sign").(string)

	//参数
	requestMap := map[string]string{
		"fields": "num_iid,title,price",
	}

	_url := ih.is.GetTaoBaoItemsUrl(MethodName, sign, TestSession, requestMap)
	_, data, err := request.Get(_url, 2*time.Second, 3)

	if err != nil {

	}

	//处理接口返回数据
	if errJson := json.Unmarshal(data, &info); errJson != nil {

	}

	if info.ErrorResponse.Code != 0 {
		//
	}else{

	}

	return common.Succ(info)
}
