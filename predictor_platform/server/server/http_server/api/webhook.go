package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"server/common/response"
	"server/libs/logger"
	"server/server/logics"
	"server/util"
)

// webhook alert
func AlertWebhook(context *gin.Context) {
	tag := context.Query("tag")
	requestBody, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		logger.Errorf("ioutil.ReadAll err: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("ioutil.ReadAll err: %v", err), context)
		return
	}
	logger.Debugf("requestBody: %s", string(requestBody))

	// todo 数据转为json
	var reqData util.WebHookRequest
	// todo 解析json
	err = json.Unmarshal(requestBody, &reqData)
	if err != nil {
		logger.Errorf("json.Unmashal error: %v", err)
		response.ResultWithoutData(202, fmt.Sprintf("json.Unmashal error: %v", err), context)
		return
	}
	webHook := logics.NewWebHook(&reqData, tag)
	err = webHook.DistributeAlert()
	if err != nil {
		logger.Errorf("DistributeAlert error: %v", err)
		response.ResultWithoutData(203, fmt.Sprintf("DistributeAlert error: %v", err), context)
		return
	}
	response.DoneWithData(string(requestBody), context)
	return
}
