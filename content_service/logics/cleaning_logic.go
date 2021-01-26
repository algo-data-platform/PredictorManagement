package logics

import (
	"content_service/env"
	"content_service/libs/logger"
	"content_service/schema"
	"fmt"
)

// 获取最新validated 版本
func GetLastValidateVersion(env *env.Env, model_name string) (*schema.ModelHistory, error) {
	db := env.Db
	var modelHistory = &schema.ModelHistory{}
	dbPtr := db.Where(schema.ModelHistory{ModelName: model_name, Desc: "Validated"}).Last(modelHistory)
	if dbPtr.RecordNotFound() {
		return modelHistory, nil
	}
	errs := dbPtr.GetErrors()
	if len(errs) != 0 {
		logger.Errorf("GetLastValidateVersion failed, err: %v", errs)
		return modelHistory, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return modelHistory, nil
}

// 获取在线模型
func GetOnlineModelNames(env *env.Env) ([]string, error) {
	db := env.Db
	modelNames := []string{}
	sql := `SELECT name FROM models`
	dbPtr := db.Raw(sql).Pluck("name", &modelNames)
	logger.Debugf(">> sql=\"%s \"", sql)
	if errs := dbPtr.GetErrors(); len(errs) > 0 {
		return modelNames, fmt.Errorf("gorm db err: sql=%s err=%v", sql, errs)
	}
	return modelNames, nil
}
