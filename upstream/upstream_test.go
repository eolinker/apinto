package upstream

import (
	"fmt"
	"testing"

	"github.com/eolinker/goku-eosc/discovery"

	"github.com/eolinker/eosc"
)

func TestUpstream(t *testing.T) {
	fmt.Println(eosc.TypeNameOf((*discovery.IDiscovery)(nil)))
}
