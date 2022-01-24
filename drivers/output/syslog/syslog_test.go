//go:build !windows && !plan9
// +build !windows,!plan9

package syslog

import (
	"testing"
)

func TestPing(t *testing.T) {
	conf := &SysConfig{
		Network: "tcp",
		Address: "172.22.219.178:514",
		Level:   "info",
	}
	_, err := newSysWriter(conf, "test")
	if err != nil {
		t.Fatal(err)
		return
	}
}
