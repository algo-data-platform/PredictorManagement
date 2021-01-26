package dao

import (
	"fmt"
	"server/schema"
)

func GetAllServices() ([]schema.Service, error) {
	var services []schema.Service
	db := Mysql_db.Find(&services)
	if db.RecordNotFound() {
		return nil, fmt.Errorf("service is empty")
	}
	errs := db.GetErrors()
	if len(errs) != 0 {
		return nil, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return services, nil
}

// 获取所有servicemap，key为service_name, value为sid
// @return map[service_name]sid
func GetAllServiceMap() (map[string]uint, error) {
	var services []schema.Service
	db := Mysql_db.Find(&services)
	if db.RecordNotFound() {
		return nil, fmt.Errorf("service is empty")
	}
	errs := db.GetErrors()
	if len(errs) != 0 {
		return nil, fmt.Errorf("gorm db err: err=%v", errs)
	}
	serviceMap := make(map[string]uint, len(services))
	for _, service := range services {
		serviceMap[service.Name] = service.ID
	}
	return serviceMap, nil
}

// 判断服务是否存在
func ExistsService(service_name string) (bool, error) {
	if service_name == "" {
		return false, fmt.Errorf("service_name is empty")
	}
	notFound := Mysql_db.Where(schema.Service{Name: service_name}).Find(&schema.Service{}).RecordNotFound()
	errs := Mysql_db.GetErrors()
	if len(errs) > 0 {
		err := fmt.Errorf("gorm db err: err=%v", errs)
		return false, err
	}
	return !notFound, nil
}

// 获取service 根据sid
func GetServiceBySid(sid uint) (schema.Service, error) {
	var service = schema.Service{}
	db := Mysql_db.Where(schema.Service{ID: sid}).First(&service)
	if db.RecordNotFound() {
		return service, db.Error
	}
	errs := db.GetErrors()
	if len(errs) != 0 {
		return service, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return service, nil
}

func CreateService(service *schema.Service) error {
	if errs := Mysql_db.Create(service).GetErrors(); len(errs) != 0 {
		return fmt.Errorf("insert record into table fail, err: %v", errs)
	}
	return nil
}

func GetServiceByName(serviceName string) (*schema.Service, error) {
	service := &schema.Service{}
	if errs := Mysql_db.Where("name = ?", serviceName).First(service).GetErrors(); len(errs) != 0 {
		return service, fmt.Errorf("GetServiceByName fail, err: %v", errs)
	}
	return service, nil
}
