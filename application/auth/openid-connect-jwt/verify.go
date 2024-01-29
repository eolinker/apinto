package openid_connect_jwt

import (
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/jws"

	"github.com/eolinker/eosc/log"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

type IVerifyClaim interface {
	Verify(obj interface{}) error
}

var claims = []IVerifyClaim{
	newNbfClaim(),
	newExpClaim(),
}

func newNbfClaim() IVerifyClaim {
	nbfExpr, _ := jp.ParseString("$.nbf")
	return &nbfClaim{nbfExpr}
}

type nbfClaim struct {
	expr jp.Expr
}

func (n *nbfClaim) Verify(obj interface{}) error {
	result := n.expr.Get(obj)
	if len(result) == 0 {
		return nil
	}
	var nbf int64
	switch r := result[0].(type) {
	case float64:
		nbf = int64(r)
	case int64:
		nbf = r
	default:
		return fmt.Errorf("nbf claim type error")
	}
	if nbf > time.Now().Unix() {
		return fmt.Errorf("token not valid yet")
	}
	return nil
}

func newExpClaim() IVerifyClaim {
	expExpr, _ := jp.ParseString("$.exp")
	return &expClaim{expExpr}
}

type expClaim struct {
	expr jp.Expr
}

func (e *expClaim) Verify(obj interface{}) error {
	result := e.expr.Get(obj)
	if len(result) == 0 {
		return nil
	}
	var exp int64
	switch r := result[0].(type) {
	case float64:
		exp = int64(r)
	case int64:
		exp = r
	default:
		return fmt.Errorf("exp claim type error")
	}
	if exp < time.Now().Unix() {
		return fmt.Errorf("token expired")
	}
	return nil
}

func verify(token string) (string, interface{}, bool) {
	id, payload, success := verifySign(token)
	if !success {
		return "", nil, false
	}
	obj, err := oj.Parse(payload)
	if err != nil {
		log.Errorf("%w, payload: %s", err, string(payload))
		return "", nil, false
	}
	for _, c := range claims {
		err = c.Verify(obj)
		if err != nil {
			log.Errorf("%w, payload: %s", err, string(payload))
			return "", nil, false
		}
	}
	return id, obj, true

}

func verifySign(token string) (string, []byte, bool) {
	header, err := extractTokenHeader(token)
	if err != nil {
		return "", nil, false
	}
	for _, issuer := range manager.Issuers.All() {
		if key, ok := issuer.JWKKeys[header.Kid]; ok {
			payload, err := jws.Verify([]byte(token), jwa.SignatureAlgorithm(key.Algorithm()), key)
			if err != nil {
				log.DebugF("%w, issuer: %s, key: %s", err, issuer.Issuer, key.KeyID())
				continue
			}
			log.DebugF("verify sign successful! payload: %s", string(payload))
			return issuer.ID, payload, true
		}
	}
	return "", nil, false
}
