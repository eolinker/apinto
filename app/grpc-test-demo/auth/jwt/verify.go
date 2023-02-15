package jwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/ohler55/ojg/jp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type Conf struct {
	KeyClaimName      string   `json:"keyClaimName"`
	SignatureIsBase64 bool     `json:"signatureIsBase64"`
	ClaimsToVerify    []string `json:"claimsToVerify"`
	ISS               string   `json:"iss"`
	Secret            string   `json:"secret"`
	RsaPublicKey      string   `json:"rsaPublicKey"`
	Algorithm         string   `json:"algorithm"`
	User              string   `json:"user"`
}

type Token struct {
	Token        string
	Header_64    string
	Claims_64    string
	Signature_64 string
	Header       map[string]interface{}
	Claims       map[string]interface{}
	Signature    string
}

type signingMethod struct {
	Name      string
	Hash      crypto.Hash
	KeySize   int
	CurveBits int
}

var (
	ErrInvalidKey           = errors.New("key is invalid")
	ErrInvalidKeyType       = errors.New("key is of invalid type")
	ErrHashUnavailable      = errors.New("the requested hash function is unavailable")
	ErrSignatureInvalid     = errors.New("signature is invalid")
	ErrInvalidSigningMethod = errors.New("signing method is invalid")
	ErrECDSAVerification    = errors.New("crypto/ecdsa: verification error")
	ErrKeyMustBePEMEncoded  = errors.New("Invalid Key: Key must be PEM encoded PKCS1 or PKCS8 private key")
	ErrNotRSAPublicKey      = errors.New("Key is not a valid RSA public key")
	ErrNotECPublicKey       = errors.New("Key is not a valid ECDSA public key")
)

func newSigningMethod(name string) *signingMethod {
	switch name {
	case "HS256":
		return &signingMethod{Name: name, Hash: crypto.SHA256}
	case "HS384":
		return &signingMethod{Name: name, Hash: crypto.SHA384}
	case "HS512":
		return &signingMethod{Name: name, Hash: crypto.SHA512}
	case "RS256":
		return &signingMethod{Name: name, Hash: crypto.SHA256}
	case "RS384":
		return &signingMethod{Name: name, Hash: crypto.SHA384}
	case "RS512":
		return &signingMethod{Name: name, Hash: crypto.SHA512}
	case "ES256":
		return &signingMethod{Name: name, Hash: crypto.SHA256, KeySize: 32, CurveBits: 256}
	case "ES384":
		return &signingMethod{Name: name, Hash: crypto.SHA384, KeySize: 48, CurveBits: 384}
	case "ES512":
		return &signingMethod{Name: name, Hash: crypto.SHA512, KeySize: 66, CurveBits: 512}
	default:
		return nil
	}
}

func (m *signingMethod) Verify(signingString, signature string, key interface{}) error {
	switch m.Name {
	case "HS256", "HS384", "HS512":
		{
			// Verify the key is the right type
			keyBytes, ok := key.([]byte)
			if !ok {
				return ErrInvalidKeyType
			}

			// Decode signature, for comparison
			sig, err := decodeSegment(signature)
			if err != nil {
				return err
			}

			// Can we use the specified hashing method?
			if !m.Hash.Available() {
				return ErrHashUnavailable
			}

			// This signing method is symmetric, so we validate the signature
			// by reproducing the signature from the signing string and key, then
			// comparing that against the provided signature.
			hasher := hmac.New(m.Hash.New, keyBytes)
			hasher.Write([]byte(signingString))
			if !hmac.Equal(sig, hasher.Sum(nil)) {
				return ErrSignatureInvalid
			}

			// No validation errors.  Signature is good.
			return nil
		}
	case "RS256", "RS384", "RS512":
		{
			var err error

			// Decode the signature
			var sig []byte
			if sig, err = decodeSegment(signature); err != nil {
				return err
			}

			var rsaKey *rsa.PublicKey
			var ok bool

			if rsaKey, ok = key.(*rsa.PublicKey); !ok {
				return ErrInvalidKeyType
			}

			// Create hasher
			if !m.Hash.Available() {
				return ErrHashUnavailable
			}
			hasher := m.Hash.New()
			hasher.Write([]byte(signingString))

			// Verify the signature
			return rsa.VerifyPKCS1v15(rsaKey, m.Hash, hasher.Sum(nil), sig)
		}
	case "ES256", "ES384", "ES512":
		{
			var err error

			// Decode the signature
			var sig []byte
			if sig, err = decodeSegment(signature); err != nil {
				return err
			}

			// Get the key
			var ecdsaKey *ecdsa.PublicKey
			switch k := key.(type) {
			case *ecdsa.PublicKey:
				ecdsaKey = k
			default:
				return ErrInvalidKeyType
			}

			if len(sig) != 2*m.KeySize {
				return ErrECDSAVerification
			}

			r := big.NewInt(0).SetBytes(sig[:m.KeySize])
			s := big.NewInt(0).SetBytes(sig[m.KeySize:])

			// Create hasher
			if !m.Hash.Available() {
				return ErrHashUnavailable
			}
			hasher := m.Hash.New()
			hasher.Write([]byte(signingString))

			// Verify the signature
			if verifystatus := ecdsa.Verify(ecdsaKey, hasher.Sum(nil), r, s); verifystatus == true {
				return nil
			} else {
				return ErrECDSAVerification
			}
		}
	default:
		{
			return ErrInvalidSigningMethod
		}
	}
}

func (m *signingMethod) Sign(signingString string, key interface{}) (string, error) {
	switch m.Name {
	case "HS256", "HS384", "HS512":
		{
			if keyBytes, ok := key.([]byte); ok {
				if !m.Hash.Available() {
					return "", ErrHashUnavailable
				}

				hasher := hmac.New(m.Hash.New, keyBytes)
				hasher.Write([]byte(signingString))

				return encodeSegment(hasher.Sum(nil)), nil
			}

			return "", ErrInvalidKeyType
		}
	case "RS256", "RS384", "RS512":
		{
			var rsaKey *rsa.PrivateKey
			var ok bool

			// Validate type of key
			if rsaKey, ok = key.(*rsa.PrivateKey); !ok {
				return "", ErrInvalidKey
			}

			// Create the hasher
			if !m.Hash.Available() {
				return "", ErrHashUnavailable
			}

			hasher := m.Hash.New()
			hasher.Write([]byte(signingString))

			// Sign the string and return the encoded bytes
			if sigBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, m.Hash, hasher.Sum(nil)); err == nil {
				return encodeSegment(sigBytes), nil
			} else {
				return "", err
			}
		}
	case "ES256", "ES384", "ES512":
		{
			// Get the key
			var ecdsaKey *ecdsa.PrivateKey
			switch k := key.(type) {
			case *ecdsa.PrivateKey:
				ecdsaKey = k
			default:
				return "", ErrInvalidKeyType
			}

			// Create the hasher
			if !m.Hash.Available() {
				return "", ErrHashUnavailable
			}

			hasher := m.Hash.New()
			hasher.Write([]byte(signingString))

			// Sign the string and return r, s
			if r, s, err := ecdsa.Sign(rand.Reader, ecdsaKey, hasher.Sum(nil)); err == nil {
				curveBits := ecdsaKey.Curve.Params().BitSize

				if m.CurveBits != curveBits {
					return "", ErrInvalidKey
				}

				keyBytes := curveBits / 8
				if curveBits%8 > 0 {
					keyBytes += 1
				}

				// We serialize the outpus (r and s) into big-endian byte arrays and pad
				// them with zeros on the left to make sure the sizes work out. Both arrays
				// must be keyBytes long, and the output must be 2*keyBytes long.
				rBytes := r.Bytes()
				rBytesPadded := make([]byte, keyBytes)
				copy(rBytesPadded[keyBytes-len(rBytes):], rBytes)

				sBytes := s.Bytes()
				sBytesPadded := make([]byte, keyBytes)
				copy(sBytesPadded[keyBytes-len(sBytes):], sBytes)

				out := append(rBytesPadded, sBytesPadded...)

				return encodeSegment(out), nil
			} else {
				return "", err
			}
		}
	default:
		{
			return "", ErrInvalidSigningMethod
		}
	}
}

func methodEnable(method string) bool {
	if method == "HS256" || method == "HS384" || method == "HS512" || method == "RS256" || method == "RS384" || method == "RS512" || method == "ES256" || method == "ES384" || method == "ES512" {
		return true
	}
	return false
}

// Decode JWT specific base64url encoding with padding stripped
func decodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	return base64.URLEncoding.DecodeString(seg)
}

// Encode JWT specific base64url encoding with padding stripped
func encodeSegment(seg []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(seg), "=")
}

// Parse PEM encoded PKCS1 or PKCS8 public key
func ParseRSAPublicKeyFromPEM(key []byte) (*rsa.PublicKey, error) {
	var err error

	// Parse PEM block
	var block *pem.Block
	if block, _ = pem.Decode(key); block == nil {
		return nil, ErrKeyMustBePEMEncoded
	}

	// Parse the key
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKIXPublicKey(block.Bytes); err != nil {
		if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
			parsedKey = cert.PublicKey
		} else {
			return nil, err
		}
	}

	var pkey *rsa.PublicKey
	var ok bool
	if pkey, ok = parsedKey.(*rsa.PublicKey); !ok {
		return nil, ErrNotRSAPublicKey
	}

	return pkey, nil
}

// Parse PEM encoded PKCS1 or PKCS8 public key
func ParseECPublicKeyFromPEM(key []byte) (*ecdsa.PublicKey, error) {
	var err error

	// Parse PEM block
	var block *pem.Block
	if block, _ = pem.Decode(key); block == nil {
		return nil, ErrKeyMustBePEMEncoded
	}

	// Parse the key
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKIXPublicKey(block.Bytes); err != nil {
		if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
			parsedKey = cert.PublicKey
		} else {
			return nil, err
		}
	}

	var pkey *ecdsa.PublicKey
	var ok bool
	if pkey, ok = parsedKey.(*ecdsa.PublicKey); !ok {
		return nil, ErrNotECPublicKey
	}

	return pkey, nil
}

func b64Encode(input string) string {
	result := base64.StdEncoding.EncodeToString([]byte(input))
	result = strings.Replace(strings.Replace(strings.Replace(result, "=", "", -1), "/", "_", -1), "+", "-", -1)
	return result
}

// base64解密
func b64Decode(input string) (string, error) {
	remainder := len(input) % 4
	// base64编码需要为4的倍数，如果不是4的倍数，则填充"="号
	if remainder > 0 {
		padlen := 4 - remainder
		input = input + strings.Repeat("=", padlen)
	}
	// 将原字符串中的"_","-"分别用"/"和"+"替换
	input = strings.Replace(strings.Replace(input, "_", "/", -1), "-", "+", -1)
	result, err := base64.StdEncoding.DecodeString(input)
	return string(result), err
}

// 根据"."分割token字符串
func tokenize(token string) []string {
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		return parts
	} else {
		return nil
	}
}

// 解析token，将token信息解析为jwtToken对象
func decodeToken(token string) (*Token, error) {
	tokenParts := tokenize(token)
	if tokenParts == nil {
		return nil, errors.New("[jwt_auth] Invalid token")
	}
	header_64 := tokenParts[0]
	claims_64 := tokenParts[1]
	signature_64 := tokenParts[2]
	var header, claims map[string]interface{}
	var signature string
	header_d64, err := b64Decode(header_64)
	if err != nil {
		return nil, errors.New("[jwt_auth] Invalid base64 encoded JSON")
	}

	if err = json.Unmarshal([]byte(header_d64), &header); err != nil {
		return nil, errors.New("[jwt_auth] Invalid JSON")
	}
	claims_d64, err := b64Decode(claims_64)
	if err != nil {
		return nil, errors.New("[jwt_auth] Invalid base64 encoded JSON")
	}
	if err = json.Unmarshal([]byte(claims_d64), &claims); err != nil {
		return nil, errors.New("[jwt_auth] Invalid JSON")
	}
	signature, err = b64Decode(signature_64)
	if err != nil {
		return nil, errors.New("[jwt_auth] Invalid base64 encoded JSON")
	}
	if _, ok := header["typ"]; !ok || strings.ToUpper(header["typ"].(string)) != "JWT" {
		return nil, errors.New("[jwt_auth] Invalid typ")
	}
	if _, ok := header["alg"]; !ok || !methodEnable(header["alg"].(string)) {
		return nil, errors.New("[jwt_auth] Invalid alg")
	}
	if len(claims) == 0 {
		return nil, errors.New("[jwt_auth] Invalid claims")
	}
	if len(signature) == 0 {
		return nil, errors.New("[jwt_auth] Invalid signature")
	}
	return &Token{Token: token, Header_64: header_64, Claims_64: claims_64, Signature_64: signature_64, Header: header, Claims: claims, Signature: signature}, nil
}

func verifySignature(token *Token, key string) error {

	var k interface{}
	switch token.Header["alg"].(string) {
	case "HS256", "HS384", "HS512":
		{
			k = []byte(key)
		}
	case "RS256", "RS384", "RS512":
		{
			var err error
			k, err = ParseRSAPublicKeyFromPEM([]byte(key))
			if err != nil {
				return err
			}
		}
	case "ES256", "ES384", "ES512":
		{
			var err error
			k, err = ParseECPublicKeyFromPEM([]byte(key))
			if err != nil {
				return err
			}
		}
	default:
		{
			return ErrInvalidSigningMethod
		}
	}
	return newSigningMethod(token.Header["alg"].(string)).Verify(token.Header_64+"."+token.Claims_64, token.Signature_64, k)
}

func verifyRegisteredClaims(token *Token, claimsToVerify []string) error {
	if claimsToVerify == nil {
		claimsToVerify = []string{}
	}
	var err error = nil
	for _, claimName := range claimsToVerify {
		var claim int64 = 0
		if _, ok := token.Claims[claimName]; ok {
			//if typeOfData(token.Claims[claimName]) == reflect.Int64 || typeOfData(token.Claims[claimName]) == reflect.Int {
			//	claim = token.Claims[claimName].(int64)
			//}

			if typeOfData(token.Claims[claimName]) == reflect.Float64 {
				claimFloat64, success := token.Claims[claimName].(float64)
				if success {
					claim = int64(claimFloat64)
				}
			}

		}
		if claim < 1 {
			return errors.New("[jwt_auth] " + claimName + " must be a number")
		}
		switch claimName {
		case "nbf":
			if claim > time.Now().Unix() {
				err = errors.New("[jwt_auth] token not valid yet")
			}
		case "exp":
			if claim <= time.Now().Unix() {
				err = errors.New("[jwt_auth] token expired")
			}
		default:
			err = errors.New("[jwt_auth] Invalid claims")
		}
	}
	return err
}

// 获取数据的类型
func typeOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

func DoJWTAuthentication(cs map[string]*Conf, md map[string][]string) (string, error) {
	tokenStr, ok := md[":authority"]
	if !ok {
		return "", grpc.Errorf(codes.Unauthenticated, "无Token认证信息")
	}
	token, err := decodeToken(tokenStr[0])
	if err != nil {
		return "", errors.New("[jwt_auth] Bad token; " + err.Error())
	}
	key := ""
	claimName := "iss"
	if v, ok := token.Claims[claimName]; ok {
		key = v.(string)
	}
	if key == "" {
		if v, ok := token.Header[claimName]; ok {
			key = v.(string)
		}
	}

	if key == "" {
		return "", errors.New("[jwt_auth] No mandatory " + claimName + " in claims")
	}
	conf, ok := cs[key]
	if !ok {
		return "", errors.New("[jwt_auth] No key " + claimName + " in claims")
	}

	jwtSecretValue := conf.RsaPublicKey
	algorithm := "HS256"
	if conf.Algorithm != "" {
		algorithm = conf.Algorithm
	}
	if algorithm == "HS256" || algorithm == "HS384" || algorithm == "HS512" {
		jwtSecretValue = conf.Secret
	}
	if conf.SignatureIsBase64 {
		jwtSecretValue, err = b64Decode(jwtSecretValue)
		if err != nil {
			return "", errors.New("[jwt_auth] Invalid key/secret")
		}
	}
	if jwtSecretValue == "" {
		return "", errors.New("[jwt_auth] Invalid key/secret")
	}
	if token.Header["alg"].(string) != algorithm {

		return "", errors.New("[jwt_auth] Invalid algorithm")
	}
	if err := verifySignature(token, jwtSecretValue); err != nil {

		return "", errors.New("[jwt_auth] Invalid signature")
	}
	if err := verifyRegisteredClaims(token, conf.ClaimsToVerify); err != nil {

		return "", err
	}

	return getUser(conf.User, token.Claims)
}

func getUser(path string, data map[string]interface{}) (string, error) {
	if !strings.HasPrefix(path, "$.") {
		path = fmt.Sprintf("$.%s", path)
	}
	x, err := jp.ParseString(path)
	if err != nil {
		return "", fmt.Errorf("fail to get user,error: %w", err)
	}
	value := x.Get(data)
	if len(value) < 1 {
		return "", fmt.Errorf("fail to get user,error,path error")
	}
	v, ok := value[0].(string)
	if !ok {
		return "", fmt.Errorf("fail to get user,error,type error")
	}
	return v, nil
}
