package service

import (
	"fmt"
	"math"
	"server/common"
	"server/conf"
	"server/libs/logger"
	"server/metrics"
	"server/server/dao"
	"server/server/logics"
	"server/util"
	"time"
)

func StartMonitor(conf *conf.Conf) {
	checkTicker := time.NewTicker(time.Second * time.Duration(conf.Monitor.CheckInterval))
	for {
		select {
		case <-checkTicker.C:
			check(conf)
		}
	}
}

func check(conf *conf.Conf) {
	// check 过期模型版本，上报模型指标
	go checkStaleModelVersion(conf)
	// check 加载服务及模型不一致，上报服务指标
	go checkServiceModelDiff(conf)
	// 监控模型大小
	go checkModelSize(conf)
}

// 模型指标监控, 将线上机器及模型版本距离当前时间上报Metrics
func checkStaleModelVersion(conf *conf.Conf) {
	// 获取数据库service -> model 映射关系 map[service]modelList
	dbServiceModelMap := dao.GetServiceModelMap()
	existKeyMap := make(map[string]bool, 0)
	for _, node_info := range common.GNodeInfos {
		if len(node_info.StatusInfo) <= 0 {
			continue
		}
		for _, serviceInfo := range node_info.StatusInfo {
			// 过滤不需要监控的服务
			if util.IsInSliceString(serviceInfo.ServiceName, conf.Monitor.ExcludedServices) {
				continue
			}
			// 过滤待卸载掉的service
			if _, ok := dbServiceModelMap[serviceInfo.ServiceName]; !ok {
				logger.Warnf("find service to be unloaded, service_name:%s", serviceInfo.ServiceName)
				continue
			}
			for _, modelRecord := range serviceInfo.ModelRecords {
				// 过滤待卸载掉的模型
				if !util.IsInSliceString(modelRecord.Name, dbServiceModelMap[serviceInfo.ServiceName]) {
					logger.Warnf("find model to be unloaded, model_name:%s", modelRecord.Name)
					continue
				}
				lastVersionTime, err := time.ParseInLocation("20060102_150405", modelRecord.Timestamp, time.Local)
				if err != nil {
					metrics.GetMeters()[metrics.TAG_PARSE_TIMESTAMP_ERROR].Mark(1)
					logger.Errorf("ParseInLocation timestamp error: %v", err)
					continue
				}
				intervalSeconds := math.Floor(float64(time.Now().Sub(lastVersionTime).Seconds()))
				metrics.GetModelGauge(serviceInfo.ServiceName, node_info.Host, modelRecord.Name).Update(intervalSeconds)
				existKeyMap[metrics.GetModelGaugeUniqKey(serviceInfo.ServiceName, node_info.Host, modelRecord.Name)] = true
			}
		}
	}
	// 清除线上不存在的模型gauge 数据
	metrics.ClearNotExistModelGauge(existKeyMap)
	logger.Infof("finish checkStaleModelVersion")
}

// check 上报服务指标
func checkServiceModelDiff(conf *conf.Conf) {
	// 获取数据库host -> service 的关系  map[host]serviceList
	dbHostServiceMap, err := dao.GetHostServiceMap()
	if err != nil {
		metrics.GetMeters()[metrics.TAG_GET_DB_HOST_SERVICE_ERROR].Mark(1)
		logger.Errorf("dao.GetHostServiceMap fail, err: %v", err)
		return
	}
	// 获取数据库service -> model 映射关系 map[service]modelList
	dbServiceModelFullNameMap, err := logics.GetDBServiceModelFullNameMap()
	if err != nil {
		metrics.GetMeters()[metrics.TAG_GET_DB_MODEL_ERROR].Mark(1)
		logger.Errorf("logics.GetDBServiceModelFullNameMap fail, err: %v", err)
		return
	}

	existDiffKeyMap := make(map[string]bool, 0)
	for _, node_info := range common.GNodeInfos {
		// todo	判断机器中的加载的service是否一致
		host := node_info.Host
		// 如果当前机器线上不存在service，但数据库中存在，则为加载不一致
		if len(node_info.StatusInfo) <= 0 {
			if _, ok := dbHostServiceMap[host]; ok && len(dbHostServiceMap[host]) > 0 {
				metrics.GetModelDiffGauge("", host, metrics.TAG_SERVICE_LOAD_DIFF).Update(1)
				existDiffKeyMap[metrics.GetModelDiffGaugeUniqKey("", host, metrics.TAG_SERVICE_LOAD_DIFF)] = true
			}
			continue
		}
		onlineHostServices := make([]string, 0, len(node_info.StatusInfo))
		for _, serviceInfo := range node_info.StatusInfo {
			// 获取service
			onlineHostServices = append(onlineHostServices, serviceInfo.ServiceName)
			// service 下的modellist
			onlineModelFullNameList := make([]string, 0, len(serviceInfo.ModelRecords))
			for _, modelRecord := range serviceInfo.ModelRecords {
				if modelRecord.State == "loaded" {
					// FullName 包含模型名和timestamp
					onlineModelFullNameList = append(onlineModelFullNameList, modelRecord.FullName)
				}
			}
			// 判断service中加载的model是否一致
			if _, ok := dbServiceModelFullNameMap[serviceInfo.ServiceName]; ok {
				// 获取数据库中service下模型列表跟线上模型的差集，差集为空则认为加载正常，因为线上模型有待卸载模型存在
				diffModels := util.DiffSliceString(dbServiceModelFullNameMap[serviceInfo.ServiceName], onlineModelFullNameList)
				logger.Debugf("diffModels %v, host: %s", diffModels, host)
				if len(diffModels) > 0 {
					// 获取模型版本距离现在最大间隔时间
					maxLoadInterval := logics.GetMaxLoadInterval(diffModels)
					logger.Debugf("GetMaxLoadInterval %d, host: %s", maxLoadInterval, host)
					metrics.GetModelDiffGauge(serviceInfo.ServiceName, host, metrics.TAG_MODEL_LOAD_DIFF).Update(float64(maxLoadInterval))
					existDiffKeyMap[metrics.GetModelDiffGaugeUniqKey(serviceInfo.ServiceName, host, metrics.TAG_MODEL_LOAD_DIFF)] = true
				}
			}
		}
		// 判断机器中的加载的service是否一致
		if _, ok := dbHostServiceMap[host]; ok {
			isEqualService := util.IsEqualSliceString(dbHostServiceMap[host], onlineHostServices)
			if !isEqualService {
				metrics.GetModelDiffGauge("", host, metrics.TAG_SERVICE_LOAD_DIFF).Update(1)
				existDiffKeyMap[metrics.GetModelDiffGaugeUniqKey("", host, metrics.TAG_SERVICE_LOAD_DIFF)] = true
			}
		}
	}
	// 清除线上不存在的模型比较gauge 数据
	metrics.ClearNotExistModelDiffGauge(existDiffKeyMap)
	logger.Infof("finish checkServiceModelDiff")
}

// 模型大小监控
func checkModelSize(conf *conf.Conf) {
	existKeyMap := make(map[string]bool, 0)
	// 获取线上机器加载的所有模型，通过node_infos获取所有线上机器加载的模型，排除一致性验证机器
	// map[model_name+timestamp]ModelVersionInfo
	onlineModelMap := logics.GetOnlineModelMap(conf, common.GNodeInfos)
	for _, modelVersionInfo := range onlineModelMap {
		modelNameWithVersion := fmt.Sprintf("%s-%s", modelVersionInfo.ModelName, modelVersionInfo.ModelTimestamp)
		modelSize, err := logics.GetRemoteModelSize(conf, modelNameWithVersion)
		if err != nil {
			logger.Errorf("getRemoteModelSize failed, modelNameWithVersion=%s, err=%s", modelNameWithVersion, err)
			continue
		}
		metrics.GetModelSizeGauge(modelVersionInfo.ModelName).Update(float64(modelSize))
		existKeyMap[metrics.GetModelSizeGaugeUniqKey(modelVersionInfo.ModelName)] = true
	}
	// 清除线上不存在的模型大小gauge数据
	metrics.ClearNotExistModelSizeGauge(existKeyMap)
	logger.Infof("finish checkModelSize")
}
