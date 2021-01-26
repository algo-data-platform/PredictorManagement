package logics

import (
	"fmt"
	"github.com/algo-data-platform/predictor/golibs/adgo/predictor_client"
	"github.com/algo-data-platform/predictor/golibs/adgo/predictor_client/if/predictor"
	"server/env"
	"server/libs/logger"
	"server/util"
	"sync"
	"time"
)

var loadPredictorClientMap sync.Map

func CalculateVector(reqData *util.CalculateRequest) (*predictor.CalculateVectorResponses, error) {
	// call predictor
	t1 := time.Now()
	requests := predictor.NewCalculateVectorRequests()
	requests.Reqs = reqData.Reqs
	serviceName := reqData.ServiceName
	timeoutMS := reqData.TimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 60
	}
	predictorClient, err := GetPredictorClient(serviceName)
	if err != nil {
		return nil, fmt.Errorf("get predictor client is fail, service_name: %s", serviceName)
	}
	predictorClient.SetTimeout(time.Duration(timeoutMS) * time.Millisecond)
	responses, err := predictorClient.CalculateVector(requests)
	t2 := time.Now()
	diff := t2.Sub(t1)
	logger.Debugf("predictorClient.CalculateVector() cost: %v\n", diff)

	if err != nil {
		logger.Errorf("predictorClient.CalculateVector fail, err: %v", err)
		return responses, err
	} else if len(responses.Resps) == 0 {
		logger.Errorf("cpredictorClient.CalculateVector fail, empty response")
		return responses, fmt.Errorf("empty response")
	} else {
		logger.Debugf("calculateVector responses: %v", responses)
		return responses, nil
	}
}

func InitPredictorClient(service_name string) (*predictor_client.PredictorClient, error) {
	var err error
	predictorClient, err := predictor_client.NewPredictorClient(env.Env.Conf.PredictorClient.ConsulAddress, service_name, env.Env.LocalIp)
	predictorClient.SetMaxConnPerServer(8)
	predictorClient.SetWaitConnection(true)
	// 召回load balance为同机房策略，256 * 256 表示IP后两段不同
	predictorClient.SetIpDiffRange(65535)
	if err != nil {
		// 如果初始化不成功，直接退出
		logger.Errorf("failed to init predictor client, serivce_name: %s", service_name)
		return predictorClient, err
	}
	return predictorClient, nil
}

func GetPredictorClient(service_name string) (*predictor_client.PredictorClient, error) {
	var instanceClient interface{}
	var ok bool
	if instanceClient, ok = loadPredictorClientMap.Load(service_name); !ok {
		predictorClient, err := InitPredictorClient(service_name)
		if err != nil {
			return nil, err
		}
		instanceClient, ok = loadPredictorClientMap.LoadOrStore(service_name, predictorClient)
	}
	return instanceClient.(*predictor_client.PredictorClient), nil
}
