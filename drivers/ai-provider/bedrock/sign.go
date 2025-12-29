package bedrock

import (
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"net/http"
	"strings"
	"time"
)

func signRequest(signer *v4.Signer, region string, uri string, headers http.Header, body string) (http.Header, error) {
	request, err := http.NewRequest(http.MethodPost, uri, nil)
	if err != nil {
		return nil, err
	}
	request.Header = headers.Clone()
	
	_, err = signer.Sign(request, strings.NewReader(body), "bedrock", region, time.Now())
	if err != nil {
		return nil, err
	}
	return request.Header, nil
	
}
