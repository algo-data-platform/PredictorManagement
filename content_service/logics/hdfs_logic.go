package logics

import (
	"content_service/env"
	"content_service/libs/logger"
	"content_service/schema"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
)

const (
	HdfsPrefix       = "hdfs:"
	HADOOP_USER_NAME = "adbot"
)

// 根据desc 前缀==hdfs: 规则获取要拉取的hdfs模型
func FetchHdfsModelsFromDb(env *env.Env) ([]schema.ModelHistory, error) {
	db := env.Db
	var modelHistoris []schema.ModelHistory
	likeParam := HdfsPrefix + "%"
	dbPtr := db.Where("`desc` LIKE ?", likeParam).Find(&modelHistoris)
	if errs := dbPtr.GetErrors(); len(errs) > 0 {
		return modelHistoris, fmt.Errorf("gorm db err: likeParam=%s err=%v", likeParam, errs)
	}
	return modelHistoris, nil
}

func InitHadoopUserNameEnv() error {
	err := os.Setenv("HADOOP_USER_NAME", HADOOP_USER_NAME)
	if err != nil {
		return fmt.Errorf("InitHadoopUserNameEnv failed, err : %v, HADOOP_USER_NAME: %s", err, HADOOP_USER_NAME)
	}
	return nil
}

// 拉取hdfs模型文件到本地
func PullHdfsModelFile(hdfsPath string, destPath string) error {
	if hdfsPath == "" || destPath == "" {
		return fmt.Errorf("hdfsPath or destPath is empty")
	}
	cmd := exec.Command("hadoop", "fs", "-copyToLocal", hdfsPath, destPath)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("hadoop cmd copyToLocal failed, err : %v, stdout: %v", err, string(stdout))
	}
	return nil
}

// 推送模型到中转机，文件已经存在不更新
func PushModelToTransmit(env *env.Env, modelVersionName string) error {
	if modelVersionName == "" {
		return fmt.Errorf("modelVersionName is empty")
	}
	srcPath := path.Join(env.Conf.HdfsService.DestPath, modelVersionName)
	transmitPath := env.Conf.HdfsService.TransmitHost + "::" + path.Join(env.Conf.HdfsService.TransmitPath)
	cmd := exec.Command("/bin/rsync", "-rput", "--bwlimit="+strconv.Itoa(env.Conf.HdfsService.RsyncBWLimit), srcPath, transmitPath)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PushModelToTransmit failed, err : %v, modelVersionName: %s, stdout: %v", err, modelVersionName, string(stdout))
	}
	return nil
}

// 更新modelHistory desc 为空，进入一致性验证
func UpdateModelHistoryStatusById(env *env.Env, mhid uint, desc string) error {
	db := env.Db
	sql := fmt.Sprintf("update model_histories set `desc`='%s' where id = %d", desc, mhid)
	logger.Debugf("sql: %v ", sql)
	err := db.Exec(sql).Error
	if err != nil {
		return fmt.Errorf("UpdateModelHistoryStatusById failed, id: %d, sql: %s, err: %+v", mhid, sql, err)
	}
	return nil
}
