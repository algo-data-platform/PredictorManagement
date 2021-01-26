package dao

import (
	"fmt"
	"server/schema"
)

func GetAllConfigs() ([]schema.Config, error) {
	var configs []schema.Config
	db := Mysql_db.Find(&configs)
	if db.RecordNotFound() {
		return nil, fmt.Errorf("Config is empty")
	}
	errs := db.GetErrors()
	if len(errs) != 0 {
		return nil, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return configs, nil
}

// 获取所有config_map，key为description, value为cid
// @return map[description]cid
func GetAllConfigMap() (map[string]uint, error) {
	var configs []schema.Config
	db := Mysql_db.Find(&configs)
	if db.RecordNotFound() {
		return nil, fmt.Errorf("Config is empty")
	}
	errs := db.GetErrors()
	if len(errs) != 0 {
		return nil, fmt.Errorf("gorm db err: err=%v", errs)
	}
	config_map := make(map[string]uint, len(configs))
	for _, config := range configs {
		config_map[config.Description] = config.ID
	}
	return config_map, nil
}

// 判断config是否存在
func ExistsConfig(config_desc string) (bool, error) {
	if config_desc == "" {
		return false, fmt.Errorf("config_desc is empty")
	}
	notFound := Mysql_db.Where(schema.Config{Description: config_desc}).Find(&schema.Config{}).RecordNotFound()
	errs := Mysql_db.GetErrors()
	if len(errs) > 0 {
		err := fmt.Errorf("gorm db err: err=%v", errs)
		return false, err
	}
	return !notFound, nil
}

// 获取config 根据cid
func GetConfigById(id uint) (*schema.Config, error) {
	config := &schema.Config{}
	db := Mysql_db.Where(schema.Config{ID: id}).First(config)
	if db.RecordNotFound() {
		return config, db.Error
	}
	errs := db.GetErrors()
	if len(errs) != 0 {
		return config, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return config, nil
}

func CreateConfig(config *schema.Config) error {
	if errs := Mysql_db.Create(config).GetErrors(); len(errs) != 0 {
		return fmt.Errorf("insert record into table fail, err: %v", errs)
	}
	return nil
}

func GetConfigByName(config_name string) (*schema.Config, error) {
	config := &schema.Config{}
	if errs := Mysql_db.Where("name = ?", config_name).First(config).GetErrors(); len(errs) != 0 {
		return config, fmt.Errorf("GetConfigByName fail, err: %v", errs)
	}
	return config, nil
}
