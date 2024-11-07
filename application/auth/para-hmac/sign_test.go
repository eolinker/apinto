package para_hmac

import (
	"net/url"
	"testing"
)

func TestSign(t *testing.T) {
	appId := "ed3DtlYh"
	appKey := "3YnxRNwULya0W56A"
	sequenceNo := "2024103117392905280623"
	timestamp := "20241031173929052"

	body := "{\"agentId\":\"nx\"}"
	t.Log(url.PathUnescape("qqY%2BkJ2LUany%2FSZf3a3Rd73scYsEXDGRoeF5FPuZJ2g%3D"))
	t.Log(sign(appId, appKey, timestamp, sequenceNo, body))
}
