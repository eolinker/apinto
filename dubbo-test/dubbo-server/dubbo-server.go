package dubbo_server

import (
	"bytes"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	"fmt"
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
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
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
	//fmt.Println(m)
	//fmt.Println(m["attachments"])
	fmt.Println(dubboPackage.Body)
	fmt.Println(dubboPackage.Codec)

}
