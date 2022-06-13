package jwt

import (
	"io"
	"net/http"
	"testing"

	"github.com/valyala/fasthttp"

	http_context "github.com/eolinker/apinto/node/http-context"
)

type responseWriter struct{}

func (w responseWriter) Header() http.Header {
	return http.Header(map[string][]string{})
}
func (w responseWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (w responseWriter) WriteHeader(statusCode int) {
}

type requestBody struct{}

func (r requestBody) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

type JWTTest struct {
	testName string
	token    string
	want     string
}

var JWTConf = &Config{
	SignatureIsBase64: false,
	ClaimsToVerify:    []string{"exp", "nbf"},
	RunOnPreflight:    false,
	HideCredentials:   true,
	Credentials: []JwtCredential{
		{Iss: "TestHS256", Secret: "eolinker", RSAPublicKey: "", Algorithm: "HS256"},
		{Iss: "TestHS384", Secret: "eolinker", RSAPublicKey: "", Algorithm: "HS384"},
		{Iss: "TestHS512", Secret: "eolinker", RSAPublicKey: "", Algorithm: "HS512"},
		{Iss: "TestRS256", Secret: "eolinker", RSAPublicKey: "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnzyis1ZjfNB0bBgKFMSv\nvkTtwlvBsaJq7S5wA+kzeVOVpVWwkWdVha4s38XM/pa/yr47av7+z3VTmvDRyAHc\naT92whREFpLv9cj5lTeJSibyr/Mrm/YtjCZVWgaOYIhwrXwKLqPr/11inWsAkfIy\ntvHWTxZYEcXLgAXFuUuaS3uF9gEiNQwzGTU1v0FqkqTBr4B8nW3HCN47XUu0t8Y0\ne+lf4s4OxQawWD79J9/5d3Ry0vbV3Am1FtGJiJvOwRsIfVChDpYStTcHTCMqtvWb\nV6L11BWkpzGXSW4Hv43qa+GSYOD2QU68Mb59oSk2OB+BtOLpJofmbGEGgvmwyCI9\nMwIDAQAB\n-----END PUBLIC KEY-----", Algorithm: "RS256"},
		{Iss: "TestRS384", Secret: "eolinker", RSAPublicKey: "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnzyis1ZjfNB0bBgKFMSv\nvkTtwlvBsaJq7S5wA+kzeVOVpVWwkWdVha4s38XM/pa/yr47av7+z3VTmvDRyAHc\naT92whREFpLv9cj5lTeJSibyr/Mrm/YtjCZVWgaOYIhwrXwKLqPr/11inWsAkfIy\ntvHWTxZYEcXLgAXFuUuaS3uF9gEiNQwzGTU1v0FqkqTBr4B8nW3HCN47XUu0t8Y0\ne+lf4s4OxQawWD79J9/5d3Ry0vbV3Am1FtGJiJvOwRsIfVChDpYStTcHTCMqtvWb\nV6L11BWkpzGXSW4Hv43qa+GSYOD2QU68Mb59oSk2OB+BtOLpJofmbGEGgvmwyCI9\nMwIDAQAB\n-----END PUBLIC KEY-----", Algorithm: "RS384"},
		{Iss: "TestRS512", Secret: "eolinker", RSAPublicKey: "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnzyis1ZjfNB0bBgKFMSv\nvkTtwlvBsaJq7S5wA+kzeVOVpVWwkWdVha4s38XM/pa/yr47av7+z3VTmvDRyAHc\naT92whREFpLv9cj5lTeJSibyr/Mrm/YtjCZVWgaOYIhwrXwKLqPr/11inWsAkfIy\ntvHWTxZYEcXLgAXFuUuaS3uF9gEiNQwzGTU1v0FqkqTBr4B8nW3HCN47XUu0t8Y0\ne+lf4s4OxQawWD79J9/5d3Ry0vbV3Am1FtGJiJvOwRsIfVChDpYStTcHTCMqtvWb\nV6L11BWkpzGXSW4Hv43qa+GSYOD2QU68Mb59oSk2OB+BtOLpJofmbGEGgvmwyCI9\nMwIDAQAB\n-----END PUBLIC KEY-----", Algorithm: "RS512"},
		{Iss: "TestES256", Secret: "eolinker", RSAPublicKey: "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEEVs/o5+uQbTjL3chynL4wXgUg2R9\nq9UU8I5mEovUf86QZ7kOBIjJwqnzD1omageEHWwHdBO6B+dFabmdT9POxg==\n-----END PUBLIC KEY-----", Algorithm: "ES256"},
		{Iss: "TestES384", Secret: "eolinker", RSAPublicKey: "-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEC1uWSXj2czCDwMTLWV5BFmwxdM6PX9p+\nPk9Yf9rIf374m5XP1U8q79dBhLSIuaojsvOT39UUcPJROSD1FqYLued0rXiooIii\n1D3jaW6pmGVJFhodzC31cy5sfOYotrzF\n-----END PUBLIC KEY-----", Algorithm: "ES384"},
		{Iss: "TestES512", Secret: "eolinker", RSAPublicKey: "-----BEGIN PUBLIC KEY-----\nMIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQBgc4HZz+/fBbC7lmEww0AO3NK9wVZ\nPDZ0VEnsaUFLEYpTzb90nITtJUcPUbvOsdZIZ1Q8fnbquAYgxXL5UgHMoywAib47\n6MkyyYgPk0BXZq3mq4zImTRNuaU9slj9TVJ3ScT3L1bXwVuPJDzpr5GOFpaj+WwM\nAl8G7CqwoJOsW7Kddns=\n-----END PUBLIC KEY-----", Algorithm: "ES512"},
	},
}

var tests = []JWTTest{
	{
		testName: "测试HS256",
		token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJpc3MiOiJUZXN0SFMyNTYiLCJuYmYiOjE2MjcyODkxNzUsImV4cCI6MTY1ODgyNTE3NX0.mOianLD1sBOVhJ9UZyrcQZgrBBBu9keviDRWydT6BXk",
		want:     "nil",
	},
	{
		testName: "测试HS384",
		token:    "eyJhbGciOiJIUzM4NCIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJpc3MiOiJUZXN0SFMzODQiLCJuYmYiOjE2MjcyODkxNzUsImV4cCI6MTY1ODgyNTE3NX0.K5xuz653Vz3fu4m6BnUFUl6Me1H-fHDS5WxG4yyz8ZlgJnzvaSz9-rGnBdE0XYep",
		want:     "nil",
	},
	{
		testName: "测试HS512",
		token:    "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJpc3MiOiJUZXN0SFM1MTIiLCJuYmYiOjE2MjcyODkxNzUsImV4cCI6MTY1ODgyNTE3NX0.G7JqIKxEZOR6LzuUb_Epphgs9FINBokziQvdKg3CmkY8-QtGLk2YpMkCpsiRW18OTPYZq19iCjpYXQCYA1O3nw",
		want:     "nil",
	},
	{
		testName: "测试RS256",
		token:    "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiVGVzdFJTMjU2IiwibmJmIjoxNjI3Mjg5MTc1LCJleHAiOjE2NTg4MjUxNzV9.HdIDSuzWhWqgI1mx4P1udieu77n3vysBC1xDT75ppafJjv7k-F9ihusIjqPNwtj82qVNtmoIgqfOJ4YHrOqmE3cRk1G99l7Tx93WtcHqsFEo5SuFElpt0dk4Yq3haSh65sR_LdqxP4-H7FvOaPrJ5Jrkq3G1WBMGrzblUA2WCUzCZ9ANhKsQjV8WE16Bh6I7oI759KCAvhtdsnpu4KERpuYbcrffaVfuYfzkjhrawqRjAFt7fA4hcq1b9srEqmXxUc9IQ0AYmUWn9yyvNgdXGa-GrioT4Up7CYpmp5lbnjwgLo3cH7kDh5iyj-evVmOsyLvE2Ha9ZC-iV7wThyyE6g",
		want:     "nil",
	},
	{
		testName: "测试RS384",
		token:    "eyJhbGciOiJSUzM4NCIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiVGVzdFJTMzg0IiwibmJmIjoxNjI3Mjg5MTc1LCJleHAiOjE2NTg4MjUxNzV9.J15a16W4S-uLwWMJ9M1cpw055u82APrNX0XARf4od4bjJl3dLdmj797bYcdOMn13NLw3y6KAftY5iDrBjIBiirMg4X0bHCrc1FrenA9vfb7lzhWztiIdkrqv1zs-nV21JVgMhYkpwBF99BDdHNTuzLWyTob0s0z1VHcRkAFvEjttBmPnlLeX_NrSwho-hPolybycvFfBJaoZ1qjtmYwVYQMt-0TVnet_jxhbPCvSBzNl-6wM6cmD5PvSJXI_tvO8TtxHuV5KLh7p1IVgVVwTURGffzIo_KcdcofM_NnQkNIl6x3yuGNkB8pFAhMDDYVQjcW2zoTzdtLBqad-UQFfLA",
		want:     "nil",
	},
	{
		testName: "测试RS512",
		token:    "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiVGVzdFJTNTEyIiwibmJmIjoxNjI3Mjg5MTc1LCJleHAiOjE2NTg4MjUxNzV9.IpkdhYcUlAZb_nAPDI8sFCkb07TQ4wZ8e-HIZvlmzdKszl02mEaJXt9OgusVS5hZ3vohQ6abpxBEu0B89MzsJIbanH6Rz3oxmA-R837ac_6dEPHHOVTHbUawQ3qfa0YCb5HXQ2X4ldu-cmxlUeWi2ZTDOQU4bT7GoBneC8X-ae7c5zt71juIUlg0JxU53YLN93POqpWcNneJIGoC9VyX0NKm55eNp96yfXXWXaOhXxOXaZyeW4ACZ6GFu5H9BKkiG_tZjzA5Pppg5YxpsIpKrdEoTUnPpmoBCUWEsMc1T6KyoHAjIAXAgiNxLe3Q5Fgsf-789RSpzk1VeWetJkwGCA",
		want:     "nil",
	},
	{
		testName: "测试ES256",
		token:    "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiVGVzdEVTMjU2IiwibmJmIjoxNjI3Mjg5MTc1LCJleHAiOjE2NTg4MjUxNzV9.4XAE1wCY_5P7qGDOq04DfJWzgRJl28W1fN-Jh7a5sXj9ylYwRh5J35yoPz0lCfXAgt_hT_UjCAlB2h9OALKUGQ",
		want:     "nil",
	},
	{
		testName: "测试ES384",
		token:    "eyJhbGciOiJFUzM4NCIsInR5cCI6IkpXVCIsImtpZCI6ImlUcVhYSTB6YkFuSkNLRGFvYmZoa00xZi02ck1TcFRmeVpNUnBfMnRLSTgifQ.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiVGVzdEVTMzg0IiwibmJmIjoxNjI3Mjg5MTc1LCJleHAiOjE2NTg4MjUxNzV9.Mt0ZXHkDM50vx1nEjJs2Ok1HnE-64kFAp4j_xGNbL0q9wWq2u6n1AhbcDGa0v6GuqWuV1vrMStUR1t-5OhYJeQxyMHNjLqxyD6GJ8THODXUDIy81Nt1zN-qpjj6H7JeC",
		want:     "nil",
	},
	{
		testName: "测试ES512",
		token:    "eyJhbGciOiJFUzUxMiIsInR5cCI6IkpXVCIsImtpZCI6InhaRGZacHJ5NFA5dlpQWnlHMmZOQlJqLTdMejVvbVZkbTd0SG9DZ1NOZlkifQ.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiVGVzdEVTNTEyIiwibmJmIjoxNjI3Mjg5MTc1LCJleHAiOjE2NTg4MjUxNzV9.AZZdNIQ_bhhMCRcjoz4exemQ2yXvtRgKp27Oc3JbrGNSTOYbxBQpHJ79zfRjdGIM6Un6cgAeJgwTsYberuun-ccDACiiWgKbER3qVssz5bG4rb7eiEHhTaRF5b0gqaDdQE3i3QWdJMxp15-QrJbOQawr2DV4stUAjiJkEFzCej3asEvd",
		want:     "nil",
	},
	{
		testName: "测试HS512过期的token",
		token:    "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJpc3MiOiJUZXN0SFM1MTIiLCJuYmYiOjE2MjcyODkxNzUsImV4cCI6MTYyNzI4OTIzNX0.5oUtf-royNU8ckXgaDDVCv5dpVVdFyL7fOVz57LaMHXlzRyPQfLNGLMeaIWY6ZUpqRbzT_Jm3OZMXs5MVRmZlw",
		want:     "错误",
	},
	{
		testName: "测试HS512未到开始时间nbf的token",
		token:    "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJpc3MiOiJUZXN0SFM1MTIiLCJuYmYiOjE2MzcyODkxNzUsImV4cCI6MTYzNzI4OTIzNX0.mHTikcPyIHUAVBaHBHHTmpORcZQrZOqFWJPBR9by4mvVpJ73_WNchd093_yDs61Eh0LHs44ww047k-S4s5KrVA",
		want:     "错误",
	},
	{
		testName: "测试HS256ISS不存在的token",
		token:    "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJpc3MiOiJUZXN0SFN4eHgiLCJuYmYiOjE2MjcyODkyMzUsImV4cCI6MTY1ODgyNTIzNX0.42_GWGNLlHG1K-bCrB-NI0fNLEa9F6LBcX-pfdQWUn1RE0nB0SewDI_PyA62NtnMOc_R5rMBHcwdW5WUCxiPVw",
		want:     "错误",
	},
}

func TestJWT(t *testing.T) {
	jwtMoudule := &jwt{}
	err := jwtMoudule.Reset(JWTConf, nil)
	if err != nil {
		t.Errorf("配置读取出错 err:%s", err)
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// http-service
			//自己造formData, query或者body内插入jwt_token
			//httpRequest, _ := http-service.NewRequest("GET", "/asd/asd/asd?jwt_token="+test.token, requestBody{})
			//httpRequest.RequestHeader.SetDriver("Content-Type", "multipart/form-data")
			//ctx := http_context.NewContext(httpRequest, responseWriter{})
			//ctx.RequestOrg.SetHeader("Authorization-Type", "Jwt")

			// fasthttp
			context := &fasthttp.RequestCtx{
				Request:  *fasthttp.AcquireRequest(),
				Response: *fasthttp.AcquireResponse(),
			}
			context.Request.Header.SetMethod(fasthttp.MethodGet)
			context.Request.Header.Set("Content-Type", "multipart/form-data")
			context.Request.SetRequestURI("/asd/asd/asd?jwt_token=" + test.token)
			ctx := http_context.NewContext(context)

			err = jwtMoudule.Auth(ctx)

			resultErr := ""
			if err != nil {
				resultErr = "错误"
			} else {
				resultErr = "nil"
			}

			if resultErr != test.want {
				t.Errorf("test %s Fail", test.testName)
			}

		})
	}

}
