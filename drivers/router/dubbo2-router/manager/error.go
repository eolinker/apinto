package manager

import (
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
)

func Dubbo2ErrorResult(err error) protocol.RPCResult {
	payload := impl.NewResponsePayload(nil, err, nil)
	return protocol.RPCResult{
		Attrs: payload.Attachments,
		Err:   payload.Exception,
		Rest:  payload.RspObj,
	}
}
