package base

/*
*错误码
 */
const (
	Error         = -1
	Success       = 0
	NotLoginError = 900
	MissingData   = 400001
	DataStatus    = 400003
	ParamIllegal  = 400004
	RedisError    = 400005
	ParamError    = 400006
)

var errorMsg = map[int]string{
	Error:         "系统异常",
	Success:       "操作成功",
	NotLoginError: "未登录",
	MissingData:   "数据缺失",
	DataStatus:    "数据参数不正确，请勿非法操作",
	ParamIllegal:  "参数传入不合法:[%s]",
	RedisError:    "redis连接操作失败",
	ParamError:    "缺失参数不能",
}

func ErrorMsg(code int) string {
	return errorMsg[code]
}
