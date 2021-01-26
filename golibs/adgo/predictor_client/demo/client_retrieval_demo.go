package main

import (
	"fmt"
	"os"
	"time"

	"github.com/algo-data-platform/predictor/golibs/ads_common_go/common/ip"
	"github.com/algo-data-platform/predictor/golibs/adgo/feature_master/if/feature_master"
	"github.com/algo-data-platform/predictor/golibs/adgo/predictor_client"
	"github.com/algo-data-platform/predictor/golibs/adgo/predictor_client/if/predictor"
)

func testCalculateVector(predictorClient *predictor_client.PredictorClient) {
	// construct request
	request := predictor.NewCalculateVectorRequest()
	request.ReqID = "100001"
	request.Channel = "test"
	request.ModelName = "tf_xfea_estimator_v1_sfst_retrieval_app"

	// populate features
  feature_names := []string{"age","from","gender","interest_category_all","interest_word_all","isloadmore","label","lift_state_all","lift_state_top5","location_code2","logid","login_freq_num","network_type","os_brand","os_version","platform_type","ps_id","session_tblog_interest","uid","user_ip","wm"}
  feature_values := []string{"1022","10A2295010","402","10700:1,10600:1,10100:3,10300:1,10200:2,100619:2.0,100617:2.0,101902:4,101803:6,101406:7,101900:4,101404:7,101800:6,101502:7,100205:7,100601:7,101401:7","游戏行业:10.0,博主:44.0,美妆:44.0,美食:63.0,母婴育儿:10.0,美妆博主:44.0,营销号:35.0,jk制服:11.0,爆料:24.0,代购:11.0,教育-考研:1.0,护肤:44.0,育儿:10.0,娱乐八卦:22.0,单机游戏:10.0,种草:11.0,买:11.0,吃吃吃:63.0,教育-英语:6.0,教育-雅思:8.0,教育-托福:2.0,买买买:11.0,教育-教师教学:3.0,游戏媒体:10.0,吃货:63.0,幽默:8.0,母婴:10.0,八卦:35.0,囧人糗事:8.0,游戏资讯:10.0,化妆:44.0,吃:63.0,游戏:10.0,娱乐博主:33.0,美食博主:63.0","0","-","","","31202","-","30","60110304","90104000","92206001","android","70102028","","5667523306","183.197.90.186","9006_2001"}
	features := feature_master.NewFeatures()
  for idx, name := range feature_names {
    features.Features = append(
      features.Features,
      &feature_master.Feature{FeatureName: name, FeatureType: feature_master.FeatureType_STRING_LIST, StringValues: []string{feature_values[idx]}})
  }

	request.Features = features
	request.OutputNames = append(request.OutputNames, "user_vec")
	request.OutputNames = append(request.OutputNames, "predict")

	requests := predictor.NewCalculateVectorRequests()
	requests.Reqs = append(requests.Reqs, request)

	// call predictor
  t1 := time.Now()
	responses, err := predictorClient.CalculateVector(requests)
  t2 := time.Now()
  diff := t2.Sub(t1)
  fmt.Printf("---------------------------- predictorClient.CalculateVector() cost: %v\n", diff)

	if err != nil {
		fmt.Fprintln(os.Stderr, "calculateVector failed. error: ", err)
	} else {
		fmt.Println("calculateVector responses: ", responses)
        fmt.Println("return code: ", responses.Resps[0].GetReturnCode())
	}
}

func main() {
	localIp, _ := ip.GetLocalIPv4Str()
	fmt.Println("local ip is ", localIp)
	predictorClient, err := predictor_client.NewPredictorClient("10.85.57.206:8500", "algo_service_dev_maolei", localIp)
	predictorClient.SetTimeout(8 * time.Millisecond)
	predictorClient.SetMaxConnPerServer(64)
	predictorClient.SetWaitConnection(true)
	predictorClient.SetIpDiffRange(65535)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to init predictor client ")
		return
	}
	time.Sleep(3 * time.Second)
	for i := 0; i < 10000; i++ {
		testCalculateVector(predictorClient)
	}
	time.Sleep(time.Duration(2000) * time.Second)
}
