package hunyuan

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/eolinker/apinto/utils"
)

func TestSign(t *testing.T) {
	payload := []byte("{\"Limit\": 1, \"Filters\": [{\"Values\": [\"\\u672a\\u547d\\u540d\"], \"Name\": \"instance-name\"}]}")
	host := "cvm.tencentcloudapi.com"
	algorithm := "TC3-HMAC-SHA256"
	service := "cvm"
	//version := "2023-09-01"
	action := "describeinstances"
	var timestamp int64 = 1551113065
	secretId := "AKIDz8krbsJ5yKBZQpn74WFkmLPx3*******"

	// step 1: build canonical request string
	httpRequestMethod := "POST"
	canonicalURI := "/"
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-tc-action:%s\n",
		"application/json; charset=utf-8", host, strings.ToLower(action))
	signedHeaders := "content-type;host;x-tc-action"
	hashedRequestPayload := utils.HexEncode(payload)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)
	t.Log("canonicalRequest", canonicalRequest)
	// step 2: build string to sign
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	hashedCanonicalRequest := utils.HexEncode([]byte(canonicalRequest))
	string2sign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithm,
		timestamp,
		credentialScope,
		hashedCanonicalRequest)
	//t.Log("hashedCanonicalRequest", utils.HexEncode([]byte("POST\n/\n\ncontent-type:application/json; charset=utf-8\nhost:cvm.tencentcloudapi.com\nx-tc-action:describeinstances\n\ncontent-type;host;x-tc-action\n35e9c5b0e3ae67532d3c9f17ead6c90222632e5b1ff7f6e89887f1398934f064")))
	t.Log("string2sign", string2sign)
	secretKey := "Gu5t9xGARNpq86cd98joQYCN3*******"
	// step 3: sign string
	secretDate := utils.HMacBySha256("TC3"+secretKey, date)
	secretService := utils.HMacBySha256(secretDate, service)
	secretSigning := utils.HMacBySha256(secretService, "tc3_request")
	signature := hex.EncodeToString([]byte(utils.HMacBySha256(secretSigning, string2sign)))
	t.Log("signature", signature)
	// step 4: build authorization
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		secretId,
		credentialScope,
		signedHeaders,
		signature)
	t.Log("authorization", authorization)
}

func TestSign2(t *testing.T) {
	payload := "{\"Limit\": 1, \"Filters\": [{\"Values\": [\"\\u672a\\u547d\\u540d\"], \"Name\": \"instance-name\"}]}"
	host := "cvm.tencentcloudapi.com"
	algorithm := "TC3-HMAC-SHA256"
	service := "cvm"
	//version := "2023-09-01"
	action := "describeinstances"
	var timestamp int64 = 1551113065
	secretId := "AKIDz8krbsJ5yKBZQpn74WFkmLPx3*******"

	// step 1: build canonical request string
	httpRequestMethod := "POST"
	canonicalURI := "/"
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-tc-action:%s\n",
		"application/json; charset=utf-8", host, strings.ToLower(action))
	signedHeaders := "content-type;host;x-tc-action"
	hashedRequestPayload := sha256hex(payload)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)
	t.Log("canonicalRequest", canonicalRequest)
	// step 2: build string to sign
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	hashedCanonicalRequest := sha256hex(canonicalRequest)
	string2sign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithm,
		timestamp,
		credentialScope,
		hashedCanonicalRequest)
	//t.Log("hashedCanonicalRequest", utils.HexEncode([]byte("POST\n/\n\ncontent-type:application/json; charset=utf-8\nhost:cvm.tencentcloudapi.com\nx-tc-action:describeinstances\n\ncontent-type;host;x-tc-action\n35e9c5b0e3ae67532d3c9f17ead6c90222632e5b1ff7f6e89887f1398934f064")))
	t.Log("string2sign", string2sign)
	secretKey := "Gu5t9xGARNpq86cd98joQYCN3*******"
	// step 3: sign string
	secretDate := hmacsha256(date, "TC3"+secretKey)
	secretService := hmacsha256(service, secretDate)
	secretSigning := hmacsha256("tc3_request", secretService)
	signature := hex.EncodeToString([]byte(hmacsha256(string2sign, secretSigning)))
	t.Log("signature", signature)
	// step 4: build authorization
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		secretId,
		credentialScope,
		signedHeaders,
		signature)
	t.Log("authorization", authorization)
}
