package extra_params_v2

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	positionQuery  = "query"
	positionHeader = "header"
	positionBody   = "body"

	paramConvert string = "convert"
	paramError   string = "error"
	paramOrigin  string = "origin"

	clientErrStatusCode = 400
	successStatusCode   = 200
)

var (
	paramPositionErrInfo = `[plugin extra-params config err] param position must be in the set ["query","header",body]. err position: %s `
	paramNameErrInfo     = `[plugin extra-params config err] param name must be not null. `
)

func encodeErr(ent string, origin string, statusCode int) error {
	if ent == "json" {
		tmp := map[string]interface{}{
			"message":     origin,
			"status_code": statusCode,
		}
		info, _ := json.Marshal(tmp)
		return fmt.Errorf("%s", info)
	}
	return errors.New(origin)
}
