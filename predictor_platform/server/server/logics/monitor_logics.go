package logics

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"os/exec"
	"path"
	"regexp"
	"server/conf"
	"server/libs/logger"
	"server/metrics"
	"server/schema"
	"server/server/dao"
	"server/util"
	"strconv"
	"strings"
	"time"
)

const UnSetTimeStampInterval int = 3600

// 获取线上服务的模型版本信息map
func GetOnlineModelMap(conf *conf.Conf, node_infos []util.NodeInfo) map[string]util.ModelVersionInfo {
	onlineModelMap := make(map[string]util.ModelVersionInfo)
	for _, node_info := range node_infos {
		if len(node_info.StatusInfo) <= 0 {
			continue
		}
		for _, serviceInfo := range node_info.StatusInfo {
			// 过滤不需要监控的服务
			if util.IsInSliceString(serviceInfo.ServiceName, conf.Monitor.ExcludedServices) {
				continue
			}
			for _, modelRecord := range serviceInfo.ModelRecords {
				uniqKey := fmt.Sprintf("%s-%s", modelRecord.Name, modelRecord.Timestamp)
				if _, ok := onlineModelMap[uniqKey]; ok {
					continue
				}
				onlineModelMap[uniqKey] = util.ModelVersionInfo{
					ModelName:      modelRecord.Name,
					ModelTimestamp: modelRecord.Timestamp,
				}
			}
		}
	}
	return onlineModelMap
}

// 通过rsync --stats 获取远端机器的模型文件大小
func GetRemoteModelSize(conf *conf.Conf, modelNameWithVersion string) (int64, error) {
	modelPath := path.Join(conf.ModelTransmit.SrcPath, modelNameWithVersion)
	descPath := "/tmp/test/" // 由于并不是真正的传送文件，所以目标目录可以随便写,不用真实存在
	cmd := exec.Command("/bin/rsync", "-az", "-n", "--stats", conf.ModelTransmit.SrcHost+"::"+modelPath, descPath)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("cmd failed: modelNameWithVersion:%s, err:%v, output:%s", modelNameWithVersion, err, string(stdout))
	}
	// 返回示例"Total file size: 189119253 bytes"，
	stdoutStr := string(stdout)
	reg := regexp.MustCompile(`(?U)Total\sfile\ssize:\s(\d+)\sbytes`)
	findSlice := reg.FindStringSubmatch(stdoutStr)
	modelSizeStr := ""
	if len(findSlice) > 0 {
		modelSizeStr = findSlice[1]
	} else {
		return 0, fmt.Errorf("regexp did not find a matching string")
	}
	modelSize, err := strconv.ParseInt(modelSizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("strconv.ParseInt failed: err:%v, modelSizeStr:%s", err, modelSizeStr)
	}
	return modelSize, nil
}

// 获取服务下的模型版本列表
// map[service_name][]model_fullname
func GetDBServiceModelFullNameMap() (map[string][]string, error) {
	dbServiceModelFullNameMap := make(map[string][]string)
	// 获取数据库service -> model 映射关系 map[service]modelList
	dbServiceModelMap := dao.GetServiceModelMap()
	for serviceName, modelList := range dbServiceModelMap {
		modelFullNameList := []string{}
		var lastModelHistory = &schema.ModelHistory{}
		var err error
		for _, modelName := range modelList {
			// 获取last validate timestamp
			lastModelHistory, err = dao.GetLockedValidModelHistory(modelName)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				lastModelHistory, err = dao.GetLatestValidModelHistory(modelName)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				} else if err != nil {
					return dbServiceModelFullNameMap, err
				}
			} else if err != nil {
				return dbServiceModelFullNameMap, err
			}
			modelFullNameList = append(modelFullNameList, modelName+"-"+lastModelHistory.Timestamp)
		}
		dbServiceModelFullNameMap[serviceName] = modelFullNameList
	}
	return dbServiceModelFullNameMap, nil
}

// 获取模型版本距离现在最大的时间间隔
func GetMaxLoadInterval(fullModelList []string) int {
	var maxLoadInterval int
	for _, fullModel := range fullModelList {
		model := strings.Split(fullModel, "-")
		if len(model) > 1 {
			version := model[1]
			versionTime, err := time.ParseInLocation("20060102_150405", version, time.Local)
			if err != nil {
				maxLoadInterval = util.MaxInt(UnSetTimeStampInterval, maxLoadInterval)
				// 添加err metrics
				metrics.GetMeters()[metrics.TAG_PARSE_TIMESTAMP_ERROR].Mark(1)
				logger.Errorf("time.Parse version fail, err: %v, version: %s", err, version)
			} else {
				interval := time.Now().Sub(versionTime).Seconds()
				maxLoadInterval = util.MaxInt(int(interval), maxLoadInterval)
			}
		}
	}
	return maxLoadInterval
}
