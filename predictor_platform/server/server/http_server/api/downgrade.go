package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"server/common/response"
	"server/libs/logger"
	"server/server/dao"
	"server/server/logics"
	"strconv"
)

// 根据service降级
func SetDowngradeByService(context *gin.Context) {
	sid, err := strconv.Atoi(context.Query("sid"))
	if err != nil {
		logger.Errorf("sid type error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("sid type error: %v", err), context)
		return
	}
	percent, err := strconv.Atoi(context.Query("percent"))
	if err != nil {
		logger.Errorf("percent type error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("percent type error: %v", err), context)
		return
	}

	// 判断service 是否存在
	_, err = dao.GetServiceBySid(uint(sid))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.ResultWithoutData(201, "service is not found", context)
		return
	}
	// 判断percent 范围
	if percent < 0 || percent > 100 {
		fmt.Println("percent 应该是0-100的数值")
		response.ResultWithoutData(201, "percent 应该是0-100的数值", context)
		return
	}

	succCount, failCount, err := logics.DowngradeService(uint(sid), percent)
	if err != nil {
		response.ResultWithoutData(202, fmt.Sprintf("降级服务失败, err: %v", err), context)
		return
	}
	if failCount > 0 {
		logger.Errorf("降级服务失败, succCount: %d, failCount: %d", succCount, failCount)
		response.ResultWithoutData(202, fmt.Sprintf("降级服务失败, 失败机器数：%d", failCount), context)
	} else {
		response.DoneWithMessage("降级服务成功", context)
	}
	return
}

// 按service重置降级
func ResetDowngradeByService(context *gin.Context) {
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
	succCount, failCount, err := logics.ResetDowngradeService(uint(sid))
	if err != nil {
		response.ResultWithoutData(202, fmt.Sprintf("重置降级服务失败, err: %v", err), context)
		return
	}
	if failCount > 0 {
		logger.Errorf("重置降级服务失败, succCount: %d, failCount: %d", succCount, failCount)
		response.ResultWithoutData(202, fmt.Sprintf("重置降级服务失败, 失败机器数: %d", failCount), context)
	} else {
		response.DoneWithMessage("重置降级服务成功", context)
	}
	return
}

func GetPromDowngradePercent(context *gin.Context) {
	servicePercentMap, err := logics.GetPromDowngradePercent()
	if err != nil {
		response.ResultWithoutData(202, fmt.Sprintf("获取prometheus降级数据失败, err: %v", err), context)
		return
	}
	response.DoneWithData(servicePercentMap, context)
	return
}
