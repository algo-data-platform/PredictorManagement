package server

import (
	"content_service/common"
	"content_service/env"
	"content_service/libs/logger"
	"content_service/logics"
	"content_service/metrics"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

type CleaningService struct {
	TransmitIP       string
	OnlineModelNames []string
}

func NewCleaningService() *CleaningService {
	return &CleaningService{}
}

type DiskModel struct {
	ModelName  string
	Timestamps []string
}

func (service *CleaningService) Run(env *env.Env) {
	logger.Infof("ip=%v: Cleaning Service running...", env.LocalIp)
	service.routineCheck(env)
}

// run check() periodically
func (service *CleaningService) routineCheck(env *env.Env) {
	for t := range time.Tick(env.Conf.CleaningService.RunInterval * time.Second) {
		ts := metrics.GetTimers()[metrics.TIMER_CLEANING_SERVICE_CHECK_TIMER].Start()
		if err := service.check(env); err != nil {
			msg := fmt.Sprintf("Cleaning Service (%s) error: %v\n", env.LocalIp, err.Error())
			logger.Infof("%v %s", t, msg)
			metrics.GetErrorMeter(metrics.TAG_CLEANING_SERVICE, metrics.TAG_CHECK_ERROR).Mark(1)
			common.Alert(
				env.Conf.Alert.Recipients,
				fmt.Sprintf("Cleaning Service (%s) error!", env.LocalIp),
				msg,
				env.Conf.Alert.Rate)
		}
		ts.Stop()
	}
}

// single round of check()
func (service *CleaningService) check(env *env.Env) error {
	// 获取中转机的ip地址，轮询获取保证切换完ip后可以及时更新
	var err error
	service.TransmitIP, err = common.ResolveIP(env.Conf.P2PModelService.SrcHost)
	if err != nil {
		return err
	}
	logger.Debugf("service.TransmitIP is : %s", service.TransmitIP)
	service.OnlineModelNames, err = logics.GetOnlineModelNames(env)
	if err != nil {
		return err
	}
	var base_path string
	if env.Conf.CleaningService.PathOverride != "" {
		base_path = env.Conf.CleaningService.PathOverride
	} else {
		base_path = env.Conf.P2PModelService.DestPath // use model service path
	}
	disk_models, err := service.fetchDiskData(env, base_path)
	if err != nil {
		return err
	}
	versionsToKeep := service.getVersionsToKeep(env)
	return service.cleanByVersion(env, disk_models, base_path, versionsToKeep,
		env.Conf.CleaningService.HoursToKeep)
}

func (service *CleaningService) fetchDiskData(env *env.Env, path string) (map[string]DiskModel, error) {
	var disk_models = map[string]DiskModel{}
	var fileNames = []string{}
	if _, err := os.Stat(path); err != nil {
		return disk_models, fmt.Errorf("err in checking stat of path=%v: %v", path, err)
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return disk_models, fmt.Errorf("err in reading path=%v: %v", path, err)
	}

	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	service.getDiskModelsByFiles(disk_models, fileNames)
	return disk_models, nil
}

// 根据disk files 生成 diskModels
func (service *CleaningService) getDiskModelsByFiles(disk_models map[string]DiskModel, fileNames []string) {
	for _, fileName := range fileNames {
		s := strings.Split(fileName, common.Delimiter)
		if len(s) != 2 {
			continue
		}
		model_name := s[0]
		timestamp := s[1]
		if disk_model, found := disk_models[model_name]; found {
			disk_models[model_name] = DiskModel{ModelName: model_name, Timestamps: append(disk_model.Timestamps, timestamp)}
		} else {
			disk_models[model_name] = DiskModel{ModelName: model_name, Timestamps: []string{timestamp}}
		}
	}
}

// clean disk files by versions, in our case, version == timestamp
func (service *CleaningService) cleanByVersion(env *env.Env, disk_models map[string]DiskModel, base_path string, versionsToKeep int, hoursToKeep int) error {
	var filesToClean []string
	var err error
	for model_name, disk_model := range disk_models {
		// 对于下线模型,按照过期时间清理
		if !common.IsInSliceString(model_name, service.OnlineModelNames) {
			filesToClean, err = service.getFilesToCleanByTime(model_name, disk_model, hoursToKeep, base_path)
		} else {
			// 在线模型，按照版本清理
			filesToClean, err = service.getFilesToCleanByVersion(env, model_name, disk_model, versionsToKeep, base_path)
			if err != nil {
				return err
			}
		}
		// do real delete
		if err = common.DeleteFiles(filesToClean); err != nil {
			logger.Errorf("DeleteFiles failed, err: %v", err)
		}
	}
	return nil
}

// 获取过期版本的模型文件列表
func (service *CleaningService) getFilesToCleanByVersion(env *env.Env, model_name string, disk_model DiskModel, versionsToKeep int, base_path string) ([]string, error) {
	// form full path of files to clean
	var filesToClean = []string{}
	if versionsToKeep <= 0 {
		versionsToKeep = 1 // safe guarding to keep at least 1 version
	}
	// sort versions (ascending)
	versions := disk_model.Timestamps
	sort.Strings(versions)
	currVersionsToKeep := versionsToKeep
	if currVersionsToKeep > len(versions) {
		currVersionsToKeep = len(versions)
	}
	versionsToClean := versions[:(len(versions) - currVersionsToKeep)]
	// 如果是中转机器，判断要删除的版本里面，是否存在最新的validated版本，是则排除掉
	if env.LocalIp == service.TransmitIP && len(versionsToClean) > 0 {
		lastValidatedHistory, err := logics.GetLastValidateVersion(env, model_name)
		if err != nil {
			return filesToClean, err
		}
		if common.IsInSliceString(lastValidatedHistory.Timestamp, versionsToClean) {
			// versionsToClean为要删除的版本，将最新的validated版本从中去掉
			common.DelSliceFirstItem(&versionsToClean, lastValidatedHistory.Timestamp)
			logger.Debugf("the last validated version was found and canceled, model_name: %s, version: %s", model_name, lastValidatedHistory.Timestamp)
		}
	}
	for _, version := range versionsToClean {
		filename := path.Join(base_path, model_name+common.Delimiter+version)
		filesToClean = append(filesToClean, filename)
	}
	logger.Debugf("cleanning files for %v (versionsToKeep=%d):\n\tsorted versions: %v\n\tversionsToClean: %v\n",
		model_name, versionsToKeep, versions, versionsToClean)
	return filesToClean, nil
}

// 按照时间获取过期版本的模型文件列表
func (service *CleaningService) getFilesToCleanByTime(model_name string, disk_model DiskModel, hoursToKeep int, base_path string) ([]string, error) {
	// form full path of files to clean
	var filesToClean = []string{}
	var versionsToClean = []string{}
	// sort versions (ascending)
	versions := disk_model.Timestamps
	for _, version := range versions {
		versionTime, err := time.ParseInLocation("20060102_150405", version, time.Local)
		if err != nil {
			logger.Errorf("time.Parse version fail, err: %v, version: %s", err, version)
			versionsToClean = append(versionsToClean, version)
		} else {
			if time.Now().Sub(versionTime) > time.Hour*time.Duration(hoursToKeep) {
				versionsToClean = append(versionsToClean, version)
			}
		}
	}
	for _, version := range versionsToClean {
		filename := path.Join(base_path, model_name+common.Delimiter+version)
		filesToClean = append(filesToClean, filename)
	}
	logger.Debugf("cleanning files for %v (hoursToKeep=%d):\n\torigin versions: %v\n\tversionsToClean: %v\n",
		model_name, hoursToKeep, versions, versionsToClean)
	return filesToClean, nil
}

func (service *CleaningService) getVersionsToKeep(env *env.Env) int {
	versionsToKeep := env.Conf.CleaningService.VersionsToKeep
	if env.LocalIp == service.TransmitIP || env.LocalIp == env.Conf.ValidateService.Host {
		versionsToKeep = env.Conf.CleaningService.VersionsToKeepForValidate
	}
	return versionsToKeep
}