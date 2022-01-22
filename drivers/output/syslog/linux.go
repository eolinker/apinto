//go:build !windows && !plan9
// +build !windows,!plan9

package syslog

//CreateTransporter 创建syslog-Transporter
func CreateTransporter(conf *SysConfig) (*SysWriter, error) {
	sysWriter, err := newSysWriter(conf, "")
	if err != nil {
		return nil, err
	}
	return &SysWriter{
		writer: sysWriter,
	}, nil
}
