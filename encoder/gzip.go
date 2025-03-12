package encoder

import (
	"bytes"
	"compress/gzip"
	"io"
)

func init() {
	//encoderManger.Set("gzip", &Gzip{})
}

type Gzip struct {
}

func (g *Gzip) ToUTF8(data []byte) ([]byte, error) {
	// 创建一个gzip reader
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// 读取解压后的数据
	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return nil, err
	}

	// 返回解压后的数据
	// 注意：这里假设解压后的数据已经是UTF-8编码
	// 如果需要处理其他编码转UTF-8，需要额外的转换步骤
	return buf.Bytes(), nil
}
