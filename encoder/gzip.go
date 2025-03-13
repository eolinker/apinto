package encoder

import (
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasthttp"
)

func init() {
	encoderManger.Set("gzip", &Gzip{})
}

type Gzip struct {
}

func (g *Gzip) ToUTF8(data []byte) ([]byte, error) {
	var bb bytebufferpool.ByteBuffer
	_, err := fasthttp.WriteGunzip(&bb, data)
	if err != nil {
		return nil, err
	}
	return bb.B, nil
	// 创建一个gzip reader
	//reader, err := gzip.NewReader(bytes.NewReader(data))
	//if err != nil {
	//	return nil, err
	//}
	//defer reader.Close()
	//
	//// 读取解压后的数据
	//var buf bytes.Buffer
	//_, err = io.Copy(&buf, reader)
	//if err != nil {
	//	return nil, err
	//}
	//
	//// 返回解压后的数据
	//// 注意：这里假设解压后的数据已经是UTF-8编码
	//// 如果需要处理其他编码转UTF-8，需要额外的转换步骤
	//return buf.Bytes(), nil
}
