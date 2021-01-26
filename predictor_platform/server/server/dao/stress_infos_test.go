package dao

import (
	"server/conf"
	"server/env"
	"server/mock"
	"server/schema"
	"strings"
	"testing"
)

func TestTransactEnableStressTest(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./stress_infos_test.db"
	db := env.InitMysql(conf)
	SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)

	db.Create(&schema.Host{ID: 2, Ip: "127.1.0.2"})
	db.Create(&schema.Host{ID: 3, Ip: "127.2.0.3"})
	db.Create(&schema.Host{ID: 4, Ip: "127.3.0.4"})
	db.Create(&schema.Host{ID: 5, Ip: "127.1.0.5"})
	db.Create(&schema.Host{ID: 6, Ip: "127.1.0.6", Desc: "elastic_expansion"})
	db.Create(&schema.Host{ID: 7, Ip: "127.3.0.7"})

	db.Create(&schema.Model{Name: "model_a", Path: "/data0/dummy/path"}) // Mid: 1
	db.Create(&schema.Model{Name: "model_b", Path: "/data0/dummy/path"}) // Mid: 2
	db.Create(&schema.Model{Name: "model_c", Path: "/data0/dummy/path"}) // Mid: 3
	db.Create(&schema.Model{Name: "model_d", Path: "/data0/dummy/path"}) // Mid: 4

	db.Create(&schema.Service{Name: "service_1"})    // Sid: 1
	db.Create(&schema.Service{Name: "service_2"})    // Sid: 2
	db.Create(&schema.Service{Name: "service_3"})    // Sid: 3
	db.Create(&schema.Service{Name: "service_test"}) // Sid: 4

	db.Create(&schema.ServiceModel{Sid: 1, Mid: 1, Desc: "service_1 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 2, Desc: "service_1 -> model_b"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 3, Desc: "service_1 -> model_c"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 1, Desc: "service_2 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 4, Desc: "service_2 -> model_d"})

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.1.0.2 -> service_1", LoadWeight: 120})
	db.Create(&schema.HostService{Hid: 5, Sid: 1, Desc: "127.1.0.5 -> service_1", LoadWeight: 130})
	db.Create(&schema.HostService{Hid: 6, Sid: 1, Desc: "127.1.0.6 -> service_1", LoadWeight: 140})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3", LoadWeight: 200})
	db.Create(&schema.HostService{Hid: 4, Sid: 2, Desc: "127.3.0.4 -> service_2", LoadWeight: 300})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.3.0.4 -> service_3", LoadWeight: 90})
	db.Create(&schema.HostService{Hid: 7, Sid: 1, Desc: "127.3.0.7 -> service_1", LoadWeight: 80})

	db.Create(&schema.StressInfo{Hid: 3, Mids: "1,2", Qps: "100,200", IsEnable: 0})

	// case 1
	{
		stressInfo := &schema.StressInfo{
			Hid:      2,
			Mids:     "1,2",
			Qps:      "100,200",
			IsEnable: 1,
		}
		stressTestService := &schema.Service{
			ID:   4,
			Name: "service_test",
		}
		err := TransactEnableStressTest(stressInfo, stressTestService)
		if err != nil {
			t.Errorf("TestTransactEnableStressTest() failed, err: %v", err)
		}
		// todo 查看获取test_service-> model数据
		var serviceModels []*schema.ServiceModel
		if errs := Mysql_db.Where("sid = ?", 4).Find(&serviceModels).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactEnableStressTest() failed, errs: %v", errs)
		}
		expectServiceModels := []*schema.ServiceModel{
			&schema.ServiceModel{
				ID:  6,
				Sid: 4,
				Mid: 1,
			},
			&schema.ServiceModel{
				ID:  7,
				Sid: 4,
				Mid: 2,
			},
		}
		if !mock.IsEqualServiceModelSlice(expectServiceModels, serviceModels) {
			t.Errorf("TestTransactEnableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				serviceModels, expectServiceModels)
		}

		// todo 查看获取test_service-> host数据
		var hostServices []*schema.HostService
		if errs := Mysql_db.Where("hid = ?", 2).Find(&hostServices).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactEnableStressTest() failed, errs: %v", errs)
		}
		expectHostServices := []*schema.HostService{
			&schema.HostService{
				ID:  9,
				Sid: 4,
				Hid: 2,
			},
		}
		if !mock.IsEqualHostServiceSlice(expectHostServices, hostServices) {
			t.Errorf("TestTransactEnableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				hostServices, expectHostServices)
		}
	}

	// case 2 测试事务，出错情况
	{
		stressInfo := &schema.StressInfo{
			Hid:      9,
			Mids:     "1,2",
			Qps:      "100,200",
			IsEnable: 1,
		}
		stressTestService := &schema.Service{
			ID:   4,
			Name: "service_test",
		}
		err := TransactEnableStressTest(stressInfo, stressTestService)
		if err.Error() != "create host_services fail, err: [FOREIGN KEY constraint failed]" {
			t.Errorf("TestTransactEnableStressTest() failed, err: %v", err)
		}
		// todo 查看获取test_service-> model数据
		var serviceModels []*schema.ServiceModel
		if errs := Mysql_db.Where("sid = ?", 4).Find(&serviceModels).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactEnableStressTest() failed, errs: %v", errs)
		}
		expectServiceModels := []*schema.ServiceModel{
			&schema.ServiceModel{
				ID:  6,
				Sid: 4,
				Mid: 1,
			},
			&schema.ServiceModel{
				ID:  7,
				Sid: 4,
				Mid: 2,
			},
		}
		if !mock.IsEqualServiceModelSlice(expectServiceModels, serviceModels) {
			t.Errorf("TestTransactEnableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				serviceModels, expectServiceModels)
		}

		// todo 查看获取test_service-> host数据
		var hostServices []*schema.HostService
		if errs := Mysql_db.Where("hid = ?", 9).Find(&hostServices).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactEnableStressTest() failed, errs: %v", errs)
		}
		expectHostServices := []*schema.HostService{
			&schema.HostService{},
		}
		if mock.IsEqualHostServiceSlice(expectHostServices, hostServices) {
			t.Errorf("TestTransactEnableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				hostServices, expectHostServices)
		}
	}
	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}

func TestTransactDisableStressTest(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./stress_infos_test.db"
	db := env.InitMysql(conf)
	SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)

	db.Create(&schema.Host{ID: 2, Ip: "127.1.0.2"})
	db.Create(&schema.Host{ID: 3, Ip: "127.2.0.3"})
	db.Create(&schema.Host{ID: 4, Ip: "127.3.0.4"})
	db.Create(&schema.Host{ID: 5, Ip: "127.1.0.5"})
	db.Create(&schema.Host{ID: 6, Ip: "127.1.0.6", Desc: "elastic_expansion"})
	db.Create(&schema.Host{ID: 7, Ip: "127.3.0.7"})

	db.Create(&schema.Model{Name: "model_a", Path: "/data0/dummy/path"}) // Mid: 1
	db.Create(&schema.Model{Name: "model_b", Path: "/data0/dummy/path"}) // Mid: 2
	db.Create(&schema.Model{Name: "model_c", Path: "/data0/dummy/path"}) // Mid: 3
	db.Create(&schema.Model{Name: "model_d", Path: "/data0/dummy/path"}) // Mid: 4

	db.Create(&schema.Service{Name: "service_1"})    // Sid: 1
	db.Create(&schema.Service{Name: "service_2"})    // Sid: 2
	db.Create(&schema.Service{Name: "service_3"})    // Sid: 3
	db.Create(&schema.Service{Name: "service_test"}) // Sid: 4

	db.Create(&schema.ServiceModel{Sid: 1, Mid: 1, Desc: "service_1 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 2, Desc: "service_1 -> model_b"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 3, Desc: "service_1 -> model_c"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 1, Desc: "service_2 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 4, Desc: "service_2 -> model_d"})
	db.Create(&schema.ServiceModel{Sid: 4, Mid: 1, Desc: "service_test -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 4, Mid: 2, Desc: "service_test -> model_b"})

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.1.0.2 -> service_1", LoadWeight: 120})
	db.Create(&schema.HostService{Hid: 5, Sid: 1, Desc: "127.1.0.5 -> service_1", LoadWeight: 130})
	db.Create(&schema.HostService{Hid: 6, Sid: 1, Desc: "127.1.0.6 -> service_1", LoadWeight: 140})
	// db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2", LoadWeight: 100})
	// db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3", LoadWeight: 200})
	db.Create(&schema.HostService{Hid: 4, Sid: 2, Desc: "127.3.0.4 -> service_2", LoadWeight: 300})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.3.0.4 -> service_3", LoadWeight: 90})
	db.Create(&schema.HostService{Hid: 7, Sid: 1, Desc: "127.3.0.7 -> service_1", LoadWeight: 80})
	db.Create(&schema.HostService{Hid: 3, Sid: 4, Desc: "127.2.0.3 -> service_test", LoadWeight: 0})

	db.Create(&schema.StressInfo{Hid: 3, Mids: "1,2", Qps: "100,200", OriginSids: "2_100,3_200", IsEnable: 1})

	// case 1
	{
		stressInfo := &schema.StressInfo{
			ID:         1,
			Hid:        3,
			Mids:       "1,2",
			Qps:        "100,200",
			OriginSids: "2_100,3_200",
			IsEnable:   1,
		}
		stressTestService := &schema.Service{
			ID:   4,
			Name: "service_test",
		}
		err := TransactDisableStressTest(stressInfo, stressTestService)
		if err != nil {
			t.Errorf("TestTransactDisableStressTest() failed, err: %v", err)
		}
		// todo 查看获取test_service-> model数据
		var serviceModels []*schema.ServiceModel
		if errs := Mysql_db.Where("sid = ?", 4).Find(&serviceModels).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactDisableStressTest() failed, errs: %v", errs)
		}
		expectServiceModels := []*schema.ServiceModel{}
		if !mock.IsEqualServiceModelSlice(expectServiceModels, serviceModels) {
			t.Errorf("TestTransactDisableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				serviceModels, expectServiceModels)
		}

		// todo 查看获取test_service-> host数据
		var hostServices []*schema.HostService
		if errs := Mysql_db.Where("hid = ?", 3).Find(&hostServices).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactDisableStressTest() failed, errs: %v", errs)
		}
		expectHostServices := []*schema.HostService{
			&schema.HostService{
				ID:         8,
				Sid:        2,
				Hid:        3,
				LoadWeight: 100,
			},
			&schema.HostService{
				ID:         9,
				Sid:        3,
				Hid:        3,
				LoadWeight: 200,
			},
		}
		if !mock.IsEqualHostServiceSlice(expectHostServices, hostServices) {
			t.Errorf("TestTransactDisableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				hostServices, expectHostServices)
		}
		// 读取压测状态
		var stressInfos []*schema.StressInfo
		if errs := Mysql_db.Where("id = ?", 1).Find(&stressInfos).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactDisableStressTest() failed, errs: %v", errs)
		}
		expectedStressInfo := []*schema.StressInfo{
			&schema.StressInfo{
				ID:         uint(1),
				Hid:        uint(3),
				Mids:       "1,2",
				Qps:        "100,200",
				OriginSids: "2_100,3_200",
				IsEnable:   uint(0),
			},
		}
		if !mock.IsEqualStressInfoSlice(expectedStressInfo, stressInfos) {
			t.Errorf("TestTransactDisableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				stressInfos, expectedStressInfo)
		}
	}
	// case 2 测试删除新增host_service ID 变化
	{
		stressInfo := &schema.StressInfo{
			ID:         1,
			Hid:        3,
			Mids:       "1,2",
			Qps:        "100,200",
			OriginSids: "2_100,3_200",
			IsEnable:   1,
		}
		stressTestService := &schema.Service{
			ID:   4,
			Name: "service_test",
		}
		err := TransactDisableStressTest(stressInfo, stressTestService)
		if err != nil {
			t.Errorf("TestTransactDisableStressTest() failed, err: %v", err)
		}
		// todo 查看获取test_service-> model数据
		var serviceModels []*schema.ServiceModel
		if errs := Mysql_db.Where("sid = ?", 4).Find(&serviceModels).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactDisableStressTest() failed, errs: %v", errs)
		}
		expectServiceModels := []*schema.ServiceModel{}
		if !mock.IsEqualServiceModelSlice(expectServiceModels, serviceModels) {
			t.Errorf("TestTransactDisableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				serviceModels, expectServiceModels)
		}

		// todo 查看获取test_service-> host数据
		var hostServices []*schema.HostService
		if errs := Mysql_db.Where("hid = ?", 3).Find(&hostServices).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactDisableStressTest() failed, errs: %v", errs)
		}
		expectHostServices := []*schema.HostService{
			&schema.HostService{
				ID:         10,
				Sid:        2,
				Hid:        3,
				LoadWeight: 100,
			},
			&schema.HostService{
				ID:         11,
				Sid:        3,
				Hid:        3,
				LoadWeight: 200,
			},
		}
		if !mock.IsEqualHostServiceSlice(expectHostServices, hostServices) {
			t.Errorf("TestTransactDisableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				hostServices, expectHostServices)
		}
	}
	// case 3 测试事务
	{
		db.Create(&schema.ServiceModel{Sid: 4, Mid: 1, Desc: "service_test -> model_a"})
		db.Create(&schema.ServiceModel{Sid: 4, Mid: 2, Desc: "service_test -> model_b"})
		db.Create(&schema.HostService{Hid: 3, Sid: 4, Desc: "127.2.0.3 -> service_test", LoadWeight: 0})
		stressInfo := &schema.StressInfo{
			ID:         1,
			Hid:        3,
			Mids:       "1,2",
			Qps:        "100,200",
			OriginSids: "a_100,b_200",
			IsEnable:   1,
		}
		stressTestService := &schema.Service{
			ID:   4,
			Name: "service_test",
		}
		err := TransactDisableStressTest(stressInfo, stressTestService)
		if !strings.Contains(err.Error(), "strconv.ParseUint fail, err") {
			t.Errorf("TestTransactDisableStressTest() failed, err: %v", err)
		}
		// todo 查看获取test_service-> model数据
		var serviceModels []*schema.ServiceModel
		if errs := Mysql_db.Where("sid = ?", 4).Find(&serviceModels).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactDisableStressTest() failed, errs: %v", errs)
		}
		expectServiceModels := []*schema.ServiceModel{
			&schema.ServiceModel{
				ID:  8,
				Sid: 4,
				Mid: 1,
			},
			&schema.ServiceModel{
				ID:  9,
				Sid: 4,
				Mid: 2,
			},
		}
		if !mock.IsEqualServiceModelSlice(expectServiceModels, serviceModels) {
			t.Errorf("TestTransactDisableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				serviceModels, expectServiceModels)
		}

		// todo 查看获取test_service-> host数据
		var hostServices []*schema.HostService
		if errs := Mysql_db.Where("hid = ?", 3).Find(&hostServices).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactDisableStressTest() failed, errs: %v", errs)
		}
		expectHostServices := []*schema.HostService{
			&schema.HostService{
				ID:         10,
				Sid:        2,
				Hid:        3,
				LoadWeight: 100,
			},
			&schema.HostService{
				ID:         11,
				Sid:        3,
				Hid:        3,
				LoadWeight: 200,
			},
			&schema.HostService{
				ID:  12,
				Sid: 4,
				Hid: 3,
			},
		}
		if !mock.IsEqualHostServiceSlice(expectHostServices, hostServices) {
			t.Errorf("TestTransactDisableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				hostServices, expectHostServices)
		}
		// 读取压测状态
		var stressInfos []*schema.StressInfo
		if errs := Mysql_db.Where("id = ?", 1).Find(&stressInfos).GetErrors(); len(errs) != 0 {
			t.Errorf("TestTransactDisableStressTest() failed, errs: %v", errs)
		}
		expectedStressInfo := []*schema.StressInfo{
			&schema.StressInfo{
				ID:         uint(1),
				Hid:        uint(3),
				Mids:       "1,2",
				Qps:        "100,200",
				OriginSids: "2_100,3_200",
				IsEnable:   uint(1),
			},
		}
		if !mock.IsEqualStressInfoSlice(expectedStressInfo, stressInfos) {
			t.Errorf("TestTransactDisableStressTest() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				stressInfos, expectedStressInfo)
		}
	}

	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}
