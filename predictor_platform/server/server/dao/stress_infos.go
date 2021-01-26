package dao

import (
	"fmt"
	"server/libs/logger"
	"server/schema"
	"strconv"
	"strings"
)

func GetAllStressInfos() ([]*schema.StressInfo, error) {
	rows := []*schema.StressInfo{}
	db_ptr := Mysql_db.Find(&rows)
	if db_ptr == nil {
		logger.Errorf("mysql db ptr nil, please check")
		return rows, fmt.Errorf("mysql db ptr nil, please check")
	}
	if db_ptr.RecordNotFound() {
		logger.Warnf("db's record of stress_infos not find")
		return rows, nil
	} else if errs := db_ptr.GetErrors(); len(errs) != 0 {
		logger.Errorf("get data from stress_infos error: %v", errs)
		return rows, fmt.Errorf("get data from stress_infos error: %v", errs)
	}
	return rows, nil
}

func GetStressInfosByStatus(is_enable uint) ([]*schema.StressInfo, error) {
	rows := []*schema.StressInfo{}
	db_ptr := Mysql_db.Where("is_enable = ?", is_enable).Find(&rows)
	if db_ptr == nil {
		logger.Errorf("mysql db ptr nil, please check")
		return rows, fmt.Errorf("mysql db ptr nil, please check")
	}
	if db_ptr.RecordNotFound() {
		logger.Warnf("db's record of stress_infos not find")
		return rows, nil
	} else if errs := db_ptr.GetErrors(); len(errs) != 0 {
		logger.Errorf("get data from stress_infos error: %v", errs)
		return rows, fmt.Errorf("get data from stress_infos error: %v", errs)
	}
	return rows, nil
}

func CreateStressInfo(stressInfo *schema.StressInfo) error {
	if errs := Mysql_db.Create(stressInfo).GetErrors(); len(errs) != 0 {
		logger.Errorf("insert record into stress_infos fail")
		return fmt.Errorf("insert record into stress_infos fail, errs: %v", errs)
	}
	return nil
}

func TransactEnableStressTest(stressInfo *schema.StressInfo, stressTestService *schema.Service) error {
	// 开启事务
	ts := Mysql_db.Begin()

	// 1.删除test_service关联的模型
	errs := ts.Where("sid = ?", stressTestService.ID).Delete(&schema.ServiceModel{}).GetErrors()
	if len(errs) != 0 {
		ts.Rollback()
		return fmt.Errorf("delete service_model by sid fail, err: %v", errs)
	}

	// 2.压测模型关联test_service
	if stressInfo.Mids != "" {
		mids := strings.Split(stressInfo.Mids, ",")
		for _, mid := range mids {
			midUint, err := strconv.ParseUint(mid, 10, 64)
			if err != nil {
				ts.Rollback()
				return fmt.Errorf("strconv.ParseUintfail, err: %v", err)
			}
			serviceModel := &schema.ServiceModel{
				Sid:  stressTestService.ID,
				Mid:  uint(midUint),
				Desc: "",
			}
			if errs := ts.Create(serviceModel).GetErrors(); len(errs) != 0 {
				ts.Rollback()
				return fmt.Errorf("create service_model fail, err: %v", errs)
			}
		}
	}

	// 3.删除机器与原service的关联
	errs = ts.Where("hid = ?", stressInfo.Hid).Delete(&schema.HostService{}).GetErrors()
	if len(errs) != 0 {
		ts.Rollback()
		return fmt.Errorf("delete host_services by hid fail, err: %v", errs)
	}

	// 4.关联test_service
	hostService := &schema.HostService{
		Hid:        stressInfo.Hid,
		Sid:        stressTestService.ID,
		Desc:       "",
		LoadWeight: 0,
	}
	if errs := ts.Create(hostService).GetErrors(); len(errs) != 0 {
		ts.Rollback()
		return fmt.Errorf("create host_services fail, err: %v", errs)
	}

	// 5. 根据压测记录ID是否存在，更新or新增压测记录
	if stressInfo.ID != 0 {
		// 更新压测记录
		if errs := ts.Model(&schema.StressInfo{}).Where("id = ?", stressInfo.ID).Updates(
			map[string]interface{}{"is_enable": 1}).GetErrors(); len(errs) != 0 {
			ts.Rollback()
			return fmt.Errorf("update stress_info fail, err: %v", errs)
		}
	} else {
		// 新增压测记录
		if errs := ts.Create(stressInfo).GetErrors(); len(errs) != 0 {
			ts.Rollback()
			return fmt.Errorf("create stress_info fail, err: %v", errs)
		}
	}
	ts.Commit()
	return nil
}

// 事务关闭压测
func TransactDisableStressTest(stressInfo *schema.StressInfo, stressTestService *schema.Service) error {
	// 开启事务
	ts := Mysql_db.Begin()

	// 1.删除test_service关联的模型
	errs := ts.Where("sid = ?", stressTestService.ID).Delete(&schema.ServiceModel{}).GetErrors()
	if len(errs) != 0 {
		ts.Rollback()
		return fmt.Errorf("delete service_model by sid fail, err: %v", errs)
	}

	// 2.删除机器与test service的关联
	errs = ts.Where("hid = ?", stressInfo.Hid).Delete(&schema.HostService{}).GetErrors()
	if len(errs) != 0 {
		ts.Rollback()
		return fmt.Errorf("delete host_services by hid fail, err: %v", errs)
	}

	// 3.关联原来的service
	if stressInfo.OriginSids != "" {
		originSids := strings.Split(stressInfo.OriginSids, ",")
		if len(originSids) > 0 {
			for _, originSidWeight := range originSids {
				sidWeights := strings.Split(originSidWeight, "_")
				if len(sidWeights) != 2 {
					logger.Warnf("strings.Split originSidWeight len != 2, originSidWeight:%s", originSidWeight)
					continue
				}
				originSidUint, err := strconv.ParseUint(sidWeights[0], 10, 64)
				if err != nil {
					ts.Rollback()
					return fmt.Errorf("strconv.ParseUint fail, err: %v", err)
				}
				originWeightUint, err := strconv.ParseUint(sidWeights[1], 10, 64)
				if err != nil {
					ts.Rollback()
					return fmt.Errorf("strconv.ParseUint fail, err: %v", err)
				}
				hostService := &schema.HostService{
					Hid:        stressInfo.Hid,
					Sid:        uint(originSidUint),
					Desc:       "",
					LoadWeight: uint(originWeightUint),
				}
				if errs := ts.Create(hostService).GetErrors(); len(errs) != 0 {
					ts.Rollback()
					return fmt.Errorf("create host_services fail, err: %v", errs)
				}
			}
		}
	}

	// 4. 更新压测状态
	if errs := ts.Model(&schema.StressInfo{}).Where("id = ?", stressInfo.ID).Updates(
		map[string]interface{}{"is_enable": 0}).GetErrors(); len(errs) != 0 {
		ts.Rollback()
		return fmt.Errorf("update stress_info fail, err: %v", errs)
	}

	ts.Commit()
	return nil
}

func GetStressInfoById(id uint) (*schema.StressInfo, error) {
	stressInfo := &schema.StressInfo{}
	dbPtr := Mysql_db.Where(schema.StressInfo{ID: id}).Find(stressInfo)
	if errs := dbPtr.GetErrors(); len(errs) > 0 {
		return stressInfo, fmt.Errorf("gorm db err: err=%v", errs)
	}
	return stressInfo, nil
}

// 判断是否存在开启的压测任务
func ExistsEnableStressByHid(hid uint) (bool, error) {
	notFound := Mysql_db.Where(schema.StressInfo{Hid: hid, IsEnable: 1}).Find(&schema.StressInfo{}).RecordNotFound()
	errs := Mysql_db.GetErrors()
	if len(errs) > 0 {
		err := fmt.Errorf("gorm db err: err=%v", errs)
		return false, err
	}
	return !notFound, nil
}

func UpdateStressQps(stressId uint, qps string) error {
	if errs := Mysql_db.Model(&schema.StressInfo{}).Where("id = ?", stressId).Updates(
		map[string]interface{}{"qps": qps}).GetErrors(); len(errs) != 0 {
		return fmt.Errorf("update stress_info fail, err: %v", errs)
	}
	return nil
}
