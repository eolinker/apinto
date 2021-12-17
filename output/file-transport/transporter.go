package file_transport

import "github.com/eolinker/eosc/formatter"

//Transporter filelog-Transporter结构
type Transporter struct {
	writer *FileWriterByPeriod
}

func (t *Transporter) Write(bytes []byte) error {
	_, err := t.writer.Write(bytes)
	return err
}

//Close 关闭
func (t *Transporter) Close() error {
	t.writer.Close()
	return nil
}

//NewtTransporter 创建file-Transporter
func NewtTransporter(cfg *Config) formatter.ITransport {

	fileWriterByPeriod := NewFileWriteByPeriod(cfg)

	return &Transporter{
		writer: fileWriterByPeriod,
	}

}
