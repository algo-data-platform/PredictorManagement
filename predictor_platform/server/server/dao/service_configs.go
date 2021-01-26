package dao

import (
	"fmt"
	"server/libs/logger"
	"server/schema"
)

// 更新service_configs表
func UpdateServiceConfigData(id uint, sc_map map[string]interface{}) bool {
	errs := Mysql_db.Model(&schema.ServiceConfig{}).Where("id = ?", id).Update(sc_map).GetErrors()
	if len(errs) != 0 {
		logger.Errorf("update table host_service error, err: %v", errs)
		return false
	}
	return true
}

func CreateServiceConfig(sc *schema.ServiceConfig) error {
	if errs := Mysql_db.Create(sc).GetErrors(); len(errs) != 0 {
		return fmt.Errorf("insert record into host_service fail, err: %v", errs)
	}
	return nil
}

// 根据主键获取config_service数据
func GetServiceConfigById(id uint) (schema.ServiceConfig, error) {
	var service_config = schema.ServiceConfig{}
	db := Mysql_db.Where(schema.ServiceConfig{ID: id}).First(&service_config)
	if db.RecordNotFound() {
		return service_config, nil
	}
	errs := db.GetErrors()
	if len(errs) != 0 {
		return service_config, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return service_config, nil
}
