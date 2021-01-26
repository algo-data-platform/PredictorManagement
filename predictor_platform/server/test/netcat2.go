// 主要模仿netcat来测试端口数据

package test

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
)

var host_port = flag.String("host_port", "127.0.0.1:10025", "default host")

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}

func main() {
	conn, err := net.Dial("tcp", *host_port)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	go mustCopy(os.Stdout, conn)
	mustCopy(conn, os.Stdin)
}
