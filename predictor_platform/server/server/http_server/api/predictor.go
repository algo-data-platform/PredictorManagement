package api

import (
	"encoding/json"
	"fmt"
	"github.com/algo-data-platform/predictor/golibs/adgo/predictor_client/if/predictor"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"server/common/response"
	"server/libs/logger"
	"server/server/logics"
	"server/util"
)

// 调用prdictor 召回接口
// input example
/*
{
    "service_name":"retrieval_ad_service",
    "timeout_ms":100,
    "reqs":[
        {
            "req_id":"20201023191042_21373",
            "channel":"advec",
            "model_name":"tf_xfea_estimator_v2s1_sfst_retrieval_app_ad",
            "features":{
                "features":[
                    {
                        "feature_name":"label",
                        "feature_type":1,
                        "string_values":[
                            "1"
                        ],
                        "float_values":null,
                        "int64_values":null
                    }
                ]
            },
            "output_names":[
                "ad_vec"
            ]
        }
    ]
}
*/
// output example
/*
{
    "code":0,
    "data":{
        "resps":[
            {
                "req_id":"20201023191042_21373",
                "model_timestamp":"20200801",
                "vector_map":{
                    "ad_vec":[
                        46061.6484375,
                        -16496.62109375
                    ]
                },
                "return_code":0
            }
        ]
    },
    "msg":"Done"
}
*/
func PredictorCalculateVector(context *gin.Context) {
	requestBody, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		logger.Errorf("ioutil.ReadAll err: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("ioutil.ReadAll err: %v", err), context)
		return
	}
	logger.Debugf("requestBody: %s", string(requestBody))

	// todo 数据转为json
	var reqData util.CalculateRequest
	// todo 解析json
	err = json.Unmarshal(requestBody, &reqData)
	if err != nil {
		logger.Errorf("json.Unmashal error: %v", err)
		response.ResultWithoutData(202, fmt.Sprintf("json.Unmashal error: %v", err), context)
		return
	}
	if reqData.ServiceName == "" {
		response.ResultWithoutData(203, fmt.Sprintf("predictor_service_name is empty"), context)
		return
	}
	var respData *predictor.CalculateVectorResponses
	respData, err = logics.CalculateVector(&reqData)
	if err != nil {
		response.ResultWithoutData(204, fmt.Sprintf("CalculateVector error: %v", err), context)
		return
	}
	response.DoneWithData(respData, context)
}
