package dao

import (
	"fmt"
	"server/libs/logger"
	"server/schema"
)

// 获取service 和 model 的映射关系
// 返回 map[service_name]model_list
func GetServiceModelMap() map[string][]string {
	serviceModelMap := make(map[string][]string, 0)
	sql := `select s.name as service_name,m.name as model_name from service_models sm
		inner join services s on sm.sid=s.id 
		inner join models m on sm.mid = m.id;`
	rows, err := Mysql_db.Raw(sql).Rows()
	if err != nil {
		logger.Errorf("read db fail, sql=%s, err=%v", sql, err)
		return serviceModelMap
	}
	// scan rows
	for rows.Next() {
		var (
			serviceName string
			modelName   string
		)
		if err := rows.Scan(&serviceName, &modelName); err != nil {
			logger.Errorf("rows.Scan fail, err=%v", err)
			continue
		}

		if _, ok := serviceModelMap[serviceName]; !ok {
			serviceModelMap[serviceName] = []string{modelName}
		} else {
			serviceModelMap[serviceName] = append(serviceModelMap[serviceName], modelName)
		}
	}
	return serviceModelMap
}

func InsertServiceModel(sid uint, mid uint) error {
	// 增加service model表
	service_model := schema.ServiceModel{
		Sid:  sid,
		Mid:  mid,
		Desc: "",
	}
	errs := Mysql_db.Create(&service_model).GetErrors()
	if len(errs) != 0 {
		logger.Errorf("insert record into table fail: %v", errs)
		return fmt.Errorf("gorm db err: err=%v", errs)
	}
	return nil
}

func DelServiceModelById(sid uint) error {
	errs := Mysql_db.Where("sid = ?", sid).Delete(&schema.ServiceModel{}).GetErrors()
	if len(errs) != 0 {
		logger.Errorf("delHostServiceById id: %d, err: %+v", sid, errs)
		return fmt.Errorf("delHostServiceById id: %d, err: %+v", sid, errs)
	}
	return nil
}

func CreateServiceModel(serviceModel *schema.ServiceModel) error {
	if errs := Mysql_db.Create(&serviceModel).GetErrors(); len(errs) != 0 {
		logger.Errorf("insert record into table fail, err: %v", errs)
		return fmt.Errorf("insert record into table fail, err: %v", errs)
	}
	return nil
}

func ExistsServiceModelBySidMid(sid uint, mid uint) (bool, error) {
	var serviceModel schema.ServiceModel
	notFound := Mysql_db.Where("sid = ? AND mid = ?", sid, mid).First(&serviceModel).RecordNotFound()
	if errs := Mysql_db.GetErrors(); len(errs) != 0 {
		logger.Errorf("find record from table fail, err : %v", errs)
		return false, fmt.Errorf("find record from table fail, err : %v", errs)
	}
	return !notFound, nil
}

// 获取service下的模型列表
func GetModelsBySid(sid uint) ([]schema.Model, error) {
	var models []schema.Model
	sql := `select m.id,m.name from service_models sm 
		inner join models m on sm.mid=m.id 
		where sm.sid = ? ;`
	db := Mysql_db.Raw(sql, sid).Scan(&models)

	errs := db.GetErrors()
	if len(errs) != 0 {
		logger.Errorf("gorm db err, sql=%s, sid=%d, err=%v", sql, sid, errs)
		return []schema.Model{}, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return models, nil
}
