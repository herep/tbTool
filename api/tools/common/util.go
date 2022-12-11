package common

import (
	"tbTool/pkg"
	"time"
)

func ResErr(code int32, msg string) pkg.Render {
	return pkg.JSON{
		Data: &Response{
			Code:    code,
			Msg:     msg,
			NowTime: time.Now().Unix(),
			Data:    nil,
		},
	}
}

func Success(data map[string]interface{}) pkg.Render {
	return pkg.JSON{
		Data: &Response{
			Code:    0,
			Msg:     "success",
			NowTime: time.Now().Unix(),
			Data:    data,
		},
	}
}

func Succ(data interface{}) pkg.Render {
	return pkg.JSON{
		Data: &ResponseInterface{
			Code:    0,
			Msg:     "success",
			NowTime: time.Now().Unix(),
			Data:    data,
		},
	}
}
