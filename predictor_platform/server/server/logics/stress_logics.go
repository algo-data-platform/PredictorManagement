package logics

import (
	"fmt"
	"server/env"
	"server/libs/logger"
	"server/schema"
	"server/server/dao"
	"server/util"
	"strconv"
	"strings"
)

// 添加压测
func EnableStressTest(stressInfo *schema.StressInfo) error {
	// 获取机器原始sid
	originSids, err := getOriginSids(stressInfo.Hid)
	if err != nil {
		return err
	}
	stressInfo.OriginSids = originSids
	stressTestService, err := dao.GetServiceByName(env.Env.Conf.StressTestService)
	if err != nil || stressTestService.ID == 0 {
		logger.Errorf("GetServiceByName fail, err: %v", err)
		return err
	}
	// 事务开启压测
	err = dao.TransactEnableStressTest(stressInfo, stressTestService)
	if err != nil {
		return err
	}
	return nil
}

// 关闭压测任务
func DisableStressTest(stressId uint) error {
	stressInfo, err := dao.GetStressInfoById(stressId)
	if err != nil {
		logger.Errorf("GetStressInfoById fail, err:%v", err)
		return err
	}
	if stressInfo.IsEnable == 0 {
		return fmt.Errorf("current stress is already closed")
	}
	stressTestService, err := dao.GetServiceByName(env.Env.Conf.StressTestService)
	if err != nil || stressTestService.ID == 0 {
		logger.Errorf("GetServiceByName fail, err: %v", err)
		return err
	}
	// 事务关闭压测
	err = dao.TransactDisableStressTest(stressInfo, stressTestService)
	if err != nil {
		return err
	}
	return nil
}

func getOriginSids(hid uint) (string, error) {
	originSids := ""
	hostServices, err := dao.GetHostServicesByHid(hid)
	if err != nil {
		logger.Errorf("GetHostServicesByHid fail, err:%v", err)
		return originSids, err
	}
	if len(hostServices) > 0 {
		var originSidStrs []string
		for _, row := range hostServices {
			originSidStrs = append(originSidStrs, fmt.Sprintf("%d_%d", row.Sid, row.LoadWeight))
		}
		originSids = strings.Join(originSidStrs, ",")
	}
	return originSids, nil
}

func GetStressList(is_enable uint) ([]*util.StressInfo, error) {
	stressList := []*util.StressInfo{}
	var err error
	var stressInfos []*schema.StressInfo
	if is_enable == uint(2) {
		stressInfos, err = dao.GetAllStressInfos()
	} else {
		stressInfos, err = dao.GetStressInfosByStatus(is_enable)
	}
	if err != nil {
		return stressList, err
	}
	// 批量获取机器及模型
	var allHids []uint
	var allMids []uint
	for _, stressInfo := range stressInfos {
		if !util.IsInSliceUint(stressInfo.Hid, allHids) {
			allHids = append(allHids, stressInfo.Hid)
		}
		curMids := []uint{}
		if stressInfo.Mids != "" {
			midstrs := strings.Split(stressInfo.Mids, ",")
			for _, midStr := range midstrs {
				midUint, _ := strconv.ParseUint(midStr, 10, 64)
				curMids = append(curMids, uint(midUint))
				allMids = append(allMids, uint(midUint))
			}
		}
		curQps := []uint{}
		if stressInfo.Qps != "" {
			qpsstrs := strings.Split(stressInfo.Qps, ",")
			for _, qpsStr := range qpsstrs {
				qpsUint, _ := strconv.ParseUint(qpsStr, 10, 64)
				curQps = append(curQps, uint(qpsUint))
			}
		}
		stressRow := &util.StressInfo{
			ID:         stressInfo.ID,
			Hid:        stressInfo.Hid,
			Mids:       curMids,
			QPS:        curQps,
			CreateTime: stressInfo.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdateTime: stressInfo.UpdatedAt.Format("2006-01-02 15:04:05"),
			IsEnable:   stressInfo.IsEnable,
		}
		stressList = append(stressList, stressRow)
	}
	allHostMap, err := dao.GetHostMapByIds(allHids)
	if err != nil {
		return stressList, err
	}
	allModelMap, err := dao.GetModelMapByIds(allMids)
	if err != nil {
		return stressList, err
	}
	for _, stress := range stressList {
		if _, exists := allHostMap[stress.Hid]; exists {
			stress.IP = allHostMap[stress.Hid].Ip
		}
		modelNames := []string{}
		if len(stress.Mids) > 0 {
			for _, mid := range stress.Mids {
				if _, exists := allModelMap[mid]; exists {
					modelNames = append(modelNames, allModelMap[mid].Name)
				} else {
					modelNames = append(modelNames, "")
				}
			}
		}
		stress.ModelNames = modelNames
	}
	return stressList, nil
}
