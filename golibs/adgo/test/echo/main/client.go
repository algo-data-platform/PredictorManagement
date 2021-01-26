package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/algo-data-platform/predictor/golibs/adgo/service_router"
	"github.com/algo-data-platform/predictor/golibs/adgo/test/echo/if/echo"
	"github.com/algo-data-platform/predictor/golibs/adgo/thirdparty/thrift"
	"github.com/hashicorp/consul/api"
)

func main() {
	//var times int = 1000000
	//t1 := time.Now()

	clientOption := service_router.ClientOptionFactory("liubang_test", service_router.ServerProtocol_THRIFT)
	cg := service_router.GetConnGroup()
	router := service_router.GetRouter(&service_router.RouterConfig{
		Consul: api.Config{
			Address: "10.85.101.119:8500",
			Scheme:  "http",
		},
		ProjectName: "liubang_test",
		LocalIp:     "10.235.33.20",
	})

	var metrics = router.GetMetrics()

	wg := sync.WaitGroup{}
	wg.Add(2)
	//for i := 0; i < times; i++ {
	go func() {
		for {
			done := false
			server, ok := router.Discover(clientOption)
			if ok {
				config := service_router.ThriftConfig{
					Host:          server.Host,
					Port:          int(server.Port),
					TransportType: service_router.THRIFT_TRANSPORT_HEADER,
					NewThriftClient: func(trans thrift.Transport, proto thrift.ProtocolFactory) service_router.ThriftClient {
						return echo.NewEchoServiceClientFactory(trans, proto)
					},
					CloseThriftClient: func(conn service_router.ThriftClient) error {
						return conn.(*echo.EchoServiceClient).Transport.Close()
					},
					ThriftIsOpen: func(conn service_router.ThriftClient) bool {
						return conn.(*echo.EchoServiceClient).Transport.IsOpen()
					},
				}
				thriftConn, idx, err := cg.GetConnection(config, service_router.PoolMinConn(5), service_router.PoolMaxConn(5))
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
				thriftConn.Do(func(conn service_router.ThriftClient) (interface{}, error) {
					client := conn.(*echo.EchoServiceClient)
					request := echo.NewEchoRequest()
					request.Message = "hello world"
					return client.Echo(request)
				}).Done(func() {
					cg.Release(idx, thriftConn)
				}).Tryresponse(func(resp interface{}) {
					msg := resp.(*echo.EchoResponse).GetMessage()
					fmt.Println(msg)
				}, func(err error) {
					done = true
					fmt.Println(err)
				})
			} else {
				fmt.Printf("no %s server available\n", clientOption.ServiceName)
			}
			metrics.Tagged(map[string]string{"addrs": fmt.Sprintf("%s:%d", server.Host, server.Port)}).Meter("test_echo").Mark(1)
			if done {
				break
			}
		}
		wg.Done()
	}()

	//elapsed := time.Since(t1)
	//fmt.Println(elapsed)

	//t2 := time.Now()

	//for i := 0; i < times; i++ {
	go func() {
		var (
			trans thrift.Transport
			err   error
		)

		trans, err = thrift.NewSocket(thrift.SocketAddr("127.0.0.1:8888"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "error resolving address:", err)
		}
		if err = trans.Open(); err != nil {
			fmt.Fprintln(os.Stderr, "Error opening socket to ", "127.0.0.1", ":", 8888, " ", err)

		}

		trans = thrift.NewHeaderTransport(trans)
		pf := thrift.NewHeaderProtocolFactory()
		client := echo.NewEchoServiceClientFactory(trans, pf)

		for {
			request := echo.NewEchoRequest()
			request.Message = "hello world"
			response, err := client.Echo(request)
			_ = response
			if err != nil {
				fmt.Fprintln(os.Stderr, "[R2] - Request error ", err)
				break
			} else {
				// fmt.Println(response.GetMessage())
			}
			metrics.Tagged(map[string]string{"addr": "127.0.0.1:8888"}).Meter("test_echo_directly").Mark(1)
			// time.Sleep(time.Second)
		}
		trans.Close()
		wg.Done()
	}()

	wg.Wait()
	//elapsed = time.Since(t2)
	//fmt.Println(elapsed)
}
