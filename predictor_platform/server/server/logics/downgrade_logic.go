package logics

import (
	"encoding/json"
	"fmt"
	"server/env"
	"server/libs/logger"
	"server/schema"
	"server/server/dao"
	"server/util"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type modelPercentMap map[string]int

func DowngradeService(sid uint, percent int) (uint32, uint32, error) {
	var succCount uint32 = 0
	var failCount uint32 = 0
	// 获取sid下的机器
	ipList, err := dao.GetIpListBySid(sid)
	if err != nil {
		return succCount, failCount, err
	}
	ipModelPercentMap, err := getIpModelPercentMap(ipList, percent)
	// 并发调用
	var wg sync.WaitGroup
	wg.Add(len(ipList))
	for ip, modelPercentMap := range ipModelPercentMap {
		go func(ip string, modelPercentMap map[string]int) {
			defer wg.Done()
			err := notifyPredictorDowngrade(ip, modelPercentMap)
			if err != nil {
				logger.Errorf("notifyPredictorDowngrade fail, err: %v", err)
				atomic.AddUint32(&failCount, 1)
			} else {
				atomic.AddUint32(&succCount, 1)
			}
		}(ip, modelPercentMap)
	}
	wg.Wait()
	return succCount, failCount, nil
}

func getIpModelPercentMap(ipList []string, percent int) (map[string]modelPercentMap, error) {
	var ipModelPercentMap = make(map[string]modelPercentMap)
	// 获取机器所在的sid
	ipSids, err := dao.GetIpSidsByIps(ipList)
	if err != nil {
		return nil, err
	}
	// 获取sidMap
	var sidMap = make(map[uint]bool)
	var ipSidsMap = make(map[string][]uint)
	for _, ipsid := range ipSids {
		if _, exists := sidMap[ipsid.Sid]; !exists {
			sidMap[ipsid.Sid] = true
		}
		if _, exists := ipSidsMap[ipsid.Ip]; !exists {
			ipSidsMap[ipsid.Ip] = []uint{ipsid.Sid}
		} else {
			ipSidsMap[ipsid.Ip] = append(ipSidsMap[ipsid.Ip], ipsid.Sid)
		}
	}
	var sidModelsMap = make(map[uint][]schema.Model)
	for sid, _ := range sidMap {
		// 获取service 下的所有模型
		models, err := dao.GetModelsBySid(sid)
		if err != nil {
			return nil, err
		}
		sidModelsMap[sid] = models
	}

	// 拼装请求map
	for ip, sids := range ipSidsMap {
		var modelPercentMap = make(map[string]int)
		for _, sid := range sids {
			if models, exists := sidModelsMap[sid]; exists {
				for _, model := range models {
					if model.Name == "" {
						continue
					}
					modelPercentMap[model.Name] = percent
				}
			}
		}
		ipModelPercentMap[ip] = modelPercentMap
	}
	return ipModelPercentMap, nil
}

func ResetDowngradeService(sid uint) (uint32, uint32, error) {
	var succCount uint32 = 0
	var failCount uint32 = 0
	// 获取sid下的机器
	ipList, err := dao.GetIpListBySid(sid)
	if err != nil {
		return succCount, failCount, err
	}

	// 并发调用
	var wg sync.WaitGroup
	wg.Add(len(ipList))
	for _, ip := range ipList {
		go func(ip string) {
			defer wg.Done()
			err := notifyPredictorResetDowngrade(ip)
			if err != nil {
				logger.Errorf("notifyPredictorResetDowngrade fail, err: %v", err)
				atomic.AddUint32(&failCount, 1)
			} else {
				atomic.AddUint32(&succCount, 1)
			}
		}(ip)
	}
	wg.Wait()
	return succCount, failCount, nil
}

// 获取prometheus 服务降级百分比
// @return map[service_name]percent
func GetPromDowngradePercent() (map[string]float64, error) {
	var servicePercentMap = make(map[string]float64)
	var pmSql = `sum(ad_core_predictor_downgrade{meter_type="min_1"} ) by (business_line) * 100 / (sum(ad_core_predictor_downgrade{meter_type="min_1"}) by (business_line) + sum(ad_core_predictor_model_consuming{timers_type="min_1"}) by (business_line))`
	promResp, err := env.Env.Prom.Query(pmSql, time.Second*2)
	if err != nil {
		logger.Errorf("prom query fail, err: %v", err)
		return servicePercentMap, err
	}
	if promResp != nil {
		if promResp.Status == "error" {
			return servicePercentMap, fmt.Errorf("prom query got error, error: %s", promResp.Error)
		}
		if promResp.Data.ResultType == "vector" {
			for _, vectorResults := range promResp.Data.Result {
				var businessLine string
				var percent float64
				var ok bool
				resMap, ok := vectorResults.(map[string]interface{})
				if !ok {
					break
				}
				if metric, exists := resMap["metric"]; exists {
					if metricMap, ok := metric.(map[string]interface{}); ok {
						businessLineI, exists := metricMap["business_line"]
						if !exists {
							break
						}
						businessLine, ok = businessLineI.(string)
						if !ok {
							break
						}
					}
				}
				if value, exists := resMap["value"]; exists {
					if values, ok := value.([]interface{}); ok {
						if len(values) >= 2 {
							percentStr, ok := values[1].(string)
							if !ok {
								break
							}
							var err error
							percent, err = strconv.ParseFloat(percentStr, 64)
							if err != nil {
								return servicePercentMap, fmt.Errorf("strconv.ParseFloat percent fail, err: %v", err)
							}
							percent, err = strconv.ParseFloat(fmt.Sprintf("%.1f", percent), 64)
							if err != nil {
								return servicePercentMap, fmt.Errorf("strconv.ParseFloat percent fail, err: %v", err)
							}
						}
					}
				}
				if businessLine != "" {
					servicePercentMap[businessLine] = percent
				}
			}
		}
	}
	return servicePercentMap, nil
}

// 通知predictor降级
// return error
func notifyPredictorDowngrade(ip string, modelPercentMap map[string]int) error {
	var reqUrl = fmt.Sprintf("http://%s:%d/update_downgrade_percent", ip, env.Env.Conf.PredictorHttpPort)

	requestBody, err := json.Marshal(modelPercentMap)
	respBody, err := util.HTTPPost(reqUrl, "application/json", requestBody, time.Second*2, 1)
	if err != nil {
		return fmt.Errorf("HTTPPost err: %+v, reqUrl: %s", err, reqUrl)
	}
	downgradeResp := &util.DowngradeResp{}
	err = json.Unmarshal(respBody, downgradeResp)
	if err != nil {
		return fmt.Errorf("Unmarshal http response err: %v, url: %v ,body : %s", err, reqUrl, string(respBody))
	}
	if downgradeResp.Code != 0 {
		return fmt.Errorf("updateDowngradePercent response fail , url: %v ,downgradeResp : %+v", reqUrl, downgradeResp)
	}
	return nil
}

// 通知predictor重置降级
// @return error
func notifyPredictorResetDowngrade(ip string) error {
	var reqUrl = fmt.Sprintf("http://%s:%d/reset_downgrade_percent", ip, env.Env.Conf.PredictorHttpPort)
	respBody, err := util.GetContentFromUrl(reqUrl, time.Millisecond*2000, 1)
	if err != nil {
		return fmt.Errorf("HTTPGetUrl err: %+v, reqUrl: %s", err, reqUrl)
	}
	var downgradeResp = &util.DowngradeResp{}
	err = json.Unmarshal(respBody, downgradeResp)
	if err != nil {
		return fmt.Errorf("Unmarshal http response err: %v, url: %v ,body : %s", err, reqUrl, string(respBody))
	}
	if downgradeResp.Code != 0 {
		return fmt.Errorf("resetDowngradePercent response fail , url: %v ,downgradeResp : %+v", reqUrl, downgradeResp)
	}
	return nil
}
