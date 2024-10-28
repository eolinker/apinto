package hunyuan

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/utils"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

func hmacsha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

func Sign(ctx http_service.IHttpContext, secretId string, secretKey string) error {
	payload, err := ctx.Proxy().Body().RawBody()
	if err != nil {
		return err
	}

	host := "hunyuan.tencentcloudapi.com"
	algorithm := "TC3-HMAC-SHA256"
	service := "hunyuan"
	version := "2023-09-01"
	action := "ChatCompletions"
	timestamp := time.Now().Unix()

	// step 1: build canonical request string
	httpRequestMethod := "POST"
	canonicalURI := "/"
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-tc-action:%s\n",
		"application/json", host, strings.ToLower(action))
	signedHeaders := "content-type;host;x-tc-action"
	hashedRequestPayload := utils.HexEncode(payload)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)
	log.Info("canonicalRequest", canonicalRequest)
	// step 2: build string to sign
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	hashedCanonicalRequest := utils.HexEncode([]byte(canonicalRequest))
	string2sign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithm,
		timestamp,
		credentialScope,
		hashedCanonicalRequest)
	log.Info("string2sign", string2sign)
	// step 3: sign string
	secretDate := hmacsha256(date, "TC3"+secretKey)
	secretService := hmacsha256(service, secretDate)
	secretSigning := hmacsha256("tc3_request", secretService)
	signature := hex.EncodeToString([]byte(hmacsha256(string2sign, secretSigning)))
	log.Info("signature", signature)
	// step 4: build authorization
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		secretId,
		credentialScope,
		signedHeaders,
		signature)
	log.Info("authorization", authorization)
	ctx.Proxy().Header().SetHeader("Authorization", authorization)
	ctx.Proxy().Header().SetHeader("Host", host)
	ctx.Proxy().Header().SetHeader("X-TC-Action", action)
	ctx.Proxy().Header().SetHeader("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	ctx.Proxy().Header().SetHeader("X-TC-Version", version)
	ctx.Proxy().Body().SetRaw("application/json", payload)
	return nil

}
