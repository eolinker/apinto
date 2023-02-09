package dubbo_server

import (
	"bytes"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	"fmt"
	"net"
	"reflect"
)

func StartDubboServer() {
	listen, err := net.Listen("tcp", "127.0.0.1:43991")
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
		go handle(conn)
	}
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

	fmt.Println(reflect.TypeOf(dubboPackage.Body))

	typeList := make([]string, 0)
	attachments := make(map[string]interface{})
	name := ""
	if bodyMap, bOk := dubboPackage.Body.(map[string]interface{}); bOk {
		if attachmentsInteface, aOk := bodyMap["attachments"]; aOk {
			if attachmentsTemp, ok := attachmentsInteface.(map[string]interface{}); ok {
				attachments = attachmentsTemp
			}

		}

		if argsMap, aOk := bodyMap["args"]; aOk {
			fmt.Println(reflect.TypeOf(argsMap))
			if argsList, lOk := argsMap.([]interface{}); lOk {

				if len(argsList) > 0 {
					if argsStr, sOk := argsList[0].(string); sOk {
						name = argsStr
					}
				}
				if len(argsList) > 1 {
					if argsStr, sOk := argsList[1].([]string); sOk {
						typeList = argsStr
					}
				}

			}
		}
	}
	fmt.Println(name)
	fmt.Println(typeList)
	return
	fmt.Println(attachments)
	//fmt.Println(m["attachments"])
	fmt.Println(dubboPackage.Body)
	fmt.Println(dubboPackage.Codec)

}
