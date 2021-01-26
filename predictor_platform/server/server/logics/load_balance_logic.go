package logics

import (
	"math"
	"server/env"
	"server/libs/logger"
	"server/server/dao"
	"server/util"
)

// 重置服务下所有机器权重
func ResetServiceWeight(sid uint) error {
	// 获取service下所有机器
	hostServices, err := dao.GetHostServiceInfoBySid(sid)
	if err != nil {
		logger.Errorf("GetHostServiceInfoBySid fail, err: %v", err)
		return err
	}
	// 更新权重
	for _, hostService := range hostServices {
		// 获取cpu核数
		var resourceInfo util.NodeResourceInfo
		if _, exists := NodeResMap[hostService.Ip]; exists {
			resourceInfo = NodeResMap[hostService.Ip]
		} else {
			resourceInfo = util.NodeResourceInfo{}
		}
		resetWeight := GetResetWeight(resourceInfo.CoreNum)
		updateData := map[string]interface{}{
			"load_weight": resetWeight,
		}
		err := dao.UpdateHostServiceData(hostService.Hsid, updateData)
		if err != nil {
			logger.Errorf("UpdateHostServiceData fail, hsid: %d, updateData: %v, err: %v", hostService.Hsid, updateData, err)
			return err
		}
	}
	return nil
}

// 获取重置权重,按照cpu核数计算
func GetResetWeight(coreNum int) int {
	var downGapCoreNum = 16
	var resetWeight = env.Env.Conf.LoadThreshold.Down_Gap
	if coreNum == 0 {
		return resetWeight
	}
	resetWeight = int(math.Ceil(float64(env.Env.Conf.LoadThreshold.Down_Gap) / float64(downGapCoreNum) * float64(coreNum)))
	return resetWeight
}
