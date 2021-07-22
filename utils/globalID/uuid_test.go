package globalID

import (
	"fmt"
	"testing"
)

func TestGenerateIDString(t *testing.T) {

	tests := []struct {
	}{
		{},
		{},
		{},
		{},
		{},
		{},
	}
	for i := range tests {
		t.Run(fmt.Sprintf("GenerateIDString-%d", i), func(t *testing.T) {
			t.Logf("GenerateID() = %0X", GenerateID())
			d := GenerateIDString()
			t.Logf("GenerateIDString() = %s[%d]", d, len(d))
		})
	}
}
