package upstream

import (
	"fmt"
	"testing"

	"github.com/eolinker/eosc"
)

func TestUpstream(t *testing.T) {
	fmt.Println(eosc.TypeNameOf((*IUpstream)(nil)))
}
