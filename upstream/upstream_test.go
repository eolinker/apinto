package upstream

import (
	"fmt"
	"testing"

	"github.com/eolinker/goku-eosc/auth"

	"github.com/eolinker/eosc"
)

func TestUpstream(t *testing.T) {
	fmt.Println(eosc.TypeNameOf((*auth.IAuth)(nil)))
}
