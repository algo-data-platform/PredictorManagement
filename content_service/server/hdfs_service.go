package server

import (
	"content_service/common"
	"content_service/env"
	"content_service/libs/logger"
	"content_service/logics"
	"content_service/metrics"
	"content_service/schema"
	"fmt"
	"path"
	"sync"
	"time"
)

type HdfsService struct {
}

func NewHdfsService() *HdfsService {
	return &HdfsService{}
}

func (service *HdfsService) Run(env *env.Env) {
	logger.Infof("ip=%v: Hdfs Service running...", env.LocalIp)
	// 初始化设置adbot的环境变量
	err := logics.InitHadoopUserNameEnv()
	if err != nil {
		logger.Fatalf("InitHadoopUserNameEnv fail, err: %v", err)
	}
	service.routineCheck(env)
}

// run check() periodically
func (service *HdfsService) routineCheck(env *env.Env) {
	for t := range time.Tick(env.Conf.HdfsService.RunInterval * time.Second) {
		ts := metrics.GetTimers()[metrics.TIMER_HDFS_SERVICE_CHECK_TIMER].Start()
		if err := service.check(env); err != nil {
			msg := fmt.Sprintf("Hdfs Service (%s) error: %s\n", env.LocalIp, err.Error())
			logger.Errorf("%v %s", t, msg)
			if tagErr, ok := err.(*common.TagError); ok {
				metrics.GetErrorMeter(metrics.TAG_HDFS_SERVICE, tagErr.ErrTag).Mark(1)
			}
			metrics.GetErrorMeter(metrics.TAG_HDFS_SERVICE, metrics.TAG_CHECK_ERROR).Mark(1)
			common.Alert(
				env.Conf.Alert.Recipients,
				fmt.Sprintf("Hdfs Service (%s) error!", env.LocalIp),
				msg,
				env.Conf.Alert.Rate)
		}
		ts.Stop()
	}
}

// single check
func (service *HdfsService) check(env *env.Env) error {
	// 1.数据库获取hdfs模型
	modelHistories, err := logics.FetchHdfsModelsFromDb(env)
	if err != nil {
		return &common.TagError{
			fmt.Sprintf("FetchHdfsModelsFromDb() err: %v", err.Error()),
			metrics.TAG_FETCHDB_ERROR,
		}
	}
	logger.Debugf("\n>> DB Hdfs Models: %v\n\n", common.Pretty(modelHistories))
	if len(modelHistories) == 0 {
		logger.Debugf("\n>> Hdfs models is not exists.\n-------------------------------------\n")
		return nil
	}

	// 2.多个gorutine执行，拉取模型及更新模型状态
	// TODO 稍后改为线程池，防止出现更多的下载任务，导致带宽打满
	var wg sync.WaitGroup
	errChan := make(chan error)
	for _, modelHistory := range modelHistories {
		wg.Add(1)
		// 拉取模型并更新状态
		go service.fetchHdfsModelAndUpdateStatus(env, &wg, modelHistory, errChan)
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
			metrics.TAG_FETCH_HDFS_AND_UPDATE_STATUS_ERROR,
		}
	}

	logger.Debugf("\n>> Hdfs pull models finished.\n-------------------------------------\n")
	return nil
}

// hdfs拉取模型及更新模型状态
func (service *HdfsService) fetchHdfsModelAndUpdateStatus(env *env.Env, wg *sync.WaitGroup, modelHistory schema.ModelHistory, errChan chan<- error) {
	defer wg.Done()
	modelVersionName := modelHistory.FullName()
	hdfsPath := modelHistory.Desc
	destPath := path.Join(env.Conf.HdfsService.DestPath, modelVersionName)
	// 1.开启hdfs拉取，先判断模型是否存在
	exists := common.IsDir(destPath)
	if exists {
		logger.Debugf("hdfs model exists, modelVersionName: %s", modelVersionName)
	} else {
		ts := metrics.GetPullSingleHdfsModelTimer(modelHistory.ModelName).Start()
		err := logics.PullHdfsModelFile(hdfsPath, destPath)
		ts.Stop()
		if err != nil {
			fmtErr := fmt.Errorf("PullHdfsModelFile err: %v, hdfsPath: %s, destPath: %s", err, hdfsPath, destPath)
			logger.Errorf("%v", fmtErr)
			errChan <- fmtErr
			return
		}
		logger.Infof("PullHdfsModelFile success, hdfsPath: %s, destPath: %s", hdfsPath, destPath)
	}

	// 2.推送模型到中转机，文件存在不更新
	err := logics.PushModelToTransmit(env, modelVersionName)
	if err != nil {
		logger.Errorf("%v", err)
		errChan <- err
		return
	}
	logger.Infof("PushModelToTransmit success, modelVersionName: %s", modelVersionName)

	// 3.更新模型状态
	desc := "" // desc置空为一致性验证开始状态
	err = logics.UpdateModelHistoryStatusById(env, modelHistory.ID, desc)
	if err != nil {
		fmtErr := fmt.Errorf("UpdateModelHistoryStatusById err: %v, id: %d", err, modelHistory.ID)
		logger.Errorf("%v", fmtErr)
		errChan <- fmtErr
		return
	}
	logger.Infof("UpdateModelHistoryStatusById success, modelVersionName: %s, id: %d", modelVersionName, modelHistory.ID)
}
