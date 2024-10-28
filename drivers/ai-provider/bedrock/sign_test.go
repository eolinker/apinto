package bedrock

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

const region = "us-east-1"
const model = "anthropic.claude-3-haiku-20240307-v1:0"

func TestSign(t *testing.T) {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	signer := v4.NewSigner(credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""))
	body := `{
    "messages": [
        {
            "role": "user",
            "content": [
                {
                    "text": "如何预防痛风？"
                }
            ]
        }
    ],
    "system": [
        {
            "text": "你是一个医生，需要针对用户的问题提供专业性的意见"
        }
    ]
}`
	bodyReader := strings.NewReader(body)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/converse", region, model), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	_, err = signer.Sign(request, bodyReader, "bedrock", region, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

func TestSignRequest(t *testing.T) {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	signer := v4.NewSigner(credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""))
	body := `{
    "messages": [
        {
            "role": "user",
            "content": [
                {
                    "text": "如何预防痛风？"
                }
            ]
        }
    ],
    "system": [
        {
            "text": "你是一个医生，需要针对用户的问题提供专业性的意见"
        }
    ]
}`
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/converse", region, model), strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("content-type", "application/json")
	headers, err := signRequest(signer, region, model, request.Header, body)
	if err != nil {
		t.Fatal(err)
	}
	request.Header = headers
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}
