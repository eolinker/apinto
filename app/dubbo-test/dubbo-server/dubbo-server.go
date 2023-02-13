package dubbo_server

import (
	"bytes"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	"encoding/json"
	"fmt"
	hessian "github.com/apache/dubbo-go-hessian2"
	gxbytes "github.com/dubbogo/gost/bytes"
	"github.com/pkg/errors"
	"io"
	"net"
)

func StartDubboServer() {
	listen, err := net.Listen("tcp", "127.0.0.1:4399")
	if err != nil {
		panic(err)
	}
	// 3. 关闭监听通道
	defer listen.Close()
	fmt.Println("server is Listening")
	for {
		// 2. 进行通道监听
		conn, err := listen.Accept()
		if err != nil {
			panic(err)
		}
		// 启动一个协程去单独处理该连接
		go testHandle(conn)
	}
}

func testHandle(conn net.Conn) {
	pktBuf := gxbytes.NewBuffer(nil)
	var (
		err      error
		ok       bool
		netError net.Error
		buf      []byte
	)

	bufLen := 0
	reader := io.Reader(conn)
	for {
		// for clause for the network timeout condition check
		// s.conn.SetReadTimeout(time.Now().Add(s.rTimeout))
		buf = pktBuf.WriteNextBegin(4 * 1024)
		bufLen, err = reader.Read(buf)
		if err != nil {
			if netError, ok = errors.Cause(err).(net.Error); ok && netError.Timeout() {
				break
			}
			if errors.Cause(err) == io.EOF {
				err = nil
				if bufLen != 0 {
					// as https://github.com/apache/dubbo-getty/issues/77#issuecomment-939652203
					// this branch is impossible. Even if it happens, the bufLen will be zero and the error
					// is io.EOF when getty continues to read the socket.
				}
				break
			}
		}
		break
	}
	dubboPackage := impl.NewDubboPackage(bytes.NewBuffer(buf))
	if err = dubboPackage.ReadHeader(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dubboPackage)
	if err = dubboPackage.Unmarshal(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dubboPackage.Header)
	fmt.Println(dubboPackage.Service)
}

func handle(conn net.Conn) {

	var info [128 * 1024]byte
	n, err := conn.Read(info[:])
	if err != nil {
		fmt.Println("conn Read fail ,err = ", err)
		return
	}
	buf := bytes.NewBuffer(info[:n])
	dubboPackage := impl.NewDubboPackage(buf)
	if err = dubboPackage.ReadHeader(); err != nil {
		fmt.Println(err)
		return
	}

	if err = dubboPackage.Unmarshal(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(dubboPackage.Header)
	fmt.Println(dubboPackage.Service)

	maps := make(map[string]interface{})
	maps["zhangzeyi"] = "123456"
	marshal, err := json.Marshal(maps)

	bytes1, err := packageMarshal(dubboPackage.Header.ID, dubboPackage.Header.SerialID, marshal)
	if err != nil {
		panic(err)
	}
	conn.Write(bytes1)
	//fmt.Println(reflect.TypeOf(dubboPackage.Body))
	//
	//unmarshal, s, strings, objects := packageUnmarshal(dubboPackage)
	//fmt.Println(unmarshal, s, strings, objects)
	return
	//fmt.Println(m["attachments"])

}

func packageMarshal(id int64, serialID byte, body interface{}) ([]byte, error) {
	dubboPackage := impl.NewDubboPackage(nil)
	dubboPackage.Header = impl.DubboHeader{
		SerialID:       serialID,
		Type:           impl.PackageResponse,
		ID:             id,
		ResponseStatus: impl.Response_OK,
	}

	dubboPackage.Header.Type = impl.PackageResponse
	dubboPackage.Header.ResponseStatus = impl.Response_OK

	dubboPackage.SetBody(impl.EnsureResponsePayload(body))
	buf, err := dubboPackage.Marshal()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func packageUnmarshal(dubboPackage *impl.DubboPackage) (map[string]interface{}, string, []string, []hessian.Object) {
	attachments := make(map[string]interface{})
	methodName := ""
	typeList := make([]string, 0)
	valueList := make([]hessian.Object, 0)
	if bodyMap, bOk := dubboPackage.Body.(map[string]interface{}); bOk {
		if attachmentsInteface, aOk := bodyMap["attachments"]; aOk {
			if attachmentsTemp, ok := attachmentsInteface.(map[string]interface{}); ok {
				attachments = attachmentsTemp
			}

		}

		if argsMap, aOk := bodyMap["args"]; aOk {
			if argsList, lOk := argsMap.([]interface{}); lOk {

				if len(argsList) > 0 {
					if argsStr, sOk := argsList[0].(string); sOk {
						methodName = argsStr
					}
				}
				if len(argsList) > 1 {
					if argsTypeList, sOk := argsList[1].([]string); sOk {
						typeList = argsTypeList
					}
				}
				if len(argsList) > 2 {
					if argsValueList, sOk := argsList[2].([]hessian.Object); sOk {
						valueList = argsValueList
					}
				}

			}
		}
	}

	return attachments, methodName, typeList, valueList
}
