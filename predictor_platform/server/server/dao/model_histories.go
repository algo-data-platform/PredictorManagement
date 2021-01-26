package dao

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"server/schema"
)

// 获取指定时间内模型更新历史
func GetModelWeeklyInfo(modelName string, weekTimestamp string) ([]schema.ModelHistory, error) {
	var modelWeeklyInfo = []schema.ModelHistory{}
	var db_ptr *gorm.DB
	db_ptr = Mysql_db.Order("timestamp desc").Where(
		"model_name = ? AND timestamp >= ? AND `desc` = 'Validated' ",
		modelName, weekTimestamp).Find(&modelWeeklyInfo)
	if db_ptr == nil {
		return modelWeeklyInfo, fmt.Errorf("mysql db ptr nil, please check")
	}
	if db_ptr.RecordNotFound() {
		return modelWeeklyInfo, fmt.Errorf("model_histories table not found")
	} else if err := db_ptr.GetErrors(); len(err) != 0 {
		return modelWeeklyInfo, fmt.Errorf("get data from table model_histories error: %v", err)
	}
	return modelWeeklyInfo, nil
}

func CreateModelHistory(modelHistory *schema.ModelHistory) error {
	if errs := Mysql_db.Create(&modelHistory).GetErrors(); len(errs) != 0 {
		return fmt.Errorf("create table fail, err:%v", errs)
	}
	return nil
}

func GetLatestValidModelHistory(modelName string) (*schema.ModelHistory, error) {
	modelHistory := &schema.ModelHistory{}
	dbPtr := Mysql_db.Where("model_name = ? AND `desc`='Validated'", modelName).Order("timestamp desc").First(modelHistory)
	if dbPtr == nil {
		return modelHistory, fmt.Errorf("mysql db ptr nil, please check")
	}
	if dbPtr.RecordNotFound() {
		return modelHistory, dbPtr.Error
	} else if err := dbPtr.GetErrors(); len(err) != 0 {
		return modelHistory, fmt.Errorf("get data from table model_histories error: %v", err)
	}
	return modelHistory, nil
}

func GetLockedValidModelHistory(modelName string) (*schema.ModelHistory, error) {
	modelHistory := &schema.ModelHistory{}
	dbPtr := Mysql_db.Where("is_locked = 1 AND model_name = ? AND `desc`='Validated'", modelName).First(modelHistory)
	if dbPtr == nil {
		return modelHistory, fmt.Errorf("mysql db ptr nil, please check")
	}
	if dbPtr.RecordNotFound() {
		return modelHistory, dbPtr.Error
	} else if err := dbPtr.GetErrors(); len(err) != 0 {
		return modelHistory, fmt.Errorf("get data from table model_histories error: %v", err)
	}
	return modelHistory, nil
}
