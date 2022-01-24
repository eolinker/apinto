//go:build !windows && !plan9
// +build !windows,!plan9

package syslog

import (
	"testing"
)

func TestPing(t *testing.T) {
	conf := &SysConfig{
		Network: "tcp",
		Address: "127.0.0.1:514",
		Level:   "info",
	}
	w, err := newSysWriter(conf, "test")
	if err != nil {
		t.Fatal(err)
		return
	}
	n, err := w.Write([]byte("test"))
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Logf("send %d;write %d", len([]byte("test")), n)
}
