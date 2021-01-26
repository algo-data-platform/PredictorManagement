// 综合表操作
package dao

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"server/conf"
	"server/libs/logger"
	"server/schema"
	"strconv"
)

//mysql 查询操作
func ShowTableData(table_name string) []byte {
	var ret_data []byte
	switch table_name {
	case "hosts":
		var rows []schema.Host
		db_ptr := Mysql_db.Find(&rows)
		if db_ptr == nil {
			logger.Errorf("mysql db ptr nil, please check")
			return ret_data
		}
		if db_ptr.RecordNotFound() {
			logger.Errorf("db's table not found: %s", table_name)
			return ret_data
		} else if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", table_name, err)
			return ret_data
		}
		host_list_info_marshal, err := json.Marshal(rows)
		if err != nil {
			logger.Errorf("marshal hosts json error: %v", err)
		}
		return host_list_info_marshal
	case "services":
		var rows []schema.Service
		db_ptr := Mysql_db.Find(&rows)
		if db_ptr == nil {
			logger.Errorf("mysql db ptr nil, please check")
			return ret_data
		}
		if db_ptr.RecordNotFound() {
			logger.Errorf("db's table not found: %s", table_name)
			return ret_data
		} else if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", table_name, err)
			return ret_data
		}
		service_list_marshal, err := json.Marshal(rows)
		if err != nil {
			logger.Errorf("marsharl services json error: %v", err)
		}
		return service_list_marshal
	case "models":
		var rows []schema.Model
		db_ptr := Mysql_db.Find(&rows)
		if db_ptr == nil {
			logger.Errorf("mysql db ptr nil, please check")
			return ret_data
		}
		if db_ptr.RecordNotFound() {
			logger.Errorf("db's record of table not find: %s", table_name)
			return ret_data
		} else if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", table_name, err)
			return ret_data
		}
		models_list_marshal, err := json.Marshal(rows)
		if err != nil {
			logger.Errorf("marshal models json error: %v", err)
		}
		return models_list_marshal
	case "host_services":
		// hosts, service
		var hosts []schema.Host
		db_ptr := Mysql_db.Find(&hosts)
		if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", "hosts", err)
			return ret_data
		}
		var services []schema.Service
		db_ptr = Mysql_db.Find(&services)
		if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", "services", err)
			return ret_data
		}
		var hostServices []schema.HostService
		db_ptr = Mysql_db.Find(&hostServices)
		if db_ptr == nil {
			logger.Errorf("mysql db ptr nil, please check")
			return ret_data
		}
		if db_ptr.RecordNotFound() {
			logger.Errorf("db's table not found: %s", table_name)
			return ret_data
		} else if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", table_name, err)
			return ret_data
		}
		type HostServiceDetail struct {
			HostServices []schema.HostService `json:"host_services"`
			Hosts        []schema.Host        `json:"hosts"`
			Services     []schema.Service     `json:"services"`
		}
		var hostServiceDetail = &HostServiceDetail{
			HostServices: hostServices,
			Hosts:        hosts,
			Services:     services,
		}
		host_services_list_marshal, err := json.Marshal(hostServiceDetail)
		if err != nil {
			logger.Errorf("marshal host_services json error: %v", err)
		}
		return host_services_list_marshal
	case "service_models":
		var services []schema.Service
		db_ptr := Mysql_db.Find(&services)
		if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", "services", err)
			return ret_data
		}
		var models []schema.Model
		db_ptr = Mysql_db.Find(&models)
		if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", "models", err)
			return ret_data
		}
		var serviceModels []schema.ServiceModel
		db_ptr = Mysql_db.Find(&serviceModels)
		if db_ptr == nil {
			logger.Errorf("mysql db ptr nil, please check")
			return ret_data
		}
		if db_ptr.RecordNotFound() {
			logger.Errorf("db's table not find: %s", table_name)
			return ret_data
		} else if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", table_name, err)
			return ret_data
		}
		type ServiceModelDetail struct {
			ServiceModels []schema.ServiceModel `json:"service_models"`
			Models        []schema.Model        `json:"models"`
			Services      []schema.Service      `json:"services"`
		}
		var serviceModelDetail = &ServiceModelDetail{
			ServiceModels: serviceModels,
			Models:        models,
			Services:      services,
		}
		service_models_list_marshal, err := json.Marshal(serviceModelDetail)
		if err != nil {
			logger.Errorf("marshal service_models json error: %v", err)
		}
		return service_models_list_marshal
	case "model_histories":
		var mhs []schema.ModelHistory
		db_ptr := Mysql_db.Find(&mhs)
		if db_ptr == nil {
			logger.Errorf("mysql db ptr nil, please check")
			return ret_data
		}
		if db_ptr.RecordNotFound() {
			logger.Errorf("db's table not found: %s", table_name)
			return ret_data
		} else if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", table_name, err)
			return ret_data
		}
		model_histories_list_marshal, err := json.Marshal(mhs)
		if err != nil {
			logger.Errorf("marshal model_histories json error: %v", err)
		}
		return model_histories_list_marshal
	case "configs":
		var rows []schema.Config
		db_ptr := Mysql_db.Find(&rows)
		if db_ptr == nil {
			logger.Errorf("mysql db ptr nil, please check")
			return ret_data
		}
		if db_ptr.RecordNotFound() {
			logger.Errorf("db's table not found: %s", table_name)
			return ret_data
		} else if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", table_name, err)
			return ret_data
		}
		config_list_marshal, err := json.Marshal(rows)
		if err != nil {
			logger.Errorf("marsharl configs json error: %v", err)
		}
		return config_list_marshal
	case "service_configs":
		var rows []schema.ServiceConfig
		db_ptr := Mysql_db.Find(&rows)
		if db_ptr == nil {
			logger.Errorf("mysql db ptr nil, please check")
			return ret_data
		}
		if db_ptr.RecordNotFound() {
			logger.Errorf("db's table not found: %s", table_name)
			return ret_data
		} else if err := db_ptr.GetErrors(); len(err) != 0 {
			logger.Errorf("get data from table: %s error: %v", table_name, err)
			return ret_data
		}
		service_config_list_marshal, err := json.Marshal(rows)
		if err != nil {
			logger.Errorf("marshal service_configs json error: %v", err)
		}
		return service_config_list_marshal
	default:
		logger.Warnf("show table not exist!")
	}
	return ret_data
}

//mysql 删除操作
func DeleteTableData(table_name string, context *gin.Context, conf *conf.Conf) bool {
	switch table_name {
	case "hosts":
		err := Mysql_db.Where("ip = ?", context.Query("ip")).Delete(&schema.Host{}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("delete table error: %v", err)
			return false
		}
	case "services":
		err := Mysql_db.Where("name = ?", context.Query("name")).Delete(&schema.Service{}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("delete table error: %v", err)
			return false
		}
	case "models":
		err := Mysql_db.Where("name = ?", context.Query("name")).Delete(&schema.Model{}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("delete table error: %v", err)
			return false
		}
	case "host_services":
		err := Mysql_db.Where("hid = ? and sid = ?", context.Query("hid"),
			context.Query("sid")).Delete(&schema.HostService{}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("delete table error: %v", err)
			return false
		}
		return true
	case "service_models":
		err := Mysql_db.Where("sid = ? and mid = ? ", context.Query("sid"),
			context.Query("mid")).Delete(&schema.ServiceModel{}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("delete table error: %v", err)
			return false
		}
	case "model_histories":
		err := Mysql_db.Where("id = ?", context.Query("id")).Delete(&schema.ModelHistory{}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("delete table error: %v", err)
			return false
		}
	case "configs":
		err := Mysql_db.Where("id = ?", context.Query("id")).Delete(&schema.Config{}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("delete table error: %v", err)
			return false
		}
	case "service_configs":
		err := Mysql_db.Where("sid = ? and cid = ?", context.Query("sid"), context.Query("cid")).Delete(&schema.ServiceConfig{}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("delete table error: %v", err)
			return false
		}
	default:
		logger.Warnf("delete table data not exist!")
		return false
	}
	return true
}

// mysql修改操作
func UpdateTableData(table_name string, context *gin.Context, conf *conf.Conf) bool {
	switch table_name {
	case "hosts":
		err := Mysql_db.Model(&schema.Host{}).Where("ip = ?", context.Query("ip")).Update(&schema.Host{
			Ip:         context.Query("ip"),
			DataCenter: context.Query("data_center"),
			Desc:       context.Query("desc"),
		}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("update table data error,table_name: %s,err: %v", table_name, err)
			return false
		}
	case "services":
		err := Mysql_db.Model(&schema.Service{}).Where("name = ?", context.Query("name")).Update(&schema.Service{
			Name: context.Query("name"),
			Desc: context.Query("desc"),
		}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("update table data error,table_name: %s,err: %v", table_name, err)
			return false
		}
	case "models":
		extension := context.Query("extension")
		extension_update := context.Query("extension_update")
		if len(extension_update) != 0 {
			ret := UpdateModelExtension(context.Query("name"), &extension, extension_update)
			if ret == false {
				return false
			}
			logger.Infof("updated_extension:%v", extension)
		}
		err := Mysql_db.Model(&schema.Model{}).Where("name = ?", context.Query("name")).Update(&schema.Model{
			Name:      context.Query("name"),
			Path:      context.Query("path"),
			Desc:      context.Query("desc"),
			Extension: extension,
		}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("update table data error,table_name: %s,err: %v", table_name, err)
			return false
		}
	case "model_histories":
		var is_locked uint64
		var err error
		is_locked_query := context.Query("is_locked")
		if len(is_locked_query) != 0 {
			is_locked, err = strconv.ParseUint(is_locked_query, 10, 64)
			if err != nil {
				logger.Errorf("parse is_locked error: %v", err)
				return false
			}
		} else {
			is_locked = 0
		}
		errs := Mysql_db.Model(&schema.ModelHistory{}).Where("id = ?", context.Query("id")).Update(map[string]interface{}{
			"model_name": context.Query("model_name"),
			"timestamp":  context.Query("timestamp"),
			"md5":        context.Query("md5"),
			"is_locked":  uint(is_locked),
			"desc":       context.Query("desc"),
		}).GetErrors()
		if len(errs) != 0 {
			logger.Errorf("update table model_histories error,table_name: %s,err: %v", table_name, err)
			return false
		}
	case "configs":
		err := Mysql_db.Model(&schema.Config{}).Where("id = ?", context.Query("id")).Update(&schema.Config{
			Description: context.Query("desc"),
			Config:      context.Query("config"),
		}).GetErrors()
		if len(err) != 0 {
			logger.Errorf("update table data error,table_name: %s,err: %v", table_name, err)
			return false
		}
	default:
		logger.Warnf("update table data not exists!")
		return false
	}
	return true
}
