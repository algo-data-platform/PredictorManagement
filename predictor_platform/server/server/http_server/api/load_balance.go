package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"server/common"
	"server/common/response"
	"server/libs/logger"
	"server/server/dao"
	"server/server/logics"
	"server/util"
	"strconv"
)

// insert load_balance service
func InsertLoadBalance(context *gin.Context) {
	service_name := context.Query("service_name")
	// validate
	if service_name == "" {
		resultData := &util.JsonRespData{
			Code: 201,
			Msg:  "input service_name is empty",
		}
		context.JSON(http.StatusOK, resultData)
		return
	}
	// 判断是否已经存在
	exists, err := dao.ExistsService(service_name)
	if err != nil {
		errMsg := fmt.Sprintf("ExistsService fail, service_name: %s, err: %v", service_name, err)
		logger.Errorf(errMsg)
		resultData := &util.JsonRespData{
			Code: 202,
			Msg:  errMsg,
		}
		context.JSON(http.StatusOK, resultData)
		return
	} else if !exists && err == nil {
		errMsg := fmt.Sprintf("service_name is not exists in db, service_name: %s", service_name)
		logger.Warnf(errMsg)
		resultData := &util.JsonRespData{
			Code: 203,
			Msg:  errMsg,
		}
		context.JSON(http.StatusOK, resultData)
		return
	}
	// 判断是否已经加入过了
	if util.IsInSliceString(service_name, common.GLoadThresholdServices) {
		errMsg := fmt.Sprintf("service_name is already inserted in conf, service_name: %s", service_name)
		logger.Warnf(errMsg)
		resultData := &util.JsonRespData{
			Code: 204,
			Msg:  errMsg,
		}
		context.JSON(http.StatusOK, resultData)
		return
	} else {
		common.GLoadThresholdServices = append(common.GLoadThresholdServices, service_name)
		errMsg := fmt.Sprintf("insert into config success, service_name: %s", service_name)
		resultData := &util.JsonRespData{
			Code: 0,
			Msg:  errMsg,
		}
		context.JSON(http.StatusOK, resultData)
		return
	}
}

// get all load_balance service
func GetLoadBalance(context *gin.Context) {
	resData := common.GLoadThresholdServices
	if common.GLoadThresholdServices == nil {
		resData = []string{}
	}
	resultData := &util.JsonRespData{
		Code: 0,
		Data: resData,
		Msg:  "get success",
	}
	context.JSON(http.StatusOK, resultData)
	return
}

// delete load_balance service
func DeleteLoadBalance(context *gin.Context) {
	service_name := context.Query("service_name")
	// 判断是否已经加入过了
	if util.IsInSliceString(service_name, common.GLoadThresholdServices) {
		util.DelSliceFirstItem(&common.GLoadThresholdServices, service_name)
		resultData := &util.JsonRespData{
			Code: 0,
			Msg:  "delete success",
		}
		context.JSON(http.StatusOK, resultData)
		return
	} else {
		errMsg := fmt.Sprintf("service_name is not exists, service_name: %s", service_name)
		resultData := &util.JsonRespData{
			Code: -1,
			Msg:  errMsg,
		}
		context.JSON(http.StatusOK, resultData)
		return
	}
}

// reset load_balance services
func ResetLoadBalance(context *gin.Context) {
	common.GLoadThresholdServices = []string{}
	resultData := &util.JsonRespData{
		Code: 0,
		Msg:  "reset success",
	}
	context.JSON(http.StatusOK, resultData)
	return
}

// 重置service下权重
func ResetLoadWeight(context *gin.Context) {
	sid, err := strconv.Atoi(context.Query("sid"))
	if err != nil {
		logger.Errorf("sid type error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("sid type error: %v", err), context)
		return
	}

	// 判断service 是否存在
	_, err = dao.GetServiceBySid(uint(sid))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.ResultWithoutData(201, "service is not found", context)
		return
	}
	err = logics.ResetServiceWeight(uint(sid))
	if err != nil {
		response.ResultWithoutData(202, fmt.Sprintf("重置服务下所有机器权重失败, err: %v", err), context)
		return
	}
	response.DoneWithMessage("重置服务下所有机器权重成功", context)
	return
}
