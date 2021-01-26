package predictor_client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/algo-data-platform/predictor/golibs/ads_common_go/common/tally"
	"github.com/algo-data-platform/predictor/golibs/ads_common_go/service_router"
	"github.com/algo-data-platform/predictor/golibs/ads_common_go/thirdparty/thrift"
	"github.com/algo-data-platform/predictor/golibs/adgo/predictor_client/if/predictor"
	"github.com/hashicorp/consul/api"
)

const (
	PREDICTOR_CLIENT_MODULE_NAME     = "predictor_client"
	PREDICTOR_PREDIC_TIMER           = "predictor_predict_timer"
	PREDICTOR_PREDICT_ERROR           = "predictor_predict_error"
	PREDICTOR_CALCULATE_VECTOR_TIMER = "predictor_calculate_vector_timer"
	PREDICTOR_CALCULATE_VECTOR_METER = "predictor_calculate_vector_meter"
	PREDICTOR_CALCULATE_VECTOR_ERROR = "predictor_calculate_vector_error"
	PREDICTOR_CALCULATE_BATCH_VECTOR_TIMER = "predictor_calculate_batch_vector_timer"
	PREDICTOR_CALCULATE_BATCH_VECTOR_METER = "predictor_calculate_batch_vector_meter"
	PREDICTOR_CALCULATE_BATCH_VECTOR_ERROR = "predictor_calculate_batch_vector_error"
	PREDICTOR_CLIENT_TIMEOUT_METER   = "predictor_client_timeout_meter"
	RPC_TIMEOUT = "rpc timeout"
)

type PredictorClientConfig struct {
	ConsulAddress        string        `json:"consul_address"`
	PredictorServiceName string        `json:"predictor_service_name"`
	LocalIp              string        `json:"local_ip"`
	Timeout              time.Duration `json:"timeout"`
	MaxConnPerServer     int           `json:"max_conn_per_server"`
	WaitConnection       bool          `json:"wait_connection"`
}

const (
	DEFAULT_TIMEOUT             = 8 * time.Millisecond
	DEFAULT_MAX_CONN_PRE_SERVER = 32
)

func NewPredictorClientConfig(consulAddress string, predictorServiceName string, localIp string) *PredictorClientConfig {
	return &PredictorClientConfig{
		ConsulAddress:        consulAddress,
		PredictorServiceName: predictorServiceName,
		LocalIp:              localIp,
		Timeout:              DEFAULT_TIMEOUT,
		MaxConnPerServer:     DEFAULT_MAX_CONN_PRE_SERVER,
		WaitConnection:       false,
	}
}

type PredictorClient struct {
	config       *PredictorClientConfig
	router       *service_router.Router
	clientOption service_router.ClientOption
	connGroup    *service_router.ConnGroup
	meters       map[string]tally.Meter
	timers       map[string]tally.Timer
}

func NewPredictorClient(consulAddress string, predictorServiceName string, localIp string) (*PredictorClient, error) {
	client := new(PredictorClient)

	client.config = NewPredictorClientConfig(consulAddress, predictorServiceName, localIp)

	// init service router related stuff
	client.clientOption = service_router.ClientOptionFactory(client.config.PredictorServiceName, service_router.ServerProtocol_THRIFT)
    client.clientOption.LocalFirstConfig.LocalIp = localIp
	client.connGroup = service_router.GetConnGroup()
	client.router = service_router.GetRouter(&service_router.RouterConfig{
		Consul: api.Config{
			Address: client.config.ConsulAddress,
			Scheme:  "http",
		},
		ProjectName: "predictor_client",
		LocalIp:     client.config.LocalIp,
	})

	metric := client.router.GetMetrics()
	tag := map[string]string{"module": PREDICTOR_CLIENT_MODULE_NAME}
	client.timers = make(map[string]tally.Timer)
	client.timers[PREDICTOR_CALCULATE_VECTOR_TIMER] = metric.Tagged(tag).Timer(PREDICTOR_CALCULATE_VECTOR_TIMER)
	client.timers[PREDICTOR_CALCULATE_BATCH_VECTOR_TIMER] = metric.Tagged(tag).Timer(PREDICTOR_CALCULATE_BATCH_VECTOR_TIMER)
	client.timers[PREDICTOR_PREDIC_TIMER] = metric.Tagged(tag).Timer(PREDICTOR_PREDIC_TIMER)
	client.meters = make(map[string]tally.Meter)
	client.meters[PREDICTOR_CALCULATE_VECTOR_METER] = metric.Tagged(tag).Meter(PREDICTOR_CALCULATE_VECTOR_METER)
	client.meters[PREDICTOR_CALCULATE_BATCH_VECTOR_METER] = metric.Tagged(tag).Meter(PREDICTOR_CALCULATE_BATCH_VECTOR_METER)
	client.meters[PREDICTOR_CLIENT_TIMEOUT_METER] = metric.Tagged(tag).Meter(PREDICTOR_CLIENT_TIMEOUT_METER)
	client.meters[PREDICTOR_PREDICT_ERROR] = metric.Tagged(tag).Meter(PREDICTOR_PREDICT_ERROR)
	client.meters[PREDICTOR_CALCULATE_VECTOR_ERROR] = metric.Tagged(tag).Meter(PREDICTOR_CALCULATE_VECTOR_ERROR)
	client.meters[PREDICTOR_CALCULATE_BATCH_VECTOR_ERROR] = metric.Tagged(tag).Meter(PREDICTOR_CALCULATE_BATCH_VECTOR_ERROR)
    // no local first by default
    client.clientOption.LocalFirstConfig.DiffRange = 0;

	fmt.Println("localfirst configis ", client.clientOption.LocalFirstConfig)

	return client, nil
}

func newPredictorServiceClient(transport thrift.Transport, protocol thrift.ProtocolFactory) service_router.ThriftClient {
	return predictor.NewPredictorServiceClientFactory(transport, protocol)
}

func closeThriftClient(conn service_router.ThriftClient) error {
	return conn.(*predictor.PredictorServiceClient).Transport.Close()
}

func thriftIsOpen(conn service_router.ThriftClient) bool {
	return conn.(*predictor.PredictorServiceClient).Transport.IsOpen()
}

type ThriftProcessRequestFunc func(c *predictor.PredictorServiceClient) (interface{}, error)

type callResp struct {
	resp interface{}
	err  error
}

func (predictorClient *PredictorClient) SetIpDiffRange(ipDiffRange int) {
    predictorClient.clientOption.LocalFirstConfig.DiffRange = ipDiffRange;
}

func (predictorClient *PredictorClient) SetTimeout(timeout time.Duration) {
	predictorClient.config.Timeout = timeout
}

func (predictorClient *PredictorClient) SetMaxConnPerServer(maxConnPerServer int) {
	predictorClient.config.MaxConnPerServer = maxConnPerServer
}

func (predictorClient *PredictorClient) SetWaitConnection(wait bool) {
	predictorClient.config.WaitConnection = wait
}

func (predictorClient *PredictorClient) commonCall(callFunc ThriftProcessRequestFunc) (interface{}, error) {
	server, ok := predictorClient.router.Discover(predictorClient.clientOption)
	if !ok {
		err := fmt.Errorf("failed to find a predictor server")
		return nil, err
	}
	thriftConfig := service_router.ThriftConfig{
		Host:              server.Host,
		Port:              int(server.Port),
		TransportType:     service_router.THRIFT_TRANSPORT_HEADER,
		CompressionMethod: service_router.ThriftCompressionMethod_None,
		NewThriftClient:   newPredictorServiceClient,
		CloseThriftClient: closeThriftClient,
		ThriftIsOpen:      thriftIsOpen,
	}
	client, err := predictorClient.connGroup.GetConnection(
		thriftConfig,
		service_router.PoolMaxActive(predictorClient.config.MaxConnPerServer),
		service_router.PoolWait(predictorClient.config.WaitConnection),
	)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), predictorClient.config.Timeout)
	defer cancel()
	response := func() <-chan *callResp {
		ch := make(chan *callResp, 1)
		go func() {
			client.Do(func(conn interface{}) (interface{}, error) {
				callResult, err := callFunc(conn.(*predictor.PredictorServiceClient))
				return callResult, err
			}).Done(func() {
				// do nothing.
			}).OnError(func(err error) {
				client.SetErr(err)
				client.Close()
				ch <- &callResp{resp: nil, err: err}
			}).OnSuccess(func(resp interface{}) {
				client.Close()
				ch <- &callResp{resp: resp, err: nil}
			})
		}()
		return ch
	}()
	select {
	case <-ctx.Done():
		// timeout
		predictorClient.meters[PREDICTOR_CLIENT_TIMEOUT_METER].Mark(1)
		return nil, errors.New(RPC_TIMEOUT)
	case res := <-response:
		return res.resp, res.err
	}
}

func (predictorClient *PredictorClient) Predict(requests *predictor.PredictRequests) (*predictor.PredictResponses, error) {
	if (requests.RequestOption == nil) {
		requests.RequestOption = predictor.NewRequestOption()
	}
	st := predictorClient.timers[PREDICTOR_PREDIC_TIMER].Start()
	resp, err := predictorClient.commonCall(func(c *predictor.PredictorServiceClient) (interface{}, error) {
		return c.Predict(requests)
	})
	st.Stop()
	if err != nil {
		if err.Error() != RPC_TIMEOUT {
			predictorClient.meters[PREDICTOR_PREDICT_ERROR].Mark(1)
		}
		return nil, err
	}
	return resp.(*predictor.PredictResponses), err
}

func (predictorClient *PredictorClient) CalculateVector(requests *predictor.CalculateVectorRequests) (*predictor.CalculateVectorResponses, error) {
	if (requests.RequestOption == nil) {
		requests.RequestOption = predictor.NewRequestOption()
	}
	st := predictorClient.timers[PREDICTOR_CALCULATE_VECTOR_TIMER].Start()
	predictorClient.meters[PREDICTOR_CALCULATE_VECTOR_METER].Mark(1)
	resp, err := predictorClient.commonCall(func(c *predictor.PredictorServiceClient) (interface{}, error) {
		return c.CalculateVector(requests)
	})
	st.Stop()
	if err != nil {
		if err.Error() != RPC_TIMEOUT {
			predictorClient.meters[PREDICTOR_CALCULATE_VECTOR_ERROR].Mark(1)
		}
		return nil, err
	}
	return resp.(*predictor.CalculateVectorResponses), err
}

func (predictorClient *PredictorClient) CalculateBatchVector(requests *predictor.CalculateBatchVectorRequests) (*predictor.CalculateBatchVectorResponses, error) {
	if (requests.RequestOption == nil) {
		requests.RequestOption = predictor.NewRequestOption()
	}
	st := predictorClient.timers[PREDICTOR_CALCULATE_BATCH_VECTOR_TIMER].Start()
	predictorClient.meters[PREDICTOR_CALCULATE_BATCH_VECTOR_METER].Mark(1)
	resp, err := predictorClient.commonCall(func(c *predictor.PredictorServiceClient) (interface{}, error) {
		return c.CalculateBatchVector(requests)
	})
	st.Stop()
	if err != nil {
		if err.Error() != RPC_TIMEOUT {
			predictorClient.meters[PREDICTOR_CALCULATE_BATCH_VECTOR_ERROR].Mark(1)
		}
		return nil, err
	}
	return resp.(*predictor.CalculateBatchVectorResponses), err
}