package server

import (
	"content_service/common"
	"content_service/env"
	"content_service/libs/logger"
	"content_service/logics"
	"content_service/metrics"
	"content_service/schema"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
	"regexp"
	"reflect"
	"encoding/json"
	"net/http"
	"bytes"
)

type P2PModelService struct {
	ServiceCheckTimesMap   sync.Map
	ModelCheckTimesMap   sync.Map
	ServiceFailedModelsMap sync.Map
	ServiceModelsMap       sync.Map
}

func NewP2PModelService() *P2PModelService {
	return &P2PModelService{}
}

func (service *P2PModelService) Run(env *env.Env) {
	logger.Infof("ip=%v: Model Service running...", env.LocalIp)
	service.routineCheck(env)
}

// run check() periodically
func (service *P2PModelService) routineCheck(env *env.Env) {
	for t := range time.Tick(env.Conf.P2PModelService.RunInterval * time.Second) {
		ts := metrics.GetTimers()[metrics.TIMER_P2P_MODEL_SERVICE_CHECK_TIMER].Start()
		if err := service.check(env); err != nil {
			msg := fmt.Sprintf("P2P Model Service (%s) error: %s\n", env.LocalIp, err.Error())
			logger.Errorf("%v %s", t, msg)
			if tagErr, ok := err.(*common.TagError); ok {
				metrics.GetErrorMeter(metrics.TAG_P2P_MODEL_SERVICE, tagErr.ErrTag).Mark(1)
			}
			metrics.GetErrorMeter(metrics.TAG_P2P_MODEL_SERVICE, metrics.TAG_CHECK_ERROR).Mark(1)
			common.Alert(
				env.Conf.Alert.Recipients,
				fmt.Sprintf("P2P Model Service (%s) error!", env.LocalIp),
				msg,
				env.Conf.Alert.Rate)
		}
		ts.Stop()
	}
}

// single round of check()
func (service *P2PModelService) check(env *env.Env) error {
	logger.Debugf("\n-------------------------------------\n>> P2P Pull model started at %v\n\n",
		time.Now().Format("20060102_150405"))
	// 初始化ServiceModelsMap, 防止service变动留下垃圾数据
	service.ServiceModelsMap = sync.Map{}

	// 1.从数据库获取当前机器所有service
	dbServices, err := logics.FetchDbServices(env)
	if err != nil {
		return &common.TagError{
			fmt.Sprintf("fetchDbServices() err: %v, DB Data: %v", err, common.Pretty(dbServices)),
			metrics.TAG_FETCH_DB_SERVICES_ERROR,
		}
	}
	// 清除service变动留下的垃圾数据
	service.cleanServiceCheckTimesMap(dbServices)

	if (env.Conf.EnalbleRouterService == true) {
		for _, dbService := range dbServices {
			if dbService.Name == common.PredictorRouterServiceName {
				return service.DoRouterModeJob(env, dbService)
			}
		}
		
		if (len(dbServices) != 0) {
			err = logics.SetPredictorWorkMode(common.PredictorServerMode, env, env.LocalIp, env.Conf.P2PModelService.TargetService.HttpPort)
			if err != nil {
				return &common.TagError{
					fmt.Sprintf("SetPredictorWorkMode %v err: %v", common.PredictorServerMode, err),
					metrics.TAG_SET_PREDICTOR_WORK_MODE_ERROR,
				}
			}
		}
  }

	// 2.拉取每个service下的模型
	var wg sync.WaitGroup
	errChan := make(chan error)
	for _, dbService := range dbServices {
		wg.Add(1)
		// 获取并拉取当前service下的模型
		go service.fetchAndPullServiceModels(env, &wg, dbService, errChan)
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return &common.TagError{
			fmt.Sprintf("fetchAndPullServiceModels() errs: %v", errs),
			metrics.TAG_FETCH_AND_PULL_SERVICE_MODELS_ERROR,
		}
	}

	// 3.汇总拉取错误次数, 获取拉取失败模型
	var currentPullErrTimes int32
	var currentFailedServiceName string
	service.ServiceCheckTimesMap.Range(func(k, v interface{}) bool {
		service_name, _ := k.(string)
		num, _ := v.(int32)
		if num > currentPullErrTimes {
			currentPullErrTimes = num
			currentFailedServiceName = service_name
		}
		return true
	})

	curServiceModelsMap := make(map[string]map[string]schema.ModelHistory)
	service.ServiceModelsMap.Range(func(k, v interface{}) bool {
		service_name, _ := k.(string)
		modelsMap, _ := v.(map[string]schema.ModelHistory)
		curServiceModelsMap[service_name] = modelsMap
		return true
	})
	logger.Debugf("\n>> DB Data serviceModelsMap: %v", common.Pretty(curServiceModelsMap))
	// 4.成功, 通知predictor, 否则判断次数报警
	if currentPullErrTimes == 0 {
		// get service load weight from db
		db_service_weight, err_weight := logics.GetServiceWeight(env, env.LocalIp)
		if err_weight != nil {
			return &common.TagError{
				fmt.Sprintf("getServiceWeight() err: %v, DB Data: %v", err_weight.Error(), common.Pretty(db_service_weight)),
				metrics.TAG_GETWEIGHT_ERROR,
			}
		}
		logger.Debugf("\n>> Service Weight: %v\n\n", common.Pretty(db_service_weight))

		// after pulling over model files, notify predictor
		// note here we post the full collection of what we found in db
		if (env.LocalIp != env.Conf.ValidateService.Host && !env.IsRegressionMode()) {
			var service_names []string
			for k, _ := range db_service_weight {
				service_names = append(service_names, k)
			}

			// get service config parameters from db
			service_configs, err_config := logics.GetServiceConfig(env, service_names)
			if err_config != nil {
				return &common.TagError{
					fmt.Sprintf("getServiceConfig() err: %v, DB Data: %v", err_config.Error(), common.Pretty(service_configs)),
					metrics.TAG_GETCONFIG_ERROR,
				}
			}
			logger.Debugf("\n>> Service Config: %v\n\n", common.Pretty(service_configs))

			if err := logics.NotifyPredictor(curServiceModelsMap, db_service_weight, service_configs, env, env.LocalIp, env.Conf.P2PModelService.TargetService.HttpPort); err != nil {
				return &common.TagError{
					fmt.Sprintf("notifyPredictor() err: %v, DB Data: %v, Service Weight: %v",
						err.Error(), common.Pretty(curServiceModelsMap), common.Pretty(db_service_weight)),
					metrics.TAG_NOTIFY_ERROR,
				}
			}
		} else {
			service_model_map_slice := logics.SplitServiceModelsMap(curServiceModelsMap)
			empty_service_config := make(map[string]string)
			for _, db_service_model := range service_model_map_slice {
				if err := logics.NotifyPredictor(db_service_model, db_service_weight, empty_service_config, env, env.LocalIp, env.Conf.P2PModelService.TargetService.HttpPort); err != nil {
					return &common.TagError{
						fmt.Sprintf("notifyPredictor() err: %v, DB Data: %v, Service Weight: %v",
							err.Error(), common.Pretty(db_service_model), common.Pretty(db_service_weight)),
						metrics.TAG_NOTIFY_ERROR,
					}
				}
				// validate
				GetValidateInstance().RunValidateService(db_service_model, env)
			}
		}
		logger.Debugf("\n>> P2P Pull model finished.\n-------------------------------------\n")
	} else if int(currentPullErrTimes) >= env.Conf.P2PModelService.ServicePullAlertLimit {
		// 判断是否达到限定次数，报警
		currentFailedModels := []string{}
		if failedModelsI, exists := service.ServiceFailedModelsMap.Load(currentFailedServiceName); exists {
			currentFailedModels = failedModelsI.([]string)
		}
		err := fmt.Errorf("Pull model over limit times then alert, currentPullErrTimes: %d, servicePullAlertLimit: %d, currentFailedModels: %v, currentFailedServiceName: %s",
			currentPullErrTimes, env.Conf.P2PModelService.ServicePullAlertLimit, currentFailedModels, currentFailedServiceName)
		logger.Errorf("check() err: %v", err)

		// quit validation if reach pull max limit for validate host
		if env.LocalIp == env.Conf.ValidateService.Host {
			service.QuitValidateForPullFailure(currentFailedModels, env)
	  }
	} else {
		logger.Debugf("Pull models from parent host not success, currentPullErrTimes: %d", currentPullErrTimes)
	}

	// 判断当前机器是否是压测Service，如果是的话，需要额外post请求，通知algo service压测相关的信息
	if service.IsStressTestService(env,dbServices) {
		if err := service.CheckAndPostStressInfo(env); err != nil {
			msg := fmt.Sprintf("P2P Model Service (%s) error: %s\n", env.LocalIp, err.Error())
			logger.Errorf("%s", msg)
			return &common.TagError{
				fmt.Sprintf("CheckAndPostStressInfo() err: %v", err.Error()),
				metrics.TAG_STRESS_SERVICE,
			}
		}
	}
	return nil
}

// 获取并拉取每个service下的模型列表
func (service *P2PModelService) fetchAndPullServiceModels(env *env.Env, wg *sync.WaitGroup, dbService *schema.Service, errChan chan<- error) {
	defer wg.Done()
	service_name := dbService.Name

	// 1.根据每个service 获取要拉取的模型列表
	modelsByService, err := service.fetchDbModelsByService(env, dbService.ID)
	if err != nil {
		errChan <- fmt.Errorf("fetchDbModelsByService err: %v, service.ID: %d", err, dbService.ID)
		return
	}
	service.ServiceModelsMap.Store(service_name, modelsByService)
	if len(modelsByService) == 0 {
		service.ServiceCheckTimesMap.Store(service_name, int32(0))
		service.ClearModelCheckTimesMapAndMetrics()
		return
	}
	logger.Debugf("\n>> DB Models By Service: %v, service_name: %s\n\n", common.Pretty(modelsByService), service_name)

	// 2.获取本地已有模型列表
	disk_models, err := logics.FetchDiskData(env, env.Conf.P2PModelService.DestPath)
	if err != nil {
		logger.Errorf("fetchDiskData err: %v, disk_models: %v", err, disk_models)
		errChan <- fmt.Errorf("fetchDiskData err: %v, disk_models: %v", err, disk_models)
		return
	}
	logger.Debugf("\n>> Disk Data: %v, service_name: %s\n\n", common.Pretty(disk_models), service_name)

	// 3.diff数据库和本地模型，找到没拉取的列表
	models_to_pull := service.findModelsByServiceToPull(modelsByService, disk_models)
	if len(models_to_pull) == 0 {
		service.ServiceCheckTimesMap.Store(service_name, int32(0))
		service.ClearModelCheckTimesMapAndMetrics()
		return
	}
	logger.Debugf("\n>> Models to pull: %v, service_name: %s\n\n", common.Pretty(models_to_pull), service_name)

	// 4.获取每个service 的parent节点
	parentIP, peerNum, err := logics.GetParentIP(env, dbService, service.ServiceCheckTimesMap, env.Conf.P2PModelService.ServicePullAlertLimit, env.Conf.P2PModelService.SrcHost)
	if err != nil {
		logger.Errorf("GetParentIP err: %v, parentIP: %s", err, parentIP)
		errChan <- fmt.Errorf("GetParentIP err: %v, parentIP: %s", err, parentIP)
		return
	}
	logger.Infof("GetParentIP success, service_name: %s, parentIP: %s, peerNum: %d", service_name, parentIP, peerNum)

	// 5.开始拉取
	failedModels := service.pullModelFiles(env, models_to_pull, parentIP, peerNum)

	// 6.未成功，记录次数，成功，清零
	if len(failedModels) != 0 {
		for _, model_name := range failedModels {
			if currentTimesI, exists := service.ModelCheckTimesMap.Load(model_name); !exists {
				service.ModelCheckTimesMap.Store(model_name, int32(1))
			} else {
				currentTimes, _ := currentTimesI.(int32)
				service.ModelCheckTimesMap.Store(model_name, currentTimes+int32(1))
			}
		}
		if currentTimesI, exists := service.ServiceCheckTimesMap.Load(service_name); !exists {
			service.ServiceCheckTimesMap.Store(service_name, int32(1))
		} else {
			currentTimes, _ := currentTimesI.(int32)
			service.ServiceCheckTimesMap.Store(service_name, currentTimes+int32(1))
		}
		service.ServiceFailedModelsMap.Store(service_name, failedModels)
		service.UpdateModelCheckTimesMapAndMetrics(failedModels)
	} else {
		service.ServiceCheckTimesMap.Store(service_name, int32(0))
		service.ServiceFailedModelsMap.Store(service_name, []string{})
		service.ClearModelCheckTimesMapAndMetrics()
	}
}

func (service *P2PModelService) fetchDbGlobalModelServiceMap(env *env.Env) (map[string]string, error) {
	db := env.Db
	globalModelServiceMap := make(map[string]string)
	type ModelService struct {
    ModelName  string
    ServiceName string
  }
	var model_service_vec []ModelService
	sql := `SELECT m.name as model_name ,s.name as service_name FROM service_models sm
		LEFT JOIN services s ON s.id = sm.sid
		LEFT JOIN models m ON m.id = sm.mid
		WHERE s.name not in (?, ?)`
	dbPtr := db.Raw(sql, env.Conf.ValidateService.ServiceName, env.Conf.StressTestService.ServiceName).Scan(&model_service_vec)
	logger.Debugf(">> sql=\"%s, ? = %s\"", sql, env.Conf.ValidateService.ServiceName)
	if errs := dbPtr.GetErrors(); len(errs) > 0 {
		return globalModelServiceMap, fmt.Errorf("gorm db err: sql=%v err=%v", sql, errs)
	}
	for _, model_service := range model_service_vec {
		globalModelServiceMap[model_service.ModelName] = model_service.ServiceName
	}
	return globalModelServiceMap, nil
}

func (service *P2PModelService) updateGlobalModelServiceMap(env *env.Env, globalModelServiceMap map[string]string) error {
	var reqUrl = fmt.Sprintf("http://%s:%s/update_global_model_service_map", env.LocalIp, env.Conf.P2PModelService.TargetService.HttpPort)

	requestBody, err := json.Marshal(globalModelServiceMap)
	resp, err := http.Post(reqUrl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("HTTPPost err: %+v, reqUrl: %s", err, reqUrl)
	}
	// parse http response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("parse http response err: %v, reqUrl: %v", err, reqUrl)
	}
	var respStruct struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	err = json.Unmarshal(body, &respStruct)
	if err != nil {
		return fmt.Errorf("Unmarshal http response err: %v, reqUrl: %v ,body : %s", err, reqUrl, string(body))
	}
	if respStruct.Code != 0 {
		return fmt.Errorf("updateGlobalModelServiceMap response fail , reqUrl: %v ,respStruct : %+v", reqUrl, respStruct)
	}
	return nil
}

// 获取单个service下要拉取的模型列表
// returns a map of: model_name -> model histories
func (service *P2PModelService) fetchDbModelsByService(env *env.Env, sid uint) (map[string]schema.ModelHistory, error) {
	db := env.Db
	// return value
	modelsByService := make(map[string]schema.ModelHistory)
	var model_names []string
	// search for responsible service names and model names
	// 针对线下一致性验证，选取所有在model中的模型作为还在线上使用的模型。
	var sql_query string
	if env.IsRegressionMode() {
		sql_query = `SELECT m.name FROM models m where m.id in (select mid from service_models where sid != ?)`		
	}	else {
		sql_query = `SELECT m.name FROM service_models sm 
		INNER JOIN models m ON sm.mid = m.id
		WHERE sm.sid = ?`
	}
	dbPtr := db.Raw(sql_query, sid).Pluck("m.name", &model_names)
	logger.Debugf(">> sql_query=\"%s, ? = %d\"", sql_query, sid)
	if errs := dbPtr.GetErrors(); len(errs) > 0 {
		logger.Errorf("fetchDbModelsByService err: %v, sid: %d", errs, sid)
		return modelsByService, fmt.Errorf("gorm db err: sql=%v err=%v", sql_query, errs)
	}

	// get locked or lastest model version from model_history
	for _, model_name := range model_names {
		if len(model_name) == 0 {
			continue
		}
		if _, exists := modelsByService[model_name]; !exists {
			// find model history by model_name
			var model_history schema.ModelHistory
			sql_condition := "AND `desc` = 'Validated'"
			if env.LocalIp == env.Conf.ValidateService.Host {
				sql_condition = ""
			}
			// we should not expect more than 1 locked model history per model name, so just try fetch the first one
			if db.Where(fmt.Sprintf("is_locked = ? AND model_name = ? %v", sql_condition), "1", model_name).First(&model_history).RecordNotFound() {
				// did not found locked model, search for newest timestamp
				if db.Where(fmt.Sprintf("model_name = ? %v", sql_condition), model_name).Order("timestamp desc").First(&model_history).RecordNotFound() {
					continue
				}
			}
			var empty_model_history schema.ModelHistory
			if reflect.DeepEqual(model_history, empty_model_history) {
				continue
			}
			// previous sql get newest model for validate service but need to check if the model should to be validated
			if env.LocalIp == env.Conf.ValidateService.Host && model_history.Desc != "" {
				continue
			}
			if service.IsLegalModelHistoryRecord(model_history, env) {
			  modelsByService[model_name] = model_history
			}
		}
	}

	return modelsByService, nil
}

// 过滤本地存在的模型并去重
func (service *P2PModelService) findModelsByServiceToPull(
	modelsByService map[string]schema.ModelHistory,
	disk_models map[string]bool) []string {
	models_to_pull := []string{}
	db_models_dedup := make(map[string]bool)

	for _, db_model := range modelsByService {
		db_model_full_name := db_model.FullName()
		if db_model_full_name == "" {
			continue
		}
		_, exists_in_results := db_models_dedup[db_model_full_name]
		_, exists_on_disk := disk_models[db_model_full_name]
		if !exists_in_results && !exists_on_disk {
			db_models_dedup[db_model_full_name] = true
			models_to_pull = append(models_to_pull, db_model_full_name)
		}
	}

	return models_to_pull
}

// fetch back a set of models we already have from local disk
func (service *P2PModelService) fetchDiskData(env *env.Env) (map[string]bool, error) {
	var disk_models = map[string]bool{}

	var dest_path = env.Conf.P2PModelService.DestPath
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

// 注意：这个model_names 为要拉取的模型目录的列表
// model_name是带版本号的名称，eg：tf_xfea_estimator_v0_fans_economy_nonfans-20200311_005210
func (service *P2PModelService) pullModelFiles(env *env.Env, model_names []string, parentIP string, peerNum int) []string {
	var failed_models = []string{}
	num := len(model_names) // num of models to pull
	if num == 0 {
		return failed_models
	}

	// pull model files concurrently
	ch := make(chan string) // a channel to record failed models
	var wg sync.WaitGroup
	wg.Add(num) // spawn "num" goroutines
	for i := 0; i < num; i++ {
		go service.pullModelFile(env, model_names[i], parentIP, peerNum, &wg, ch)
	}

	go func() {
		t1 := time.Now()
		ts := metrics.GetTimers()[metrics.TIMER_P2P_MODEL_SERVICE_PULLMODEL_TIMER].Start()
		wg.Wait() // wait for "num" goroutines to finish
		ts.Stop()
		t2 := time.Now()
		diff := t2.Sub(t1)
		logger.Infof("pullModelFiles() cost: %v\n", diff)
		close(ch)
	}()

	for name := range ch {
		failed_models = append(failed_models, name)
	}

	return failed_models
}

// pull one model file from transfer host, with retries
func (service *P2PModelService) pullModelFile(env *env.Env, model_name string, parentIP string, peerNum int, wg *sync.WaitGroup, ch chan<- string) {
	defer wg.Done()
	model_name_without_version := common.TrimModelVersion(model_name)
	ts := metrics.GetPullSingleModelTimer(model_name_without_version).Start()
	defer ts.Stop()
	i := 0
	for ; i <= env.Conf.P2PModelService.Retry; i++ {
		model_path := path.Join(env.Conf.P2PModelService.SrcPath, model_name)
		destPath := path.Join(env.Conf.P2PModelService.DestPath, model_name)
		tmpDestPath := path.Join(env.Conf.P2PModelService.DestPath, model_name+".tmp")
		peerNum = common.MaxInt(1, peerNum)
		bwLimit := common.MinInt(int(env.Conf.P2PModelService.SrcRsyncBWLimit/peerNum), env.Conf.P2PModelService.RsyncBWLimit)
		cmd := exec.Command("/bin/rsync", "-rp", "--bwlimit="+strconv.Itoa(bwLimit), parentIP+"::"+model_path+"/", tmpDestPath+"/")
		logger.Debugf("cmd: %s", "/bin/rsync -rp --bwlimit="+strconv.Itoa(bwLimit)+" "+parentIP+"::"+model_path+"/ "+tmpDestPath+"/")
		if stdout, err := cmd.CombinedOutput(); err != nil {
			if strings.Contains(string(stdout), "No such file or directory") {
				logger.Debugf("rsync model not exists, model_path=%s, err=%v, output=%s", model_path, err, stdout)
				break
			}
			logger.Errorf("rsync cmd failed: err=%v, output=%s", err, stdout)
			continue
		} else {
			// succeeded mv tmp to real destpath
			cmd := exec.Command("mv", tmpDestPath, destPath)
			logger.Debugf("cmd: mv %s %s", tmpDestPath, destPath)
			if stdout, err := cmd.CombinedOutput(); err != nil {
				logger.Errorf("mv cmd failed: err=%v, output=%s", err, stdout)
				continue
			}
			return
		}
	}

	// failed after all retries, add failed model name to channel
	ch <- model_name
	logger.Errorf("pull model=%v failed after retried %v times", model_name, env.Conf.P2PModelService.Retry)
}

// 去除service变动留下的垃圾计数
func (service *P2PModelService) cleanServiceCheckTimesMap(dbServices []*schema.Service) {
	if len(dbServices) == 0 {
		service.ServiceCheckTimesMap = sync.Map{}
		service.ServiceFailedModelsMap = sync.Map{}
	}
	var serviceMap = make(map[string]bool)
	for _, dbService := range dbServices {
		serviceMap[dbService.Name] = true
	}
	service.ServiceCheckTimesMap.Range(func(k, v interface{}) bool {
		service_name, _ := k.(string)
		_, exists := serviceMap[service_name]
		if !exists {
			service.ServiceCheckTimesMap.Delete(service_name)
		}
		return true
	})
	service.ServiceFailedModelsMap.Range(func(k, v interface{}) bool {
		service_name, _ := k.(string)
		_, exists := serviceMap[service_name]
		if !exists {
			service.ServiceFailedModelsMap.Delete(service_name)
		}
		return true
	})
}

func (service *P2PModelService) QuitValidateForPullFailure(currentFailedModels []string, env *env.Env) {
  for _, failed_model := range currentFailedModels {
		currentTimesI, exists := service.ModelCheckTimesMap.Load(failed_model)
		if !exists {
			continue
		}
		if currentTimesI.(int32) >= int32(env.Conf.P2PModelService.ModelPullMaxLimit) {
			// send validation mail and DingDing alert
			model_version := strings.Split(failed_model, common.Delimiter)
			var model_name, timestamp string
			if len(model_version) != 2 {
				model_name = failed_model
			} else {
				model_name = model_version[0]
				timestamp = model_version[1]
			}
			err_str := fmt.Sprintf("[%v][P2P Model Service] pull model failed for: %v", env.LocalIp, failed_model)
			logger.Errorf(err_str)
			err_pull := &common.TagError{fmt.Sprintf("%v", err_str), "拉取模型失败"}
			GetValidateInstance().ConcludeAndSendBaseValidateReport(model_name, timestamp, env, "基础验证", err_pull)
			common.DingDingAlert(env.Conf.DingDingWebhookUrl, err_str)
		}
	}
}

// 判定模型名和时间戳是否为合法，不合法则直接验证不通过
func (service *P2PModelService) IsLegalModelHistoryRecord(model_history schema.ModelHistory, env *env.Env) bool {
	name_match, _ := regexp.MatchString("^[a-z]{1,}([a-z0-9_]*)[a-z0-9]$", model_history.ModelName)
	version_match, _ := regexp.MatchString("^[0-9]{8}_[0-9]{6}$", model_history.Timestamp)
	if name_match && version_match {
		return true
	}
	err_str := fmt.Sprintf("[%v][P2P Model Service] illegal model history: %v", env.LocalIp, model_history)
	logger.Errorf(err_str)
	err_checkhistory := &common.TagError{fmt.Sprintf("%v", err_str), "非法模型名或时间戳"}
	if env.LocalIp == env.Conf.ValidateService.Host && model_history.Desc == "" {
		GetValidateInstance().ConcludeAndSendBaseValidateReport(model_history.ModelName, model_history.Timestamp, env, "基础验证", err_checkhistory)
	}
	common.DingDingAlert(env.Conf.DingDingWebhookUrl, err_str)
	return false
}

func (service *P2PModelService) ClearModelCheckTimesMapAndMetrics() {
	service.ModelCheckTimesMap.Range(func(k, v interface{}) bool {
		service.ModelCheckTimesMap.Store(k, int32(0))
		tag := map[string]string{"model_name": k.(string)}
		metrics.GetMetrics().Tagged(tag).Gauge(metrics.GAUGE_MODEL_PULL_ERR_TIMES).Update(float64(0))
		return true
	})
}

func (service *P2PModelService) UpdateModelCheckTimesMapAndMetrics(currentFailedModels []string) {
	for _, failed_model := range currentFailedModels {
		currentTimesI, exists := service.ModelCheckTimesMap.Load(failed_model)
		if !exists {
			continue
		}
	  tag := map[string]string{"model_name": failed_model}
		metrics.GetMetrics().Tagged(tag).Gauge(metrics.GAUGE_MODEL_PULL_ERR_TIMES).Update(float64(currentTimesI.(int32)))
	}
	service.ModelCheckTimesMap.Range(func(k, v interface{}) bool {
		model_name := k.(string)
		if !common.Contains(currentFailedModels, model_name) {
			service.ModelCheckTimesMap.Store(model_name, int32(0))
			tag := map[string]string{"model_name": model_name}
			metrics.GetMetrics().Tagged(tag).Gauge(metrics.GAUGE_MODEL_PULL_ERR_TIMES).Update(float64(0))
		}
		return true
	})
}

func (service *P2PModelService) IsStressTestService(env *env.Env, cur_services []*schema.Service) bool {
	if len(cur_services) != 1 {
		logger.Debugf("current ip=%v has more service,error", env.LocalIp)
		return false
	}
	cur_service := cur_services[0]
	if cur_service.Name != env.Conf.StressTestService.ServiceName {
		return false
	}
	return true
}


func (service *P2PModelService) CheckAndPostStressInfo(env *env.Env) error {
	model_names_list, qps, err := service.FetchStressInfo(env)
	if err != nil {
		return err
	}
	
	if err := logics.NotifyPredictorStressInfo(model_names_list, qps, env, env.LocalIp, env.Conf.P2PModelService.TargetService.HttpPort); err != nil {
		return &common.TagError{
			fmt.Sprintf("NotifyPredictorStressInfo() err: %v",
				err.Error()),
			metrics.TAG_NOTIFY_STRESS_ERROR,
		}
	}
	return nil
}

func (service *P2PModelService) FetchStressInfo(env *env.Env) (string, string, error) {
	db := env.Db
	sql_query := `select si.mids,si.qps FROM stress_infos si
	LEFT JOIN hosts h ON h.id = si.hid 
	WHERE si.is_enable = ? and h.ip = ?`
	rows, err := db.Raw(sql_query, "1", env.LocalIp).Rows()	
	logger.Debugf("\n>> sql_query=\"%v, ? = %v\"\n", sql_query, env.LocalIp)
	if err != nil {
		return "", "", fmt.Errorf("gorm db err: sql=%v err=%v", sql_query, err.Error())
	}

	var mids string
	var qps string
	for rows.Next() {
		if err := rows.Scan(&mids, &qps); err != nil {
			logger.Errorf("rows.Scan err: %v", err)
			continue
		}
	}

	if len(mids) == 0 || len(qps) == 0 {
		return "", "", fmt.Errorf("mids or qps is nil")
	}

	model_id_list := strings.Split(mids, ",")
	var model_names []string
	for _, model_id := range model_id_list {
		i_type, type_err := strconv.Atoi(model_id)
		if type_err != nil {
			 return "", "", type_err
		}
		models := &schema.Model{}
		db.Where(schema.Model{ID: uint(i_type)}).Find(models)
		errs := db.GetErrors()
		if len(errs) > 0 {
			err := fmt.Errorf("gorm db err: err=%v", errs)
			return "", "", err
		}

		model_names = append(model_names, models.Name)
	}

	qps_list := strings.Split(qps, ",")
	if len(model_names) != len(qps_list) {
		return "","", fmt.Errorf("model and qps is not fit")
	}

	var model_names_list string
	for _, model := range model_names {
		model_names_list += (model + ",")
	}

	model_names_list = strings.TrimRight(model_names_list, ",")
	return model_names_list, qps, nil
}

func (service *P2PModelService) DoRouterModeJob(env *env.Env, dbService *schema.Service) error {
	global_model_service_map, err := service.fetchDbGlobalModelServiceMap(env)
	if err != nil {
		return &common.TagError{
			fmt.Sprintf("fetchDbGlobalModelServiceMap() err: %v, DB Data: %v", err, common.Pretty(global_model_service_map)),
			metrics.TAG_FETCH_DB_GLOBAL_MODEL_SERVICE_MAP_ERROR,
		}
	}

	err = service.updateGlobalModelServiceMap(env, global_model_service_map)
	if err != nil {
		return &common.TagError{
			fmt.Sprintf("updateGlobalModelServiceMap() err: %v", err),
			metrics.TAG_UPDATE_GLOBAL_MODEL_SERVICE_MAP_ERROR,
		}
	}

	err = logics.SetPredictorWorkMode(common.PredictorRouterMode, env, env.LocalIp, env.Conf.P2PModelService.TargetService.HttpPort)
	if err != nil {
		return &common.TagError{
			fmt.Sprintf("SetPredictorWorkMode %v err: %v", common.PredictorRouterMode, err),
			metrics.TAG_SET_PREDICTOR_WORK_MODE_ERROR,
		}
	}
	db_service_weight, err := logics.GetServiceWeight(env, env.LocalIp)
	if err != nil {
		return &common.TagError{
			fmt.Sprintf("getServiceWeight() err: %v, DB Data: %v", err.Error(), common.Pretty(db_service_weight)),
			metrics.TAG_GETWEIGHT_ERROR,
		}
	}
	logger.Debugf("\n>> Service Weight: %v\n\n", common.Pretty(db_service_weight))

	// get service config parameters from db
	service_names := []string{common.PredictorRouterServiceName}
	service_configs, err_config := logics.GetServiceConfig(env, service_names)
	if err_config != nil {
		return &common.TagError{
			fmt.Sprintf("getServiceConfig() err: %v, DB Data: %v", err_config.Error(), common.Pretty(service_configs)),
			metrics.TAG_GETCONFIG_ERROR,
		}
	}
	logger.Debugf("\n>> Service Config: %v\n\n", common.Pretty(service_configs))

	service_model_map := make(map[string]map[string]schema.ModelHistory)
	service_model_map[common.PredictorRouterServiceName] = make(map[string]schema.ModelHistory)
	if err := logics.NotifyPredictor(service_model_map, db_service_weight, service_configs, env, env.LocalIp, env.Conf.P2PModelService.TargetService.HttpPort); err != nil {
		return &common.TagError{
			fmt.Sprintf("notifyPredictor() err: %v, DB Data: %v, Service Weight: %v",
				err.Error(), common.Pretty(service_model_map), common.Pretty(db_service_weight)),
			metrics.TAG_NOTIFY_ERROR,
		}
	}
	return nil
}