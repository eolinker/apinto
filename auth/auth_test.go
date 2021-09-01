package auth

import (
	"fmt"
	"github.com/eolinker/eosc"
	"testing"
)

func TestAuth(t *testing.T) {
	fmt.Println(eosc.TypeNameOf((*IAuth)(nil)))
}