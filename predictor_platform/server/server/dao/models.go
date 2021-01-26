package dao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"server/libs/logger"
	"server/schema"
	"server/util"
	"strings"
)

func GetModelType(model_name string) string {
	// 获取model name的模型分类
	var model schema.Model
	db_ptr := Mysql_db.Where("name = ? ", model_name).Find(&model)
	if db_ptr == nil {
		logger.Errorf("mysql db ptr nil, please check")
		return ""
	}
	if db_ptr.RecordNotFound() {
		logger.Errorf("models table not found")
		return ""
	} else if err := db_ptr.GetErrors(); len(err) != 0 {
		logger.Errorf("get data from table models error: %v", err)
		return ""
	}
	return model.Desc
}

func GetModelByName(model_name string) (*schema.Model, error) {
	var model = &schema.Model{}
	db_ptr := Mysql_db.Where("name = ? ", model_name).First(model)
	if db_ptr == nil {
		logger.Errorf("mysql db ptr nil, please check")
		return model, fmt.Errorf("mysql db ptr nil, please check")
	}
	if db_ptr.RecordNotFound() {
		return model, nil
	} else if errs := db_ptr.GetErrors(); len(errs) != 0 {
		logger.Errorf("get data from table models error: %v", errs)
		return model, fmt.Errorf("gorm db err: %v", errs)
	}
	return model, nil
}

func UpdateModelExtension(model_name string, extension *string, extension_update string) bool {
	if len(*extension) == 0 {
		*extension = "{}"
	}
	var extension_map_update map[string]interface{}
	json.NewDecoder(strings.NewReader(extension_update)).Decode(&extension_map_update)
	logger.Infof("update:%v", extension_map_update)
	var model_record schema.Model
	if Mysql_db.Where("name = ?", model_name).First(&model_record).RecordNotFound() {
		return false
	}
	logger.Infof("database_json:%v", model_record.Extension)
	extension_map_database := make(map[string]interface{})
	json.NewDecoder(strings.NewReader(model_record.Extension)).Decode(&extension_map_database)
	for i := range extension_map_update {
		extension_map_database[i] = extension_map_update[i]
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.Encode(extension_map_database)
	*extension = strings.TrimSpace(string(buf.Bytes()))
	return true
}

// 获取所有模型数据
func GetAllModels() ([]schema.Model, error) {
	var models = []schema.Model{}
	db := Mysql_db.Find(&models)
	errs := db.GetErrors()
	if len(errs) != 0 {
		return models, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return models, nil
}

func GetModelExtensionMap(modelList []string) (map[string]string, error) {
	var modelExtensionMap = make(map[string]string)
	var modelExtensions = []util.ModelExtension{}
	if len(modelList) == 0 {
		return modelExtensionMap, fmt.Errorf("modelList is empty")
	}
	var modelsStr string
	for idx, modelName := range modelList {
		modelsStr = modelsStr + fmt.Sprintf("'%s'", modelName)
		if idx != (len(modelList) - 1) {
			modelsStr = modelsStr + ","
		}
	}
	sql := `SELECT name, extension 
		FROM models
		WHERE name IN (` + modelsStr + `)`
	db := Mysql_db.Raw(sql).Scan(&modelExtensions)

	errs := db.GetErrors()
	if len(errs) != 0 {
		return modelExtensionMap, fmt.Errorf("gorm db err: err=%v, sql=%s, modelList=%v", errs, sql, modelList)
	}
	for _, modelExtension := range modelExtensions {
		modelExtensionMap[modelExtension.Name] = modelExtension.Extension
	}
	return modelExtensionMap, nil
}

func GetAllModelExtensions() ([]util.ModelExtension, error) {
	var modelExtensions = []util.ModelExtension{}
	sql := `SELECT name, extension 
		FROM models
		WHERE extension != ''`
	db := Mysql_db.Raw(sql).Scan(&modelExtensions)

	errs := db.GetErrors()
	if len(errs) != 0 {
		return modelExtensions, fmt.Errorf("gorm db err: err=%v, sql=%s", errs, sql)
	}

	return modelExtensions, nil
}

func CreateModel(model *schema.Model) error {
	if errs := Mysql_db.Create(model).GetErrors(); len(errs) != 0 {
		return fmt.Errorf("insert record into  table fail, err: %v", errs)
	}
	return nil
}

func ExistsModel(modelName string) (bool, error) {
	var model schema.Model
	notFound := Mysql_db.Where("name = ?", modelName).First(&model).RecordNotFound()
	if errs := Mysql_db.GetErrors(); len(errs) != 0 {
		return false, fmt.Errorf("find record from table fail, err : %v", errs)
	}
	return !notFound, nil
}

// 批量获取模型记录
func GetModelMapByIds(mids []uint) (map[uint]*schema.Model, error) {
	modelMap := make(map[uint]*schema.Model, 0)
	models := []*schema.Model{}
	if len(mids) == 0 {
		return modelMap, nil
	}
	var midsStr string
	for idx, mid := range mids {
		midsStr = midsStr + fmt.Sprintf("%d", mid)
		if idx != (len(mids) - 1) {
			midsStr = midsStr + ","
		}
	}
	sql := `SELECT * FROM models
		WHERE id IN (` + midsStr + `)`
	db := Mysql_db.Raw(sql).Scan(&models)
	if errs := db.GetErrors(); len(errs) != 0 {
		return modelMap, fmt.Errorf("gorm db err: err=%v, sql=%s, hids=%v", errs, sql, mids)
	}
	for _, model := range models {
		modelMap[model.ID] = model
	}
	return modelMap, nil
}
