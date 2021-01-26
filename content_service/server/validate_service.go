package server

import (
  "sync"
  "os"
  "os/exec"
  "fmt"
  "time"
  "strings"
  "errors"
  "bufio"
  "io"
  "io/ioutil"
  "math"
  "net/http"
  "html/template"
  "bytes"
  "encoding/json"
  "content_service/env"
  "content_service/common"
  "content_service/api"
  "content_service/schema"
  "strconv"
  "path"
  "sort"
  "github.com/algo-data-platform/predictor/golibs/adgo/feature_master/if/feature_master"
  "github.com/algo-data-platform/predictor/golibs/adgo/predictor_client"
  "github.com/algo-data-platform/predictor/golibs/adgo/predictor_client/if/predictor"
  "content_service/libs/logger"
  "github.com/robfig/cron"
)

type ValidateService struct {
}

var validateInstance *ValidateService
var validateOnce sync.Once

func GetValidateInstance() *ValidateService {
  validateOnce.Do(func() {
    validateInstance = &ValidateService{}
  })
  return validateInstance
}

const (
  _ = iota
  String
  Float
  Int64
)

const MaxReportDataSize = 100

func (service *ValidateService) Start(env *env.Env){
  logger.Infof("LocalIp=%v: Validate Service cron start...", env.LocalIp)
  service.startCronForValidateDailyReport(env)
}

func (service *ValidateService) RunValidateService(db_service_models map[string]map[string]schema.ModelHistory, env *env.Env) {
	for service_name, db_models := range db_service_models {
		if service_name != env.Conf.ValidateService.ServiceName || len(db_service_models[service_name]) == 0 {
			continue
		}
		logger.Infof("\n>> Models to validate: %v\n\n", common.Pretty(db_models))
		for _, db_model := range db_models {
			service.Run(db_model.ModelName, db_model.Timestamp, env, env.Conf.P2PModelService.DestPath)
		}
	}
}

func (service *ValidateService) startCronForValidateDailyReport(env *env.Env) {
	c := cron.New()
	c.AddFunc("0 0 8 * * *", func() {
		//get the formatted date of yesterday
		date := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		service.GenerateAndSendDailyReport(path.Join(env.Conf.ValidateService.SummaryResultDir, "summary_result_"+date), date, env)
	})
	c.Start()
}

func (service *ValidateService) Run(model_name string, timestamp string, env *env.Env, model_path string) {
  model_full_name := model_name + common.Delimiter + timestamp
  sample_file_path := path.Join(model_path, model_full_name, "sample_file.txt")
  logger.Infof("LocalIp=%v: Validate Service running...\n", env.LocalIp)
// CheckModelLoadStatusWithRetry
  err_checkload := service.CheckModelLoadStatusWithRetry(model_name, timestamp, env)
  if err_checkload != nil {
    logger.Errorf("CheckModelLoadStatusWithRetry failed! err: %v", err_checkload)
  }

//check if there is sample file supplied if not then return and load the model online 
  _, err_file_exists := os.Stat(sample_file_path)
  if err_file_exists == nil && err_checkload != nil {
    // send mail if load model failed
    service.ConcludeAndSendBaseValidateReport(model_name, timestamp, env, "样本验证", err_checkload)
    return
  } else if err_file_exists != nil {
    // send mail if sample_file.txt is not exist
    err := &common.TagError{fmt.Sprintf("%v", err_file_exists), "未发现样本文件"}
    service.ConcludeAndSendBaseValidateReport(model_name, timestamp, env, "样本验证", err)
    return
  }

// GetModelFeature
  feature_map, err_getFeature := service.GetModelFeature(model_name, timestamp, env, model_path)
  if err_getFeature != nil {
    logger.Errorf("GetModelFeature failed! err: %v", err_getFeature)
    service.ConcludeAndSendBaseValidateReport(model_name, timestamp, env, "样本验证", err_getFeature)
    //send alert to dingding
    err_str := fmt.Sprintf("[%v][Validate Service] GetModelFeature failed for: %v", env.LocalIp, model_full_name)
    common.DingDingAlert(env.Conf.DingDingWebhookUrl, err_str)
    return
  }

  if !strings.Contains(model_name, "retrieval") && !strings.Contains(model_name, "crank") {
    // get request from sample file
    requests := predictor.NewPredictRequests()
    reqid_offlinectr_map := make(map[int]float64)
    invalid_samples := make(map[int]string)
    err_getRequests := service.GetRequestsFromSampleFile(model_name, timestamp, sample_file_path, feature_map, env, requests, reqid_offlinectr_map, invalid_samples)
    if err_getRequests != nil {
      logger.Errorf("GetRequestsFromSampleFile failed! err: %v", err_getRequests)
      service.ConcludeAndSendBaseValidateReport(model_name, timestamp, env, "样本验证", err_getRequests)
      return
    }
    
    // init predictor client and predict
    responses, err_predict := service.Predict(requests, env)
    if err_predict != nil {
      logger.Errorf("Predict failed! err: %v", err_predict)
      err := &common.TagError{fmt.Sprintf("%v", err_predict), "调用Predictor失败"}
      service.ConcludeAndSendBaseValidateReport(model_name, timestamp, env, "样本验证", err)
      err_str := fmt.Sprintf("[%v][Validate Service] Call Predictor failed for: %v", env.LocalIp, model_full_name)
      common.DingDingAlert(env.Conf.DingDingWebhookUrl, err_str)
      return
    }

    // generate validate report and send
    service.ConcludeAndSendSampleValidateReport(responses, reqid_offlinectr_map, invalid_samples, model_name, timestamp, env)
  } else {
    // GetCalcVectorRequestsFromSampleFile
    requests := predictor.NewCalculateVectorRequests()
    reqid_offline_output_map := make(map[int]map[string][]float64)
    invalid_samples := make(map[int]string)
    var output_names []string
    err_getRequests := service.GetCalcVectorRequestsFromSampleFile(requests, model_name, timestamp, sample_file_path, env, reqid_offline_output_map, invalid_samples, &output_names, feature_map)
    if err_getRequests != nil {
      logger.Errorf("GetCalcVectorRequestsFromSampleFile failed! err: %v", err_getRequests)
      service.ConcludeAndSendBaseValidateReport(model_name, timestamp, env, "样本验证", err_getRequests)
      return
    }

    // CalculateVector
    responses, err_calcVector := service.CalculateVector(requests, env)
    if err_calcVector != nil {
      logger.Errorf("CalculateVector failed! err: %v", err_calcVector)
      err := &common.TagError{fmt.Sprintf("%v", err_calcVector), "调用Predictor失败"}
      service.ConcludeAndSendBaseValidateReport(model_name, timestamp, env, "样本验证", err)
      err_str := fmt.Sprintf("[%v][Validate Service] Call Predictor failed for: %v", env.LocalIp, model_full_name)
      common.DingDingAlert(env.Conf.DingDingWebhookUrl, err_str)
      return
    }

    // GetComparisonResult
    fail_reasons_map := make(map[int]string)
    online_offline_result_map := make(map[int]map[string]string)
    service.GetComparisonResult(responses, reqid_offline_output_map, output_names, fail_reasons_map, online_offline_result_map);

    // GenerateAndSendCalcVectorSampleValidateReport
    service.GenerateAndSendCalcVectorSampleValidateReport(online_offline_result_map, fail_reasons_map, invalid_samples, output_names, model_name, timestamp, env)
  }
}

func (service *ValidateService) CheckModelLoadStatusWithRetry(model_name string, timestamp string,env *env.Env) *common.TagError {
  for i:=1; i<=env.Conf.ValidateService.RetryTimes; i++ {
    err_checkload := service.CheckModelLoadStatus(model_name, timestamp, env)
    if err_checkload == nil {
      logger.Infof("Check model load status success")
      break
    } else if i==env.Conf.ValidateService.RetryTimes {
      err_str := fmt.Sprintf("Check model load status out of time, retry times:%d, retry interval:%ds", env.Conf.ValidateService.RetryTimes, env.Conf.ValidateService.RetryInterval)
      return &common.TagError{err_str, "模型加载失败"}
    } else {
      logger.Errorf("Check model load status failed, will retry... err: %v", err_checkload)
      time.Sleep(time.Second * env.Conf.ValidateService.RetryInterval) //sleep for 1s * env.Conf.ValidateService.RetryInterval
    }
  }
  return nil
}

func (service *ValidateService) CheckModelLoadStatus(model_name string, timestamp string,env *env.Env) error {
  model_full_name := model_name + common.Delimiter + timestamp
  logger.Infof("The model_name is: " + model_full_name)
  var url = "http://" + env.LocalIp + ":" + env.Conf.P2PModelService.TargetService.HttpPort + "/get_service_model_info"

  resp, err_httpGet := http.Get(url)
  if err_httpGet != nil {
    err_str := fmt.Sprintf("Http get request to predictor failed! err: %v", err_httpGet)
    return errors.New(err_str)
  }
  defer resp.Body.Close()
  logger.Debugf("Http get response from predictor: %v", resp)
  body, _ := ioutil.ReadAll(resp.Body)
  model_service := new(api.ModelServiceInfo)
  err_unmarshal := json.Unmarshal(body, model_service)
  if err_unmarshal == nil {
    for _, service := range (*model_service).Msg.Services {
      if service.ServiceName != env.Conf.ValidateService.ServiceName {
        continue
      }
      for _, model_record := range service.ModelRecords {
        if model_record.FullName == model_full_name && model_record.State == "loaded" {
          logger.Infof("Model: " + model_record.FullName + " is loaded successfully!")
          return nil
        }
      }
    }
    return errors.New("Model: " + model_full_name + " is not loaded!")
  } 

  err_str := fmt.Sprintf("Unmarshal the json of ModelServiceInfo failed! err: %v", err_unmarshal)
  return errors.New(err_str)
}

func (service *ValidateService) GetModelFeature(model_name string, timestamp string, env *env.Env, model_path string) (map[string]*api.ModelFeature, *common.TagError) {
  model_feature_map := make(map[string]*api.ModelFeature)
  if strings.Contains(model_name, "catboost") {
    model_feature_map["float_features"] = new(api.ModelFeature)
    model_feature_map["float_features"].Type = 2  // Float
    model_feature_map["float_features"].Dim = "item"
    model_feature_map["cat_features"] = new(api.ModelFeature)
    model_feature_map["cat_features"].Type = 1  // String
    model_feature_map["cat_features"].Dim = "item"
  }
  model_full_name := model_name + common.Delimiter + timestamp
  meta_info_file_path := path.Join(model_path, model_full_name, "meta_info.json")
  _, err_file_exists := os.Stat(meta_info_file_path)
  if err_file_exists == nil {
    if err_get_feature_info := GetFeatureInfoFromMetaInfoFile(model_feature_map, meta_info_file_path); err_get_feature_info != nil {
      return nil, err_get_feature_info
    }
  }
  return model_feature_map, nil
}

func (service *ValidateService) GetRequestsFromSampleFile(model_name string, timestamp string, sample_file string, feature_map map[string]*api.ModelFeature, env *env.Env, requests *(predictor.PredictRequests), reqid_offlinectr_map map[int]float64, invalid_samples map[int]string) (*common.TagError) {
  logger.Infof("Open sample file: " + sample_file)
  fp, err_file := os.Open(sample_file)
  if err_file != nil {
    err_str := fmt.Sprintf("Open sample file failed! err: %v", err_file)
    return &common.TagError{err_str, "打开样本文件失败"}
  }
  defer fp.Close()
  r := bufio.NewReader(fp)
  var line_num int
  var keys []string
  // feature list provided both in sample files and feature service platform
  var common_features_list []string
  var item_features_list []string

  //seperator of sample file
  var seperator string
  for {
    if line_num > env.Conf.ValidateService.MaxSampleCount {
      break
    }
    buf, err := r.ReadString('\n')
    if err == io.EOF || err != nil { //-1
        break
    }
    buf = strings.Replace(buf, "\n", "", -1)
    if(buf == "") {
      continue
    }
    if line_num == 0 {
      buf = strings.Replace(buf, " ", "", -1)
      if strings.Count(buf, ",") > 0 {
        seperator = ","
      } else if strings.Count(buf, ";") > 0 {
        seperator = ";"
      } else if strings.Count(buf, "\t") > 0 {
        seperator = "\t"
      } else {
        err_str := "no seperator found in the first line of sample file"
        return &common.TagError{err_str, "样本文件首行未发现分隔符"}
      }
      keys = strings.Split(buf, seperator)
      header_exists := false
      for _, key := range keys {
        if (key == "offline_ctr") {
          header_exists = true
          continue
        }
        feature_attr, ok := feature_map[key]
        if ok {
          switch feature_attr.Dim {
            case "common":
              common_features_list = append(common_features_list, key)
            case "item":
              item_features_list = append(item_features_list, key)
          }
        } else {
          item_features_list = append(item_features_list, key)
        }
      }
      if (!header_exists) {
        err_str := "no header found in the first line of sample file"
        return &common.TagError{err_str, "样本文件首行未发现offline_ctr字段名"}
      }
      logger.Infof("There are %v common features provided: %v", len(common_features_list), common_features_list)
      logger.Infof("There are %v item features provided: %v", len(item_features_list), item_features_list)
    } else {
      request := predictor.NewPredictRequest()
      request.ReqID = strconv.Itoa(line_num) + common.Delimiter + model_name + common.Delimiter + timestamp
      request.Channel = "test"
      request.ModelName = model_name
      values := strings.Split(buf, seperator)
      if len(values) != len(keys) {
        invalid_samples[line_num] = "样本字段数与特征数不一致"
        line_num += 1
        continue
      }
      common_features := feature_master.NewFeatures()
      item_features := feature_master.NewFeatures()
      var offlinectr float64
      var err_parsefloat error
    J:
      for i, value := range values {
        value = strings.Trim(value, " ")
        if(keys[i] == "offline_ctr") {
          offlinectr, err_parsefloat = strconv.ParseFloat(value, 64)
          if(err_parsefloat != nil) {
            err_str := fmt.Sprintf("Data parse err: %v", err_parsefloat)
            logger.Errorf("Invalid offline_ctr data! " + err_str)
            break J
          }
        }
        feature_name := keys[i]
        feature_attr, ok := feature_map[feature_name]
        if ok {
          feature, err := MakeFeature(feature_name, value, feature_attr)
          if err != nil {
            break J
          }
          switch feature_attr.Dim {
            case "common":
              common_features.Features = append(common_features.Features, feature)
            case "item":
              item_features.Features = append(item_features.Features, feature)
          }
        } else if feature_name != "offline_ctr" {
          feature := feature_master.NewFeature()
          feature.FeatureName = feature_name
          feature.StringValues = append(feature.StringValues, value)
          feature.FeatureType = feature_master.FeatureType_STRING_LIST
          item_features.Features = append(item_features.Features, feature)
        }
      }
      if (len(common_features.Features)+len(item_features.Features)) == (len(common_features_list)+len(item_features_list)) {
        reqid_offlinectr_map[line_num] = offlinectr
        request.CommonFeatures = common_features
        request.ItemFeatures = make(map[int64]*feature_master.Features)
        request.ItemFeatures[int64(line_num)] = item_features
        requests.Reqs = append(requests.Reqs, request)
      } else {
        logger.Errorf("The sample %v is invalid!", line_num)
        invalid_samples[line_num] = "样本数据类型不正确"
      }
    }
    line_num += 1
  }
  if line_num == 0 {
    err_str := "no data found in the sample file"
    return &common.TagError{err_str, "样本文件为空"}
  } else if line_num == 1 {
    err_str := "no sample found in the sample file"
    return &common.TagError{err_str, "样本文件未发现样本"}
  }
  logger.Infof("There are " + strconv.Itoa(len(requests.Reqs)) + " samples provided")
  return nil
}

func (service *ValidateService) Predict(requests *(predictor.PredictRequests), env *env.Env) (*(predictor.PredictResponses), error) {
  predictorClient, err_predictClient := predictor_client.NewPredictorClient(env.Conf.ValidateService.ConsulAddress, env.Conf.ValidateService.ServiceName, env.LocalIp)
  if err_predictClient != nil {
      logger.Errorf("Failed to init predictor client! err: %v", err_predictClient)
      return nil, err_predictClient
  }
  predictorClient.SetTimeout(time.Duration(env.Conf.ValidateService.PredictorTimeout) * time.Second)
  responses := predictor.NewPredictResponses()
  var err_predict error
  for i:=1; i<=env.Conf.ValidateService.RetryTimes; i++ {
    responses, err_predict = predictorClient.Predict(requests)
    if err_predict == nil {
      logger.Infof("There are " + strconv.Itoa(len(responses.Resps)) + " results")
      break
    } else if i==env.Conf.ValidateService.RetryTimes {
      err_str := fmt.Sprintf("Predict out of time, retry times:%d, retry interval:1s", env.Conf.ValidateService.RetryTimes)
      return nil, errors.New(err_str)
    } else {
      logger.Errorf("Predict failed, will retry... err: %v", err_predict)
      time.Sleep(time.Second) //sleep for 1s
    }
  }
  return responses, nil
}

func (service *ValidateService) ConcludeAndSendBaseValidateReport(model_name string, timestamp string, env *env.Env, validate_type string, err *common.TagError) {
  report_map := make(map[string]interface{})
  report_map["ModelName"] = model_name
  report_map["ModelVersion"] = timestamp
  report_map["ValidateType"] = validate_type
  if err != nil {
    service.UpdateDatabaseModelDesc(model_name, timestamp, "Abandoned", env)
    report_map["ValidateConclusion"] = "不通过，" + err.ErrTag
  } else {
    err_update_database := service.UpdateDatabaseModelDesc(model_name, timestamp, "Validated", env)
    if err_update_database != nil {
      report_map["ValidateConclusion"] = "不通过，更新数据库失败"
    } else {
      report_map["ValidateConclusion"] = "通过"
    }
  }
  service.SaveSummaryResult(fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", model_name, timestamp, report_map["ValidateType"], report_map["ValidateConclusion"], time.Now().Format("2006-01-02 15:04:05")), env)
  if report_map["ValidateConclusion"] == "通过" {
    logger.Infof("Model: %v-%v is validated successfully", model_name, timestamp)
    return
  }
  // 如果是regression mode，屏蔽每个模型验证失败之后的邮件发送
  if env.IsRegressionMode() {
    return
  }
  log_file_url, log_str, err_log := service.GetLogInfo(model_name, timestamp, env) 
  if err_log != nil {
    //send dingding alert
    err_str := fmt.Sprintf("[%v][Validate Service] GetAlgoLogInfo failed, err:%v", env.LocalIp, err_log)
    common.DingDingAlert(env.Conf.DingDingWebhookUrl, err_str)
  }
  report_map["AlgoLogUrl"] = log_file_url
  report_map["AlgoLog"] = log_str
  report_recipients, claim_status := service.GetReportRecipientsByModel(model_name, env, report_map["ValidateConclusion"].(string))
  report_map["ClaimStatus"] = claim_status
  err_sendReport := service.SendHtmlReport(report_map, report_recipients, "一致性验证报告"+claim_status, path.Join(env.Conf.ValidateService.HtmlTemplateDir,"validate_report.tpl"), env)
  if err_sendReport != nil {
    logger.Errorf("%v", err_sendReport)
  }
  
}

func (service *ValidateService) ConcludeAndSendSampleValidateReport(responses *(predictor.PredictResponses), reqid_offlinectr_map map[int]float64, invalid_samples map[int]string, model_name string, timestamp string, env *env.Env) {
  reqid_pctr_map := make(map[int]float64)
  report_map := make(map[string]interface{})
  report_map["ModelName"] = model_name
  report_map["ModelVersion"] = timestamp
  report_map["ValidateType"] = "样本验证"
  report_map["Header"] = []string{"样本ID", "离线CTR", "线上CTR", "验证结果"}
  var report_data []([]interface{})
  var sample_pass_count, sample_fail_count int64
  has_resp_map := make(map[int]int)
  for _, resp := range responses.Resps {
    values := strings.Split(resp.ReqID, common.Delimiter);
    reqid, _ := strconv.Atoi(values[0])
    if resp.ResultsMap[int64(reqid)] == nil {
      continue
    }
    has_resp_map[reqid] = 1;
    predict_ctr := resp.ResultsMap[int64(reqid)].Preds["ctr"]
    reqid_pctr_map[reqid] = predict_ctr
    if math.Abs(reqid_pctr_map[reqid]*100000000-reqid_offlinectr_map[reqid]*100000000) < 120 {
      logger.Infof("The Request ID: " + resp.ReqID + " 一致性验证通过")
      sample_pass_count += 1
    } else {
      logger.Infof("The Request ID: " + resp.ReqID + " 一致性验证失败")
      logger.Infof("The Request ID: " + resp.ReqID + "    ========>    The Predict CTR: " + strconv.FormatFloat(predict_ctr, 'f', 8, 64) + "    ========>    The Offline CTR: " + strconv.FormatFloat(reqid_offlinectr_map[reqid], 'f', 8, 64))
      data := []interface{}{reqid, reqid_offlinectr_map[reqid], reqid_pctr_map[reqid], "不通过"}
      report_data = append(report_data, data)
      sample_fail_count += 1
    }
  }

  var reqid_slice []int
  for reqid, _ := range reqid_offlinectr_map {
    reqid_slice = append(reqid_slice, reqid)
  }
  sort.Ints(reqid_slice)
  for _, reqid := range reqid_slice {
    _, ok := has_resp_map[reqid]
    if !ok {
      logger.Infof("ResultsMap is empty for ReqID:%v", reqid)
      data := []interface{}{reqid, reqid_offlinectr_map[reqid], "Nil", "不通过"}
      report_data = append(report_data, data)
      sample_fail_count += 1
    }
  }
  if len(report_data) > MaxReportDataSize {
    report_map["Data"] = report_data[:MaxReportDataSize]
  } else {
    report_map["Data"] = report_data
  }
  if sample_fail_count == 0 && len(invalid_samples) == 0{
    err_update_database := service.UpdateDatabaseModelDesc(model_name, timestamp, "Validated", env)
    if err_update_database != nil {
      report_map["ValidateConclusion"] = "不通过，更新数据库失败"
    } else {
      report_map["ValidateConclusion"] = "通过"
    }
  } else if sample_fail_count != 0 {
    service.UpdateDatabaseModelDesc(model_name, timestamp, "Abandoned", env)
    report_map["ValidateConclusion"] = "不通过，样本线上预估CTR与离线CTR不一致"
  } else if len(invalid_samples) != 0 {
    service.UpdateDatabaseModelDesc(model_name, timestamp, "Abandoned", env)
    report_map["ValidateConclusion"] = "不通过，有无效样本"
  }
  report_map["SampleCount"] = sample_pass_count + sample_fail_count + int64(len(invalid_samples))
  report_map["SamplePassCount"] = sample_pass_count
  report_map["SampleFailCount"] = sample_fail_count
  report_map["SampleInvalidCount"] = len(invalid_samples)
  report_map["InvalidHeader"] = []string{"样本ID", "无效原因"}
  var invalid_data []([]interface{})
  var sample_id_slice []int
  for sample_id, _ := range invalid_samples {
    sample_id_slice = append(sample_id_slice, sample_id)
  }
  sort.Ints(sample_id_slice)
  for _, sample_id := range sample_id_slice {
    data := []interface{}{sample_id, invalid_samples[sample_id]}
    invalid_data = append(invalid_data, data)
  }
  if len(invalid_data) > MaxReportDataSize {
    report_map["InvalidData"] = invalid_data[:MaxReportDataSize]
  } else {
    report_map["InvalidData"] = invalid_data
  }
  service.SaveSummaryResult(fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", model_name, timestamp, report_map["ValidateType"], report_map["ValidateConclusion"], time.Now().Format("2006-01-02 15:04:05")), env)
  if report_map["ValidateConclusion"] == "通过" {
    logger.Infof("Model: %v-%v is validated successfully", model_name, timestamp)
  }
  // 如果是regression mode，屏蔽每个模型验证失败之后的邮件发送
  if env.IsRegressionMode() {
    return
  }
  report_recipients, claim_status := service.GetReportRecipientsByModel(model_name, env, report_map["ValidateConclusion"].(string))
  report_map["ClaimStatus"] = claim_status
  err_sendReport := service.SendHtmlReport(report_map, report_recipients, "一致性验证报告"+claim_status, path.Join(env.Conf.ValidateService.HtmlTemplateDir,"validate_report.tpl"), env)
  if err_sendReport != nil {
    logger.Errorf("%v", err_sendReport)
  }
  
}

func (service *ValidateService) SendHtmlReport(report_map map[string]interface{}, mailTo []string, subject string, html_template string, env *env.Env) error {
  logger.Infof("Send validate report mail to %v", mailTo)
  tpl := template.Must(template.ParseFiles(html_template))
  buf := new(bytes.Buffer)
  tpl.Execute(buf, report_map)
  html_str := buf.String()
  auth := common.NewLoginAuth("algo", "algo")
  msg_byte := []byte(html_str)
  err := common.SendMail("gmail.com:25", auth, "algo@gmail.com", mailTo, subject, msg_byte)
  if err != nil {
    //send alert to dingding
    err_str := fmt.Sprintf("[%v][Validate Service] Send validate report mail failed! err: %v", env.LocalIp, err)
    common.DingDingAlert(env.Conf.DingDingWebhookUrl, err_str)
    return errors.New(err_str)
  }
  logger.Infof("Send validate report mail successfully!")
  return nil
}

func (service *ValidateService) UpdateDatabaseModelDesc(model_name string, timestamp string, desc string, env *env.Env) error {
  db := env.Db
  sql_cmd := fmt.Sprintf("update model_histories set `desc`='%v' where model_name='%v' and timestamp='%v'", desc, model_name, timestamp)
  logger.Infof("SQL CMD: " + sql_cmd)
  res := db.Exec(sql_cmd)
  if res.Error == nil {
    logger.Infof("SQL result: %v", res)
  } else {
    //send alert to dingding
    err_str := fmt.Sprintf("[%v][Validate Service] UpdateDatabaseModelDesc failed for:%v%v%v", env.LocalIp, model_name, common.Delimiter, timestamp)
    common.DingDingAlert(env.Conf.DingDingWebhookUrl, err_str)
  }
  return res.Error
}

func (service *ValidateService) SaveSummaryResult(result string, env *env.Env) {
  date := fmt.Sprintf(time.Now().Format("2006-01-02"))
  file_name := path.Join(env.Conf.ValidateService.SummaryResultDir, "summary_result_" + date)
  if env.IsRegressionMode() {
    file_name += "_" + env.Conf.RegressionService.PacketName
  }
  _, err := service.WriteFile(file_name, []byte(result))
  if err != nil {
    logger.Errorf("WriteFile %v failed, err:", file_name, err)
  }
}

func (service *ValidateService) WriteFile(file_name string, data []byte) (int, error) {
  fl, err := os.OpenFile(file_name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
  if err != nil {
    return 0, err
  }
  defer fl.Close()
  n, err := fl.Write(data)
  if err == nil && n < len(data) {
    err = io.ErrShortWrite
  }
  return n, err
}

func (service *ValidateService) GenerateAndSendDailyReport(summary_file string, date string, env *env.Env) {
  report_map := make(map[string]interface{})
  report_map["Subject"] = "一致性验证报告--日报"
  if env.IsRegressionMode() {
    report_map["Subject"] = "回归测试报告_" + "DIFF_ID:" + env.Conf.RegressionService.PacketName
  }
  
  report_map["Date"] = date
  report_map["SummaryHeader"] = []string{"模型名称", "总验证次数", "通过次数", "未通过次数"}
  var summary_data []([]interface{})
  if env.IsRegressionMode() {
    summary_file += "_" + env.Conf.RegressionService.PacketName
  }
  logger.Infof("Open summary result file: " + summary_file)
  fp, err_file := os.Open(summary_file)
  if err_file != nil {
    logger.Infof("Open summary result file failed!")
  }
  defer fp.Close()
  r := bufio.NewReader(fp)
  var values []string
  model_total_map := make(map[string]int)
  model_pass_map := make(map[string]int)
  model_fail_map := make(map[string]int)
  model_results := make(map[string][]interface{})
  for {
    buf, err := r.ReadString('\n')
    if err == io.EOF || err != nil { //-1
        break
    }
    buf = strings.TrimSpace(buf)
    if(buf == "") {
      continue
    }
    values = strings.Split(buf, "\t")
    if len(values) != 5 {
      continue
    }
    logger.Infof("%v", values)
    model_total_map[values[0]] += 1
    if values[3] == "通过" {
      model_pass_map[values[0]] += 1
    }
    model_results[values[0]] = append(model_results[values[0]], values)
  }
  var models_slice []string
  for model, _ := range model_total_map {
    models_slice = append(models_slice, model)
  }
  sort.Strings(models_slice)
  var detail_results []interface{}
  for _, model := range models_slice {
    model_fail_map[model] = model_total_map[model]-model_pass_map[model]
    data := []interface{}{model, model_total_map[model], model_pass_map[model], model_fail_map[model]}
    summary_data = append(summary_data, data)
    detail_result := map[string]interface{} {"ModelName": model, "DetailHeader": []string{"模型名称", "模型版本", "验证类型", "验证结果", "验证完成时间"}, "DetailData": model_results[model]}
    detail_results = append(detail_results, detail_result)
  }
  report_map["SummaryData"] = summary_data
  report_map["DetailResults"] = detail_results
  report_recipients := service.GetAllReportRecipients(env)
  subject := "一致性验证报告"
  if env.IsRegressionMode() {
    subject = "回归测试报告"
  }
  err_sendReport := service.SendHtmlReport(report_map, report_recipients, subject, path.Join(env.Conf.ValidateService.HtmlTemplateDir,"validate_daily_report.tpl"), env)
  if err_sendReport != nil {
    logger.Errorf("%v", err_sendReport)
  }
}

func (service *ValidateService) GetLogInfo(model_name string, timestamp string, env *env.Env) (string, string, error) {
  dir_list, err_dir := ioutil.ReadDir(env.Conf.ValidateService.AlgoLogDir)
  if err_dir != nil {
      logger.Errorf("read dir error")
      return "", "", errors.New(fmt.Sprintf("Read dir err: %v", err_dir))
  }
  var file_slice []string
  for _, v := range dir_list {
      if strings.Contains(v.Name(), "INFO") {
          file_slice = append(file_slice, v.Name())
      }
  }
  if len(file_slice) > 0 {
    sort.Sort(sort.Reverse(sort.StringSlice(file_slice)))
    logger.Infof("Current algo service log: %v", path.Join(file_slice[0]))
    log_file_url := env.Conf.ValidateService.AlgoLogBaseUrl + file_slice[0];
    cmd_str := "grep -E " + fmt.Sprintf("'%v.*%v|%v.*%v' ", model_name, timestamp, timestamp, model_name) + path.Join(env.Conf.ValidateService.AlgoLogDir, file_slice[0])
    logger.Infof("cmd_str: %v", cmd_str)
    cmd := exec.Command("sh", "-c", cmd_str)
    out, err_sh := cmd.Output()
    if err_sh != nil {
        return "", "", errors.New(fmt.Sprintf("Execute grep command failed: %v", err_sh))
    } else {
      log_str := string(out)
      return log_file_url, log_str, nil
    }
  } else {
    return "", "", errors.New(fmt.Sprintf("algo service log not found"))
  }
}

func (service *ValidateService) GetReportRecipientsByModel(model_name string, env *env.Env, validate_result string) ([]string, string) {
  //if model record not exists will not trigger validate service, so must be a record here, just query
  var mail_recipients []string
  if env.IsRegressionMode() {
    default_recipients := env.Conf.ValidateService.ReportRecipients
    mail_recipients = append(mail_recipients, default_recipients...)
    return mail_recipients, ""
  }

  db := env.Db
  var model_recipients []string
  var model schema.Model
  if db.Where("name = ?", model_name).First(&model).RecordNotFound() {
    logger.Errorf(fmt.Sprintf("model_name=%v doesn't exist in table models", model_name))
  }
  if (model.Extension != "") {
    model_extension := new(api.ModelExtensionInfo)
    err_unmarshal := json.Unmarshal([]byte(model.Extension), model_extension)
    if err_unmarshal != nil {
      logger.Errorf(fmt.Sprintf("Unmarshal mysql extension json failed for model=%v, err:%v", model_name, err_unmarshal))
      //send alert to dingding
      err_str := fmt.Sprintf("[%v][Validate Service] Unmarshal mysql extension json failed for model=%v, err:%v", env.LocalIp, model_name, err_unmarshal)
      common.DingDingAlert(env.Conf.DingDingWebhookUrl, err_str)
    } else {
      model_recipients = append(model_recipients, model_extension.MailRecipients...)
    }
  }
  // when mail_recipients is not configured in models table then send mail to all
  if len(model_recipients) == 0 {
    return service.GetAllReportRecipients(env), "(待认领)"
  }

  if validate_result != "通过" {
    default_recipients := env.Conf.ValidateService.ReportRecipients
    mail_recipients = common.RemoveRepeatAndEmpty(append(model_recipients, default_recipients...))
  } else {
    mail_recipients = common.RemoveRepeatAndEmpty(model_recipients);
  }  

  return mail_recipients, ""
}

func (service *ValidateService) GetAllReportRecipients(env *env.Env) []string {
  var mail_recipients []string
  if env.IsRegressionMode() {
    default_recipients := env.Conf.ValidateService.ReportRecipients
    mail_recipients = append(mail_recipients, default_recipients...)
    return mail_recipients
  }
  
  db := env.Db
  var models_recipients []string
  var models []schema.Model
  db.Find(&models)
  for _, model := range models {
    if (model.Extension == "") {
      continue
    }
    model_extension := new(api.ModelExtensionInfo)
    err_unmarshal := json.Unmarshal([]byte(model.Extension), model_extension)
    if err_unmarshal != nil {
      logger.Errorf(fmt.Sprintf("Unmarshal mysql extension json failed for model=%v, err:%v", model.Name, err_unmarshal))
      //send alert to dingding
      err_str := fmt.Sprintf("[%v][Validate Service] Unmarshal mysql extension json failed for model=%v, err:%v", env.LocalIp, model.Name, err_unmarshal)
      common.DingDingAlert(env.Conf.DingDingWebhookUrl, err_str)
    } else {
      models_recipients = append(models_recipients, model_extension.MailRecipients...)
    }
  }
  default_recipients := env.Conf.ValidateService.ReportRecipients
  mail_recipients = common.RemoveRepeatAndEmpty(append(models_recipients, default_recipients...))
  return mail_recipients
}

func (service *ValidateService) GetCalcVectorRequestsFromSampleFile(requests *(predictor.CalculateVectorRequests), model_name string, timestamp string,
                                                                    sample_file string, env *env.Env, reqid_offline_output_map map[int]map[string][]float64,
                                                                    invalid_samples map[int]string, output_names *[]string, feature_map map[string]*api.ModelFeature) (*common.TagError) {
  logger.Infof("Open sample file: " + sample_file)
  fp, err_file := os.Open(sample_file)
  if err_file != nil {
    err_str := fmt.Sprintf("Open sample file failed! err: %v", err_file)
    return &common.TagError{err_str, "打开样本文件失败"}
  }
  defer fp.Close()
  r := bufio.NewReader(fp)
  var line_num int
  var keys []string
  // seperator of sample file
  var seperator string
  var features_name_list []string
  for {
    if line_num > env.Conf.ValidateService.MaxSampleCount {
      break
    }
    buf, err := r.ReadString('\n')
    if err == io.EOF || err != nil { //-1
        break
    }
    buf = strings.Replace(buf, "\n", "", -1)
    if(buf == "") {
      continue
    }
    if line_num == 0 {
      buf = strings.Replace(buf, " ", "", -1)
      if strings.Count(buf, ",") > 0 {
        seperator = ","
      } else if strings.Count(buf, ";") > 0 {
        seperator = ";"
      } else if strings.Count(buf, "\t") > 0 {
        seperator = "\t"
      } else {
        err_str := "no seperator found in the first line of sample file"
        return &common.TagError{err_str, "样本文件首行未发现分隔符"}
      }
      keys = strings.Split(buf, seperator)
      for i, key := range keys {
        if (i == 0) {
          *output_names = strings.Split(key, ":")
          logger.Infof("output_names=%v", *output_names)
          continue
        } else {
          features_name_list = append(features_name_list, key)
        }
      }
      logger.Infof("There are %v features provided: %v", len(features_name_list), features_name_list)
    } else { // 生成calculate vector request
      request := predictor.NewCalculateVectorRequest()
      request.ReqID = strconv.Itoa(line_num) + common.Delimiter + model_name + common.Delimiter + timestamp
      request.Channel = "validate_service"
      request.ModelName = model_name
      request.OutputNames = *output_names
      features := feature_master.NewFeatures()
      values := strings.Split(buf, seperator)
      if len(values) != len(keys) {
        invalid_samples[line_num] = "样本字段数与特征数不一致"
        line_num += 1
        continue
      }
      offline_outputs := make(map[string][]float64)
    J:
      for i, value := range values {
        value = strings.Trim(value, " ")
        if(i == 0) { // output
          outputs := strings.Split(value, ":") 
          for j, output := range outputs {
            vec_members := strings.Split(output, "\001")
            for _, member := range vec_members {
              member_float, err_parsefloat := strconv.ParseFloat(member, 64)
              if(err_parsefloat != nil) {
                err_str := fmt.Sprintf("Data parse err: %v", err_parsefloat)
                logger.Errorf("Invalid offline_output data! " + err_str)
                break J
              } else {
                offline_outputs[(*output_names)[j]] = append(offline_outputs[(*output_names)[j]], member_float)
              }
            }
          }
          reqid_offline_output_map[line_num] = offline_outputs
        } else {
          feature_name := keys[i]
          feature_attr, ok := feature_map[feature_name]
          if ok {
            feature, err := MakeFeature(feature_name, value, feature_attr)
            if err != nil {
              break J
            }
            features.Features = append(features.Features, feature)
          } else {
            feature := feature_master.NewFeature()
            feature.FeatureName = feature_name
            feature.StringValues = append(feature.StringValues, value)
            feature.FeatureType = feature_master.FeatureType_STRING_LIST
            features.Features = append(features.Features, feature)
          }
        }
      }
      if (len(features.Features) == len(features_name_list)) { // 特征数正确
        request.Features = features
        requests.Reqs = append(requests.Reqs, request)
      } else { // 特征数不正确，代码逻辑上由离线output数据类型不正确或字段数不正确造成
        logger.Errorf("The sample %v is invalid!", line_num)
        invalid_samples[line_num] = "样本数据不正确"
      }
    }
    line_num += 1
  }
  if line_num == 0 {
    err_str := "no data found in the sample file"
    return &common.TagError{err_str, "样本文件为空"}
  } else if line_num == 1 {
    err_str := "no sample found in the sample file"
    return &common.TagError{err_str, "样本文件未发现样本"}
  }
  logger.Infof("There are " + strconv.Itoa(len(requests.Reqs)) + " samples provided")
  return nil
}

func (service *ValidateService) GetComparisonResult(responses *(predictor.CalculateVectorResponses), reqid_offline_output_map map[int]map[string][]float64, output_names []string, fail_reasons_map map[int]string, online_offline_result_map map[int]map[string]string) {
  reqid_online_output_map := make(map[int]map[string][]float64)
  for _, resp := range responses.Resps {
    values := strings.Split(resp.ReqID, common.Delimiter);
    reqid, _ := strconv.Atoi(values[0])
    reqid_online_output_map[reqid] = resp.VectorMap
    online_offline_result_map[reqid] = make(map[string]string)
    online_offline_result_map[reqid]["online"] = PrintVectorMap(output_names, resp.VectorMap)
  }
  for id := range reqid_offline_output_map {
    if _, ok := online_offline_result_map[id]; !ok {
      online_offline_result_map[id] = make(map[string]string)
    }
    online_offline_result_map[id]["offline"] = PrintVectorMap(output_names, reqid_offline_output_map[id])
    if _, ok := reqid_online_output_map[id]; ok {
      for output := range reqid_offline_output_map[id] {
        if _, ok := reqid_online_output_map[id][output]; ok {
          is_similar,similarity := common.IsFloatVectorCosineSim(reqid_online_output_map[id][output], reqid_offline_output_map[id][output])
          if !is_similar {
            fail_reasons_map[id] = fmt.Sprintf("Predictor预测结果不一致，output_name:%v, similarity:%v", output, similarity)
          } else {
            logger.Infof("Predictor预测结果一致:", id)
          }
        } else {
          fail_reasons_map[id] = fmt.Sprintf("Predictor预测结果未发现:%v", output)
        }
      }
    } else {
      logger.Infof("response not found for reqid:", id)
      fail_reasons_map[id] = "无Predictor预测结果"
    }
  }
}

func (service *ValidateService) CalculateVector(requests *(predictor.CalculateVectorRequests), env *env.Env) (*(predictor.CalculateVectorResponses), error) {
  predictorClient, err_predictorClient := predictor_client.NewPredictorClient(env.Conf.ValidateService.ConsulAddress, env.Conf.ValidateService.ServiceName, env.LocalIp)
  if err_predictorClient != nil {
      logger.Errorf("Failed to init predictor client! err: %v", err_predictorClient)
      return nil, err_predictorClient
  }
  predictorClient.SetTimeout(time.Duration(env.Conf.ValidateService.PredictorTimeout) * time.Second)
  responses := predictor.NewCalculateVectorResponses()
  var err_calcVector error
  for i:=1; i<=env.Conf.ValidateService.RetryTimes; i++ {
    responses, err_calcVector = predictorClient.CalculateVector(requests)
    if err_calcVector == nil {
      logger.Infof("There are " + strconv.Itoa(len(responses.Resps)) + " results")
      break
    } else if i==env.Conf.ValidateService.RetryTimes {
      err_str := fmt.Sprintf("CalculateVector out of time, retry times:%d, retry interval:1s", env.Conf.ValidateService.RetryTimes)
      return nil, errors.New(err_str)
    } else {
      logger.Errorf("CalculateVector failed, will retry... err: %v", err_calcVector)
      time.Sleep(time.Second) //sleep for 1s
    }
  }
  return responses, nil
}

// 生成并发送向量计算模型的一致性验证报告
func (service *ValidateService) GenerateAndSendCalcVectorSampleValidateReport(online_offline_result_map map[int]map[string]string, fail_reasons_map map[int]string, invalid_samples map[int]string, output_names []string, model_name string, timestamp string, env *env.Env) {
  report_map := make(map[string]interface{})
  report_map["ModelName"] = model_name
  report_map["ModelVersion"] = timestamp
  report_map["ValidateType"] = "样本验证"
  report_map["Header"] = []string{"样本ID", "离线计算值", "线上计算值", "验证结果"}
  var report_data []([]interface{})
  var sample_pass_count, sample_fail_count int
  sample_fail_count = len(fail_reasons_map)
  sample_pass_count = len(online_offline_result_map) - len(fail_reasons_map)

  var reqid_slice []int
  for reqid, _ := range online_offline_result_map {
    reqid_slice = append(reqid_slice, reqid)
  }
  sort.Ints(reqid_slice)
  for _, reqid := range reqid_slice {
    if _, ok := fail_reasons_map[reqid]; ok {
      data := []interface{}{reqid, online_offline_result_map[reqid]["offline"], online_offline_result_map[reqid]["online"], fail_reasons_map[reqid]}
      report_data = append(report_data, data)
    }
  }
  if len(report_data) > MaxReportDataSize {
    report_map["Data"] = report_data[:MaxReportDataSize]
  } else {
    report_map["Data"] = report_data
  }
  if sample_fail_count == 0 && len(invalid_samples) == 0 {
    err_update_database := service.UpdateDatabaseModelDesc(model_name, timestamp, "Validated", env)
    if err_update_database != nil {
      report_map["ValidateConclusion"] = "不通过，更新数据库失败"
    } else {
      report_map["ValidateConclusion"] = "通过"
    }
  } else if sample_fail_count != 0 {
    service.UpdateDatabaseModelDesc(model_name, timestamp, "Abandoned", env)
    report_map["ValidateConclusion"] = "不通过，样本线上计算值与离线计算值不一致"
  } else if len(invalid_samples) != 0 {
    service.UpdateDatabaseModelDesc(model_name, timestamp, "Abandoned", env)
    report_map["ValidateConclusion"] = "不通过，有无效样本"
  }
  report_map["SampleCount"] = sample_pass_count + sample_fail_count + len(invalid_samples)
  report_map["SamplePassCount"] = sample_pass_count
  report_map["SampleFailCount"] = sample_fail_count
  report_map["SampleInvalidCount"] = len(invalid_samples)
  report_map["InvalidHeader"] = []string{"样本ID", "无效原因"}
  var invalid_data []([]interface{})
  var sample_id_slice []int
  for sample_id, _ := range invalid_samples {
    sample_id_slice = append(sample_id_slice, sample_id)
  }
  sort.Ints(sample_id_slice)
  for _, sample_id := range sample_id_slice {
    data := []interface{}{sample_id, invalid_samples[sample_id]}
    invalid_data = append(invalid_data, data)
  }
  if len(invalid_data) > MaxReportDataSize {
    report_map["InvalidData"] = invalid_data[:MaxReportDataSize]
  } else {
    report_map["InvalidData"] = invalid_data
  }
  service.SaveSummaryResult(fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", model_name, timestamp, report_map["ValidateType"], report_map["ValidateConclusion"], time.Now().Format("2006-01-02 15:04:05")), env)
  if report_map["ValidateConclusion"] == "通过" {
    logger.Infof("Model: %v-%v is validated successfully", model_name, timestamp)
  }
  // 如果是regression mode，屏蔽每个模型验证完成之后的邮件发送
  if env.IsRegressionMode() {
    return
  }
  report_recipients, claim_status := service.GetReportRecipientsByModel(model_name, env, report_map["ValidateConclusion"].(string))
  report_map["ClaimStatus"] = claim_status
  err_sendReport := service.SendHtmlReport(report_map, report_recipients, "一致性验证报告"+claim_status, path.Join(env.Conf.ValidateService.HtmlTemplateDir,"validate_report.tpl"), env)
  if err_sendReport != nil {
    logger.Errorf("%v", err_sendReport)
  }
}

func PrintVectorMap(output_names []string, output_vector map[string][]float64) string {
  var vector_str string
  for _, output := range output_names {
    vector_str += output + ":"
    for _, i := range output_vector[output] {
      vector_str += fmt.Sprintf("%.6f", i) + ","
    }
    vector_str = strings.Trim(vector_str, ",")
    vector_str += "|"
  }
  vector_str = strings.Trim(vector_str, "|")
  return vector_str
}

func GetFeatureInfoFromMetaInfoFile(model_feature_map map[string]*api.ModelFeature, meta_info_file_path string) (*common.TagError) {
  logger.Infof("Open meta_info file:%v", meta_info_file_path)
  fp, err_file := os.Open(meta_info_file_path)
  if err_file != nil {
    err_str := fmt.Sprintf("Open meta info file failed! err: %v", err_file)
    logger.Errorf(err_str)
    return &common.TagError{err_str, "打开Meta信息文件失败"}
  }
  defer fp.Close()
  content, _ := ioutil.ReadAll(fp)
  var dat map[string]interface{}
  if err_unmarshal := json.Unmarshal(content, &dat); err_unmarshal != nil {
    err_str := fmt.Sprintf("Unmarshal meta info file failed! err: %v", err_unmarshal)
    return &common.TagError{err_str, "解析Meta信息文件失败"}
  }

  if v, ok := dat["FeatureType"]; ok {
    feature_infos := v.([]interface{})
    for _, info := range feature_infos {
      info_map := info.(map[string]interface{})
      feature_name, _ := info_map["Name"].(string)
      feature_type, _ := info_map["Type"].(string)
      feature_dim, _ := info_map["Dim"].(string)
      model_feature_map[feature_name] = new(api.ModelFeature)
      model_feature_map[feature_name].Dim = feature_dim
      if feature_type == "string" {
        model_feature_map[feature_name].Type = 1  // string
      } else if feature_type == "float" {
        model_feature_map[feature_name].Type = 2  // float
      } else if feature_type == "int" {
        model_feature_map[feature_name].Type = 3  // int
      }
    }
  }
  return nil
}

func MakeFeature(feature_name string, value string, feature_attr *api.ModelFeature) (*feature_master.Feature, error) {
  feature := feature_master.NewFeature()
  feature.FeatureName = feature_name
  switch feature_attr.Type {
    case String:
      mems := strings.Split(value, common.CTRL_A)
      for _, mem := range mems {
        feature.StringValues = append(feature.StringValues, mem)
      }
      feature.FeatureType = feature_master.FeatureType_STRING_LIST
    case Float:
      mems := strings.Split(value, common.CTRL_A)
      for _, mem := range mems {
        format_value, err_parseFloat := strconv.ParseFloat(mem, 64)
        if err_parseFloat != nil {
          err_str := fmt.Sprintf("Data parse err: %v", err_parseFloat)
          logger.Errorf("The type of feature: " + feature_name + " is different from the config in feature service. " + err_str)
          return feature, fmt.Errorf(err_str);
        }
        feature.FloatValues = append(feature.FloatValues, format_value)
      }
      feature.FeatureType = feature_master.FeatureType_FLOAT_LIST
    case Int64:
      mems := strings.Split(value, common.CTRL_A)
      for _, mem := range mems {
        format_value, err_parseInt := strconv.ParseInt(mem, 10, 64)
        if err_parseInt != nil {
          err_str := fmt.Sprintf("Data parse err: %v", err_parseInt)
          logger.Errorf("The type of feature: " + feature_name + " is different from the config in feature service." + err_str)
          return feature, fmt.Errorf(err_str);
        }
        feature.Int64Values = append(feature.Int64Values, format_value)
      }
      feature.FeatureType = feature_master.FeatureType_INT64_LIST
  }
  return feature, nil;
}