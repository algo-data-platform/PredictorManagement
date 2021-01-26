package service

import (
	"server/common"
	"server/conf"
	"server/server/logics"
	"server/util"
	"time"
)

func TimeUpdateNodeInfos(conf *conf.Conf) {
	checkTicker := time.NewTicker(time.Second * time.Duration(conf.Monitor.CheckInterval))
	for {
		select {
		case <-checkTicker.C:
			UpdateNodeInfos(conf)
		}
	}
}

func UpdateNodeInfos(conf_info *conf.Conf) {
	common.GNodeInfos = logics.GetNodeInfoList(conf_info.PredictorHttpPort, conf_info.HttpTimeout)
	logics.NodeResMap = make(map[string]util.NodeResourceInfo)
	for _, nodeInfo := range common.GNodeInfos {
		logics.NodeResMap[nodeInfo.Host] = nodeInfo.ResourceInfo
	}
}
