package server

import (
	"content_service/common"
	"content_service/env"
	"content_service/libs/logger"
	"content_service/logics"
	"content_service/metrics"
	"strings"
	"fmt"
	"path"
	"sync"
	"time"
	"os/exec"
	"strconv"
)

type FileSyncService struct {
	ServiceFileSyncTimesMap   sync.Map
}

func NewFileSyncService() *FileSyncService {
	return &FileSyncService{}
}

func (service *FileSyncService) Run(env *env.Env) {
	logger.Infof("ip=%v: File Sync Service running...", env.LocalIp)
	service.routineFileSync(env)
}

// run fileSync() periodically
func (service *FileSyncService) routineFileSync(env *env.Env) {
	// file sync for the first time, don't care if successful
	service.fileSync(env);
	for t := range time.Tick(env.Conf.FileSyncService.RunInterval * time.Second) {
		if err := service.fileSync(env); err != nil {
			msg := fmt.Sprintf("File Sync Service (%s) error: %s\n", env.LocalIp, err.Error())
			logger.Errorf("%v %s", t, msg)
			if tagErr, ok := err.(*common.TagError); ok {
				metrics.GetErrorMeter(metrics.TAG_FILE_SYNC_SERVICE, tagErr.ErrTag).Mark(1)
			}
			metrics.GetErrorMeter(metrics.TAG_FILE_SYNC_SERVICE, metrics.TAG_FILE_SYNC_ERROR).Mark(1)
			common.Alert(
				env.Conf.Alert.Recipients,
				fmt.Sprintf("File Sync Service (%s) error!", env.LocalIp),
				msg,
				env.Conf.Alert.Rate)
		}
	}
}

// file sync
func (service *FileSyncService) fileSync(env *env.Env) error {
	
	err := service.syncPredictorStaticListFiles(env)
	if err != nil {
		logger.Errorf("syncPredictorStaticListFiles failed, err:%v", err);
	}

	logger.Debugf("\n>> syncPredictorStaticListFiles finished.\n-------------------------------------\n")
	return err
}

func (service *FileSyncService) syncPredictorStaticListFiles(env *env.Env) error {
	// 1.从数据库获取当前机器所有service
	dbServices, err := logics.FetchDbServices(env)
	if err != nil {
		return &common.TagError{
			fmt.Sprintf("fetchDbServices() err: %v, DB Data: %v", err, common.Pretty(dbServices)),
			metrics.TAG_FETCH_DB_SERVICES_ERROR,
		}
	}
	logger.Debugf("dbServices:%v", dbServices)
	for _, dbService := range dbServices {
		parentIP, peerNum, err := logics.GetParentIP(env, dbService, service.ServiceFileSyncTimesMap, env.Conf.FileSyncService.SyncTimesLimit, env.Conf.FileSyncService.SrcHost)
		if err != nil {
			return &common.TagError{
				fmt.Sprintf("GetParentIP err: %v, parentIP: %s", err, parentIP),
				metrics.TAG_GET_PARENT_IP_ERROR,
			}
		}
		logger.Infof("GetParentIP success, dbService: %v, parentIP: %s, peerNum: %d", dbService, parentIP, peerNum)
		err = pullPredictorStaticListFiles(env , parentIP, peerNum)
		if err != nil {
			if currentTimesI, exists := service.ServiceFileSyncTimesMap.Load(dbService.Name); !exists {
				service.ServiceFileSyncTimesMap.Store(dbService.Name, int32(1))
			} else {
				currentTimes, _ := currentTimesI.(int32)
				service.ServiceFileSyncTimesMap.Store(dbService.Name, currentTimes+int32(1))
			}
			return &common.TagError{
				fmt.Sprintf("pullPredictorStaticListFiles err: %v, parentIP: %s", err, parentIP),
				metrics.TAG_PULL_PREDICTOR_STATIC_LIST_ERROR,
			}
		}
		service.ServiceFileSyncTimesMap.Store(dbService.Name, int32(0))
	}
	return nil
}

// pull predictor static list file from transfer host, with retries
func pullPredictorStaticListFiles(env *env.Env, parentIP string, peerNum int) error {
	for i:=0; i <= env.Conf.FileSyncService.RetryTimes; i++ {
		destPath := path.Join(env.Conf.FileSyncService.DestPath, env.Conf.FileSyncService.PredictorStaticListDir)
		peerNum = common.MaxInt(1, peerNum)
		bwLimit := common.MinInt(int(env.Conf.FileSyncService.SrcRsyncBWLimit/peerNum), env.Conf.FileSyncService.RsyncBWLimit)
		file_source := parentIP+"::"+env.Conf.FileSyncService.SrcPath+"/"+env.Conf.FileSyncService.PredictorStaticListDir+"/"
		cmd := exec.Command("/bin/rsync", "-avz", "--bwlimit="+strconv.Itoa(bwLimit), file_source, destPath+"/")
		logger.Debugf("cmd: %s", "/bin/rsync -avz --bwlimit="+strconv.Itoa(bwLimit)+" "+file_source+" "+destPath+"/")
		if stdout, err := cmd.CombinedOutput(); err != nil {
			if strings.Contains(string(stdout), "No such file or directory") {
				logger.Debugf("rsync predictor static list not exists, err=%v, output=%s", err, stdout)
				break
			}
			logger.Errorf("rsync cmd failed: err=%v, output=%s", err, stdout)
			continue
		} else {
			return nil
		}
	}
	
	logger.Errorf("pull predictor static list failed after retried %v times", env.Conf.P2PModelService.Retry)
	return fmt.Errorf("pull predictor static list failed after retried %v times", env.Conf.P2PModelService.Retry)
}