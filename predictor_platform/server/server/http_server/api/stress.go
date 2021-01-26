package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"server/common/response"
	"server/libs/logger"
	"server/schema"
	"server/server/dao"
	"server/server/logics"
	"strconv"
	"strings"
)

// 开启压测
func StressInsert(context *gin.Context) {
	hid, err := strconv.ParseUint(context.Query("hid"), 10, 64)
	if err != nil {
		logger.Errorf("parse hid error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("parse hid error: %v", err), context)
		return
	}
	mids := context.Query("mids")
	qps := context.Query("qps")
	stressInfo := &schema.StressInfo{
		Hid:      uint(hid),
		Mids:     mids,
		Qps:      qps,
		IsEnable: 1,
	}
	if !CheckStressInfoValid(stressInfo, context) {
		return
	}
	err = logics.EnableStressTest(stressInfo)
	if err != nil {
		response.ResultWithoutData(202, fmt.Sprintf("add stress fail, err: %v", err), context)
		return
	}
	response.Done(context)
	return
}

// 关闭压测
func StressDisable(context *gin.Context) {
	stressId, err := strconv.ParseUint(context.Query("stress_id"), 10, 64)
	if err != nil {
		logger.Errorf("parse stress_id error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("parse stress_id error: %v", err), context)
		return
	}
	err = logics.DisableStressTest(uint(stressId))
	if err != nil {
		response.ResultWithoutData(202, fmt.Sprintf("close stress fail, err: %v", err), context)
		return
	}
	response.Done(context)
	return
}

// 查看压测列表
func StressList(context *gin.Context) {
	is_enable := context.Query("is_enable")
	var is_enable_uint uint
	if is_enable == "" {
		is_enable_uint = 1
	} else {
		is_enable_int, err := strconv.Atoi(is_enable)
		if err != nil {
			logger.Errorf("parse is_enable error: %v", err)
			response.ResultWithoutData(201, fmt.Sprintf("parse is_enable error, err: %v", err), context)
			return
		}
		is_enable_uint = uint(is_enable_int)
	}
	stressList, err := logics.GetStressList(is_enable_uint)
	if err != nil {
		response.ResultWithoutData(202, fmt.Sprintf("get stress list fail, err: %v", err), context)
		return
	}
	response.DoneWithData(stressList, context)
	return
}

// 重新开启压测
func StressEnable(context *gin.Context) {
	stressId, err := strconv.ParseUint(context.Query("stress_id"), 10, 64)
	if err != nil {
		logger.Errorf("parse stress_id error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("parse stress_id error: %v", err), context)
		return
	}
	stressInfo, err := dao.GetStressInfoById(uint(stressId))
	if err != nil {
		logger.Errorf("GetStressInfoById fail, err:%v", err)
		response.ResultWithoutData(201, fmt.Sprintf("GetStressInfoById fail, err:%v", err), context)
		return
	}
	if stressInfo.IsEnable == 1 {
		response.ResultWithoutData(201, fmt.Sprintf("current stress is already enabled"), context)
		return
	}
	if !CheckStressInfoValid(stressInfo, context) {
		return
	}
	err = logics.EnableStressTest(stressInfo)
	if err != nil {
		response.ResultWithoutData(202, fmt.Sprintf("enable stress fail, err: %v", err), context)
		return
	}
	response.Done(context)
	return
}

func CheckStressInfoValid(stressInfo *schema.StressInfo, context *gin.Context) bool {
	// 校验机器是否存在，模型是否存在
	exists, err := dao.ExistsHostById(stressInfo.Hid)
	if err != nil {
		response.ResultWithoutData(201, fmt.Sprintf("ExistsHostById fail: %v", err), context)
		return false
	}
	if !exists {
		response.ResultWithoutData(201, "当前任务机器已经不存在，不可开启", context)
		return false
	}
	// 校验模型是否存在
	var midstrs = []string{}
	if stressInfo.Mids != "" {
		var allMids []uint
		midstrs = strings.Split(stressInfo.Mids, ",")
		for _, midStr := range midstrs {
			midInt, err := strconv.Atoi(midStr)
			if err != nil {
				response.ResultWithoutData(201, "mids is not int", context)
				return false
			}
			allMids = append(allMids, uint(midInt))
		}
		allModelMap, err := dao.GetModelMapByIds(allMids)
		if err != nil {
			response.ResultWithoutData(201, fmt.Sprintf("获取模型失败， err: %v", err), context)
			return false
		}
		if len(allModelMap) != len(midstrs) {
			response.ResultWithoutData(201, "压测模型已经不存在，不可开启", context)
			return false
		}
	} else {
		response.ResultWithoutData(201, "当前任务模型为空，不可开启", context)
		return false
	}
	if stressInfo.Qps == "" {
		response.ResultWithoutData(201, "qps is empty", context)
		return false
	}
	qpsstrs := strings.Split(stressInfo.Qps, ",")
	if len(midstrs) != len(qpsstrs) {
		logger.Errorf("the length of mids and qps is different,len(midstrs):%d, len(qpsstrs):%d", len(midstrs), len(qpsstrs))
		response.ResultWithoutData(201, "the length of mids and qps is different", context)
		return false
	}
	for _, qpsstr := range qpsstrs {
		_, err = strconv.Atoi(qpsstr)
		if err != nil {
			response.ResultWithoutData(201, "qps is not int", context)
			return false
		}
	}

	// 校验机器是否已经有开启任务
	exists, err = dao.ExistsEnableStressByHid(stressInfo.Hid)
	if err != nil {
		response.ResultWithoutData(201, fmt.Sprintf("校验压测任务失败，err: %v", err), context)
		return false
	}
	if exists {
		response.ResultWithoutData(201, "当前机器已经存在开启的压测任务，暂不支持单个机器开启多个任务", context)
		return false
	}
	return true
}

// 保存压测qps
func StressSaveQps(context *gin.Context) {
	stressId, err := strconv.ParseUint(context.Query("stress_id"), 10, 64)
	if err != nil {
		logger.Errorf("parse stress_id error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("parse stress_id error: %v", err), context)
		return
	}
	qps := context.Query("qps")
	if qps == "" {
		response.ResultWithoutData(201, "qps is empty", context)
		return
	}
	qpsstrs := strings.Split(qps, ",")
	for _, qpsstr := range qpsstrs {
		_, err = strconv.Atoi(qpsstr)
		if err != nil {
			response.ResultWithoutData(201, "qps is not int", context)
			return
		}
	}
	err = dao.UpdateStressQps(uint(stressId), qps)
	if err != nil {
		response.ResultWithoutData(202, fmt.Sprintf("更新压测qps失败，err: %v", err), context)
	}
	response.Done(context)
	return
}
