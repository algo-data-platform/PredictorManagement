package logics

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"math"
	"net/http"
	"server/conf"
	"server/env"
	"server/libs/logger"
	"server/schema"
	"server/server/dao"
	"server/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 获取模型版本loaded状态的百分比
// @return map[timestamp]percent
func GetModelVersionStatePercent(modelName string) (map[string]int, error) {
	var versionPercentMap = make(map[string]int)
	// 排除一致性验证机器
	consistenceService, err := dao.GetServiceByName(env.Env.Conf.ConsistenceService)
	if err != nil || consistenceService.ID == 0 {
		logger.Errorf("GetServiceByName fail, err: %v", err)
		return versionPercentMap, err
	}
	hostList, err := dao.GetHostByModel(modelName, consistenceService.ID)
	if err != nil {
		logger.Errorf("GetHostByModel fail, modelName:%s, err: %v", modelName, err)
		return versionPercentMap, err
	}
	if len(hostList) > 0 {
		var stateChan = make(chan *util.VersionStatus, len(hostList))
		wg_ := sync.WaitGroup{}
		wg_.Add(len(hostList))
		for _, host := range hostList {
			// 协程获取线上每个机器模型版本状态
			go getHostModelStatus(host.Ip, modelName, &wg_, stateChan)
		}
		go func() {
			wg_.Wait()
			close(stateChan)
		}()

		// 统计
		var loadedCountMap = make(map[string]int)
		for versionState := range stateChan {
			if versionState.State == "loaded" {
				loadedCountMap[versionState.Timestamp]++
			}
		}
		// 计算load百分比
		for timestamp, count := range loadedCountMap {
			versionPercentMap[timestamp] = int(math.Floor(float64(count) / float64(len(hostList)) * float64(100)))
		}
	}
	return versionPercentMap, nil
}

func getHostModelStatus(ip string, modelName string, wg_ *sync.WaitGroup, stateChan chan *util.VersionStatus) {
	defer wg_.Done()
	versionStatus := &util.VersionStatus{}
	model_info_meta, err := getNodeInfoByHost(ip)
	if err != nil {
		stateChan <- versionStatus
		return
	}
	model_info_service := model_info_meta.Msg.Services
	for _, service := range model_info_service {
		for _, model_record := range service.ModelRecords {
			if model_record.Name == modelName {
				versionStatus.Timestamp = model_record.Timestamp
				versionStatus.State = model_record.State
				break
			}
		}
	}
	stateChan <- versionStatus
}

func getNodeInfoByHost(host string) (*util.NodeMetaInfo, error) {
	model_info_url := fmt.Sprintf("http://%s:%d/get_service_model_info", host, env.Env.Conf.PredictorHttpPort)
	model_info_resp, err := util.GetContentFromUrl(model_info_url, time.Millisecond*time.Duration(env.Env.Conf.HttpTimeout), env.Env.Conf.HttpRetryConn)
	if err != nil {
		return nil, err
	}
	var model_info_meta = &util.NodeMetaInfo{}
	if err_ := json.Unmarshal(model_info_resp, model_info_meta); err_ != nil {
		logger.Errorf("json unmarshal model info error: %v", err_)
		return nil, fmt.Errorf("json unmarshal model info error: %v", err_)
	}
	return model_info_meta, nil
}

//add get model all info
func GetModelInfo(model_name string, show_range string, conf *conf.Conf) []util.ModelHistoryInfo {
	var model_history_info []util.ModelHistoryInfo
	var model_history []schema.ModelHistory
	var db_ptr *gorm.DB
	if show_range == "all" {
		db_ptr = env.Env.MysqlDB.Where("model_name = ? ", model_name).Find(&model_history)
	} else {
		request_model_num, err := strconv.Atoi(show_range)
		if err != nil {
			logger.Errorf("strconv error: %v, so default num is 5", err)
			request_model_num = 5
		}
		db_ptr = env.Env.MysqlDB.Order("timestamp desc").Limit(request_model_num).Where("model_name = ?", model_name).Find(&model_history)
	}
	if db_ptr == nil {
		logger.Errorf("mysql db ptr nil, please check")
		return model_history_info
	}
	if db_ptr.RecordNotFound() {
		logger.Errorf("model_histories table not found")
		return model_history_info
	} else if err := db_ptr.GetErrors(); len(err) != 0 {
		logger.Errorf("get data from table model_histories error: %v", err)
		return model_history_info
	}

	// 获取每个版本下loaded 状态所占有的比例（0-100）
	versionPercentMap, err := GetModelVersionStatePercent(model_name)
	if err != nil {
		logger.Errorf("GetModelVersionStatePercent error: %v", err)
		return model_history_info
	}
	model_history_num := len(model_history)
	for i := 0; i < model_history_num; i++ {
		var model_info_ util.ModelHistoryInfo
		cur_model_info := model_history[i]
		model_info_.ModelName = cur_model_info.ModelName
		model_info_.Desc = cur_model_info.Desc
		model_info_.Timestamp = cur_model_info.Timestamp
		model_info_.IsLocked = cur_model_info.IsLocked
		model_info_.Md5 = cur_model_info.Md5
		model_info_.CreatedAt = cur_model_info.CreatedAt.Format("2006-01-02 15:04:05")
		model_info_.UpdatedAt = cur_model_info.UpdatedAt.Format("2006-01-02 15:04:05")
		if _, exists := versionPercentMap[model_info_.Timestamp]; exists {
			model_info_.Status = "loaded"
			model_info_.Percent = versionPercentMap[model_info_.Timestamp]
		}
		model_history_info = append(model_history_info, model_info_)
	}
	return model_history_info
}

func GetModelsMailRecipients(model_list string) map[string]string {
	models_mail_recipients := make(map[string]string)
	model_name_list := strings.Split(model_list, ",")
	for _, model_name := range model_name_list {
		var model schema.Model
		db_ptr := env.Env.MysqlDB.Where("name = ? ", model_name).Find(&model)
		if db_ptr == nil {
			logger.Errorf("mysql db ptr nil, please check")
			return models_mail_recipients
		}
		if db_ptr.RecordNotFound() {
			logger.Errorf("models table not found")
			return models_mail_recipients
		} else if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table models error: %v", err)
			return models_mail_recipients
		}
		var extension_map map[string]interface{}
		json.NewDecoder(strings.NewReader(model.Extension)).Decode(&extension_map)
		value, ok := extension_map["MailRecipients"].([]interface{})
		if !ok {
			logger.Warnf("It's not ok for type []interface{}")
			continue
		}
		mail_recipients := ""
		for _, recipient := range value {
			mail_recipients += recipient.(string)
			mail_recipients += ","
		}
		mail_recipients = strings.TrimRight(mail_recipients, ",")
		models_mail_recipients[model_name] = mail_recipients
	}
	return models_mail_recipients
}

func SetModelMailRecipients(model_name string, mail_list string, conf *conf.Conf) {
	mail_list_array := strings.Split(mail_list, ",")
	mail_recipients_map := make(map[string]interface{})
	var mail_recipients []string
	for _, mail_address := range mail_list_array {
		mail_address := strings.TrimSpace(mail_address)
		if len(mail_address) == 0 {
			continue
		}
		mail_recipients = append(mail_recipients, mail_address)
	}
	mail_recipients_map["MailRecipients"] = mail_recipients
	json_byte, err := json.Marshal(mail_recipients_map)
	if err != nil {
		logger.Errorf("MapToJson err: ", err)
		return
	}
	mail_recipients_json_str := string(json_byte)
	logger.Infof(mail_recipients_json_str)
	host_port := (conf.HttpHost) + ":" + strconv.Itoa(conf.HttpPort)
	var url = "http://" + host_port + "/mysql/update?table=models&name=" + model_name + "&extension_update=" + mail_recipients_json_str
	_, err_http := http.Get(url)
	if err_http != nil {
		logger.Errorf(fmt.Sprintf("Http get request to update model extension failed! err: %v", err_http))
	}
}
