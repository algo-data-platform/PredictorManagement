package main

import (
	"fmt"
	"os"
	"time"

	"github.com/algo-data-platform/predictor/golibs/adgo/ads_common_go/common/ip"
	"github.com/algo-data-platform/predictor/golibs/adgo/feature_master/if/feature_master"
	"github.com/algo-data-platform/predictor/golibs/adgo/predictor_client"
	"github.com/algo-data-platform/predictor/golibs/adgo/predictor_client/if/predictor"
)

func testPredict(predictorClient *predictor_client.PredictorClient) {
	// construct request
	request := predictor.NewPredictRequest()
	request.ReqID = "100001"
	request.Channel = "test"
	request.ModelName = "llearner_lrfea_v0_fans_economy"

	// populate common features
	common_features := feature_master.NewFeatures()
	uid_feature := feature_master.NewFeature()
	uid_feature.FeatureName = "uid"
	uid_feature.Int64Values = append(uid_feature.Int64Values, 300)
	uid_feature.FeatureType = feature_master.FeatureType_INT64_LIST
	common_features.Features = append(common_features.Features, uid_feature)
	f1_feature := feature_master.NewFeature()
	f1_feature.FeatureName = "f1"
	f1_feature.FloatValues = append(f1_feature.FloatValues, 222.222)
	f1_feature.FeatureType = feature_master.FeatureType_FLOAT_LIST
	common_features.Features = append(common_features.Features, f1_feature)
	gender_feature := feature_master.NewFeature()
	gender_feature.FeatureName = "gender"
	gender_feature.FeatureType = feature_master.FeatureType_STRING_LIST
	gender_feature.StringValues = append(gender_feature.StringValues, "402")
	common_features.Features = append(common_features.Features, f1_feature)
	request.CommonFeatures = common_features

	// populate item features
	item_features := feature_master.NewFeatures()
	ind_id_feature := feature_master.NewFeature()
	ind_id_feature.FeatureName = "ind_id"
	ind_id_feature.Int64Values = append(ind_id_feature.Int64Values, 20003)
	item_features.Features = append(item_features.Features, ind_id_feature)
	request.ItemFeatures = make(map[int64]*feature_master.Features)
	request.ItemFeatures[1] = item_features

	requests := predictor.NewPredictRequests()
	requests.Reqs = append(requests.Reqs, request)

	// call predictor
	responses, err := predictorClient.Predict(requests)
	if err != nil {
		fmt.Fprintln(os.Stderr, "predict failed, error: ", err)
	} else {
		fmt.Println("predict responses: ", responses)
	}
}

func testCalculateVector(predictorClient *predictor_client.PredictorClient) {
	// construct request
	request := predictor.NewCalculateVectorRequest()
	request.ReqID = "100001"
	request.Channel = "test"
	request.ModelName = "tf_xfea_estimator_v0_superfans_retrieval"

	// populate features
	features := feature_master.NewFeatures()
	{
		feature := feature_master.NewFeature()
		feature.FeatureName = "gender"
		feature.StringValues = append(feature.StringValues, "300")
		features.Features = append(features.Features, feature)
	}
	{
		feature := feature_master.NewFeature()
		feature.FeatureName = "location_code2"
		feature.StringValues = append(feature.StringValues, "300")
		features.Features = append(features.Features, feature)
	}
	request.Features = features
	request.OutputNames = append(request.OutputNames, "user_vec")
	request.OutputNames = append(request.OutputNames, "predict")

	requests := predictor.NewCalculateVectorRequests()
	requests.Reqs = append(requests.Reqs, request)

	// call predictor
	responses, err := predictorClient.CalculateVector(requests)
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
	predictorClient, err := predictor_client.NewPredictorClient("10.85.57.204:8500", "predictor_service_dev_changyu", localIp)
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
    time.Sleep(1 * time.Second)
		testPredict(predictorClient)
		//testCalculateVector(predictorClient)
	}
	fmt.Println("========test timeout========")
	predictorClient.SetTimeout(1 * time.Millisecond)
	testPredict(predictorClient)
	testCalculateVector(predictorClient)
	time.Sleep(time.Duration(2000) * time.Second)
}
