package api

type Resp struct {
	ErrNo  Code        `json:"err_no"`
	ErrMsg string      `json:"err_msg"`
	Data   interface{} `json:"data"`
}

func RespOK(data interface{}) Resp {
	return Resp{
		Data: data,
	}
}

func RespErr(errNo Code, errMsg string) Resp {
	return Resp{
		ErrNo:  errNo,
		ErrMsg: errMsg,
	}
}

func (a *Resp) ApiRespErr(errNo Code, errMsg string) {
	a.ErrNo = errNo
	a.ErrMsg = errMsg
}

func (a *Resp) ApiRespOK(data interface{}) {
	a.ErrNo = 0
	a.Data = data
}
