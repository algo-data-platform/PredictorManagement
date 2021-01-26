package logics

import (
	"bytes"
	"content_service/api"
	"content_service/env"
	"content_service/libs/logger"
	"content_service/schema"
	"content_service/common"
	"strings"
	"math"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

// fetch back a set of models we already have from local disk
func FetchDiskData(env *env.Env, dest_path string) (map[string]bool, error) {
	var disk_models = map[string]bool{}
	if _, err := os.Stat(dest_path); err != nil {
		return disk_models, fmt.Errorf("err in checking stat of dest_path=%v: %v", dest_path, err)
	}
	files, err := ioutil.ReadDir(dest_path)
	if err != nil {
		return disk_models, fmt.Errorf("err in reading dest_path=%v: %v", dest_path, err)
	}
	for _, file := range files {
		disk_models[file.Name()] = true
	}

	return disk_models, nil
}

func SetPredictorWorkMode(work_mode string, env *env.Env, predictor_ip string, predictor_http_port string) error {
	// send post request to notify predictor
	var url = "http://" + predictor_ip + ":" + predictor_http_port + "/set_work_mode?work_mode=" + work_mode
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http GET err: %v, url: %v", err, url)
	}

	// parse http response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("parse http response err: %v, url: %v", err, url)
	}
	var respStruct struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	err = json.Unmarshal(body, &respStruct)
	if err != nil {
		return fmt.Errorf("Unmarshal http response err: %v, url: %v ,body : %s", err, url, string(body))
	}
	if respStruct.Code != 0 {
		return fmt.Errorf("SetPredictorWorkMode response fail , url: %v ,respStruct : %+v", url, respStruct)
	}

	logger.Debugf("\n>> Response: %v\n\n", string(body))
	return nil
}

func RegisterPredictorService(service_list string, env *env.Env, predictor_ip string, predictor_http_port string) error {
	// send post request to notify predictor
	var url = "http://" + predictor_ip + ":" + predictor_http_port + "/register_service_name?service_name_list=" + service_list
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http GET err: %v, url: %v", err, url)
	}

	// parse http response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("parse http response err: %v, url: %v", err, url)
	}
	var respStruct struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	err = json.Unmarshal(body, &respStruct)
	if err != nil {
		return fmt.Errorf("Unmarshal http response err: %v, url: %v ,body : %s", err, url, string(body))
	}
	if respStruct.Code != 0 {
		return fmt.Errorf("RegisterPredictorService response fail , url: %v ,respStruct : %+v", url, respStruct)
	}

	logger.Debugf("\n>> Response: %v\n\n", string(body))
	return nil
}

// format a json body and post http request
func NotifyPredictor(
	// format data into request json
	service_models map[string]map[string]schema.ModelHistory,
	service_weight map[string]int,
	service_config map[string]string,
	env *env.Env,
	predictor_ip string,
	predictor_http_port string,
) error {
	if len(service_models) == 0 {
		return nil
	}

	payload := getPayloadByServiceModels(env, service_models, service_weight, service_config)
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal() err: %v", err)
	}
	if env.Conf.Log.Level == "debug" {
		requestBodyPretty, _ := json.MarshalIndent(payload, "", "  ")
		logger.Debugf("\n>> Request: %v\n\n", string(requestBodyPretty))
	}

	// send post request to notify predictor
	var url = "http://" + predictor_ip + ":" + predictor_http_port + "/load_and_register"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("http POST err: %v, url: %v", err, url)
	}

	// parse http response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("parse http response err: %v, url: %v", err, url)
	}
	var respStruct struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	err = json.Unmarshal(body, &respStruct)
	if err != nil {
		return fmt.Errorf("Unmarshal http response err: %v, url: %v ,body : %s", err, url, string(body))
	}
	if respStruct.Code != 0 {
		return fmt.Errorf("loadAndRegister response fail , url: %v ,respStruct : %+v", url, respStruct)
	}

	logger.Debugf("\n>> Response: %v\n\n", string(body))
	return nil
}

// split service_models map into slice
func SplitServiceModelsMap(service_models map[string]map[string]schema.ModelHistory) []map[string]map[string]schema.ModelHistory {
	var service_model_map_slice []map[string]map[string]schema.ModelHistory
	for service := range service_models {
		for model, his := range service_models[service] {
		  model_his := map[string]schema.ModelHistory{model:his}
			service_model := map[string]map[string]schema.ModelHistory{service:model_his}
			service_model_map_slice = append(service_model_map_slice, service_model)
		}
	}
	return service_model_map_slice
}

// 根据ModelHistory数据组装请求predictor接口的payload
func getPayloadByServiceModels(env *env.Env,
	service_models map[string]map[string]schema.ModelHistory,
	service_weight map[string]int,
	service_config map[string]string,
) api.PredictorPayload {
	var payload = api.PredictorPayload{
		[]api.PredictorService{},
	}
	for service_name, model_histories := range service_models {
		var service = api.PredictorService{}
		service.ServiceName = service_name
		service.ServiceWeight = service_weight[service_name]
		service.ModelRecords = []api.PredictorModelRecord{}
		for _, model_history := range model_histories {
			if len(model_history.ModelName) == 0 || len(model_history.Timestamp) == 0 {
				logger.Infof("skipping invalid model_history=%v", model_history)
				continue
			}
			service.ModelRecords = append(service.ModelRecords, model_history.ToPredictorModelRecord(env))
		}
		config_json, exists := service_config[service_name]
		if exists {
			var config_map map[string]interface{}
			buf := []byte(config_json)
			if err := json.Unmarshal(buf, &config_map); err != nil {
				logger.Errorf("parse config for service[%v] err: [%v] json: [%v]", service_name, err, config_json)
			} else {
				service.ServiceConfig = config_map
			}
		}
		payload.Services = append(payload.Services, service)
	}
	return payload
}

// return map[service_name]load_weight
func GetServiceWeight(env *env.Env, ip string) (map[string]int, error) {
	db := env.Db
	// return value
	service_weight := make(map[string]int)

	// search for responsible service names and service load_weight
	sql_query := `select services.name,host_services.load_weight 
		from hosts join host_services join services 
		where hosts.id = host_services.hid and host_services.sid = services.id and hosts.ip = ?`
	rows, err := db.Raw(sql_query, ip).Rows()
	logger.Debugf("\n>> sql_query=\"%v, ? = %v\"\n", sql_query, ip)
	if err != nil {
		return service_weight, fmt.Errorf("gorm db err: sql=%v err=%v", sql_query, err.Error())
	}
	for rows.Next() {
		var service_name string
		var load_weight int
		if err := rows.Scan(&service_name, &load_weight); err != nil {
			logger.Errorf("rows.Scan err: %v", err)
			continue
		}
		service_weight[service_name] = load_weight
	}
	return service_weight, nil
}

// return map[service_name]config
func GetServiceConfig(env *env.Env, service_names []string) (map[string]string, error) {
	db := env.Db
	// return value
	service_config := make(map[string]string)

	// search for service names and config
	sql_query := `select services.name, configs.config 
		from services join service_configs join configs
		where services.id = service_configs.sid and service_configs.cid = configs.id and services.name in (?)`
	rows, err := db.Raw(sql_query, service_names).Rows()
	logger.Debugf("\n>> sql_query=\"%v, ? = %v\"\n", sql_query, service_names)
	if err != nil {
		return service_config, fmt.Errorf("gorm db err: sql=%v err=%v", sql_query, err.Error())
	}
	for rows.Next() {
		var service_name string
		var config string
		if err := rows.Scan(&service_name, &config); err != nil {
			logger.Errorf("rows.Scan err: %v", err)
			continue
		}
		service_config[service_name] = config
	}
	return service_config, nil
}

func NotifyPredictorStressInfo(
	// format data into request json
	model_names string, qps string,
	env *env.Env,
	predictor_ip string,
	predictor_http_port string,
) error {
	var payload = api.PredictorStressInfoPayload {
		ModelNames: model_names,
		Qps: qps,
		Service: env.Conf.StressTestService.ServiceName,
	}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal() err: %v", err)
	}
	if env.Conf.Log.Level == "debug" {
		requestBodyPretty, _ := json.MarshalIndent(payload, "", "  ")
		logger.Debugf("\n>> Request: %v\n\n", string(requestBodyPretty))
	}

	// send post request to notify predictor
	var url = "http://" + predictor_ip + ":" + predictor_http_port + "/set_stress_params"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("http POST err: %v, url: %v", err, url)
	}

	// parse http response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("parse http response err: %v, url: %v", err, url)
	}
	var respStruct struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	err = json.Unmarshal(body, &respStruct)
	if err != nil {
		return fmt.Errorf("Unmarshal http response err: %v, url: %v ,body : %s", err, url, string(body))
	}
	if respStruct.Code != 0 {
		return fmt.Errorf("set_stress_params response fail , url: %v ,respStruct : %+v", url, respStruct)
	}

	logger.Debugf("\n>> Response: %v\n\n", string(body))
	return nil
}

// 获取当前host所属的service
func FetchDbServices(env *env.Env) ([]*schema.Service, error) {
	db := env.Db
	var dbServices []*schema.Service
	sql := `SELECT s.id,s.name FROM host_services hs 
		LEFT JOIN services s ON s.id = hs.sid
		LEFT JOIN hosts h ON h.id = hs.hid 
		WHERE h.ip = ?`
	dbPtr := db.Raw(sql, env.LocalIp).Find(&dbServices)
	logger.Debugf(">> sql=\"%s, ? = %s\"", sql, env.LocalIp)
	if errs := dbPtr.GetErrors(); len(errs) > 0 {
		return []*schema.Service{}, fmt.Errorf("gorm db err: sql=%v err=%v", sql, errs)
	}
	return dbServices, nil
}

// 机器列表划分AB段，只取相同AB段机器列表，暂时分为两层
// return parentIp, peerNum, error
func GetParentIP(env *env.Env, dbService *schema.Service, serviceSyncFileTimesMap sync.Map, syncFileLimit int, defaultSrcHost string) (string, int, error) {
	// validate service, parentIP is transmit
	if env.Conf.ValidateService.Host == env.LocalIp {
		return defaultSrcHost, 0, nil
	}
	// 容错，判断如果获取超过拉取限制次数恢复使用中转机拉取
	var currentTimes int32
	if timesI, exists := serviceSyncFileTimesMap.Load(dbService.Name); exists {
		currentTimes, _ = timesI.(int32)
	}
	if int(currentTimes) >= syncFileLimit {
		return defaultSrcHost, 0, nil
	}
	// 获取service下跟当前ip AB段相同的ip列表
	ipFields := strings.Split(env.LocalIp, ".")
	prefixIp := ipFields[0] + "." + ipFields[1] + "."
	abIps, err := FetchDbHostsByService(env, dbService.ID, prefixIp)
	if err != nil {
		return "", 0, fmt.Errorf("fetchDbHostsByService err: %v, sid: %d, service_name: %s, prefixIp: %s",
			err, dbService.ID, dbService.Name, prefixIp)
	}
	if len(abIps) == 0 {
		err := fmt.Errorf("not found abIps in service, sid: %d, service_name: %s, prefixIp: %s",
			dbService.ID, dbService.Name, prefixIp)
		logger.Errorf("getParentIP err: %v", err)
		return "", 0, err
	}
	logger.Debugf("abIps: %v, service_name: %s", abIps, dbService.Name)
	// 根据ab段相同ip列表获取parentIP
	parentIp, peerNum, err := GetParentIpFromABIps(env, abIps)
	return parentIp, peerNum, err
}

// 根据ab段相同ip列表获取parentIP
func GetParentIpFromABIps(env *env.Env, abIps []string) (string, int, error) {
	if len(abIps) == 0 {
		return "", 0, fmt.Errorf("abIps is empty")
	}

	// 1.如果未达到子节点限制数量，取第一个节点为parent
	if len(abIps)-1 <= env.Conf.P2PModelService.PeerLimit {
		if abIps[0] == env.LocalIp {
			return env.Conf.P2PModelService.SrcHost, 0, nil
		} else {
			logger.Debugf("parentIp: %s, peers: %v", abIps[0], abIps[1:])
			return abIps[0], len(abIps) - 1, nil
		}
	}
	// 2.如果大于子节点限制数量，平均分为多个parent和多个peer
	parentSize := int(math.Ceil(float64(len(abIps)) / float64(env.Conf.P2PModelService.PeerLimit+1)))
	parentIps := abIps[:parentSize]
	lastPeerIps := abIps[parentSize:]
	// 平均切割slice
	peerIpsList := common.DivideSlices(lastPeerIps, parentSize)
	exists, parentIndex := common.GetIndexFromSlice(env.LocalIp, parentIps)
	// 如果是父节点，从中转机获取
	if exists {
		return env.Conf.P2PModelService.SrcHost, 0, nil
	}
	// 找到本机节点对应parentIps的index
	exists, parentIndex = common.GetIndexFromChildList(env.LocalIp, peerIpsList)
	if !exists {
		return "", 0, fmt.Errorf("localIp is not in abIps")
	} else {
		logger.Debugf("parentIp: %s, peers: %v", parentIps[parentIndex], peerIpsList[parentIndex])
		return parentIps[parentIndex], len(peerIpsList[parentIndex]), nil
	}
}

// 获取当前service下的所有hosts
func FetchDbHostsByService(env *env.Env, sid uint, prefixIp string) ([]string, error) {
	db := env.Db
	hosts := []string{}
	sql := `SELECT h.ip FROM host_services hs 
		LEFT JOIN hosts h ON h.id = hs.hid
		WHERE hs.sid = ?`
	var andLikeSql string
	if prefixIp != "" {
		andLikeSql = fmt.Sprintf(` AND h.ip like '%s%%'`, prefixIp)
	}
	orderSql := ` ORDER BY hs.id ASC`
	sql = sql + andLikeSql + orderSql
	dbPtr := db.Raw(sql, sid).Pluck("h.ip", &hosts)
	logger.Debugf(">> sql=\"%s, ?= %d \"", sql, sid)
	if errs := dbPtr.GetErrors(); len(errs) > 0 {
		return hosts, fmt.Errorf("gorm db err: sql=%s err=%v", sql, errs)
	}
	return hosts, nil
}
