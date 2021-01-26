package dao

import (
	"reflect"
	"server/conf"
	"server/env"
	"server/mock"
	"server/schema"
	"server/util"
	"testing"
)

func TestGetOriginHostNumMap(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./host_services_test.db"
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

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.1.0.2 -> service_1", LoadWeight: 120})
	db.Create(&schema.HostService{Hid: 5, Sid: 1, Desc: "127.1.0.5 -> service_1", LoadWeight: 130})
	db.Create(&schema.HostService{Hid: 6, Sid: 1, Desc: "127.1.0.6 -> service_1", LoadWeight: 140})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3", LoadWeight: 200})
	db.Create(&schema.HostService{Hid: 4, Sid: 2, Desc: "127.3.0.4 -> service_2", LoadWeight: 300})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.3.0.4 -> service_3", LoadWeight: 90})
	db.Create(&schema.HostService{Hid: 7, Sid: 1, Desc: "127.3.0.7 -> service_1", LoadWeight: 80})

	// case 1
	{
		sids := []uint{1, 2, 3}
		expectOriginHostNumMap := map[uint]int{
			1: 3,
			2: 2,
			3: 2,
		}
		originHostNumMap, err := GetOriginHostNumMap(sids)
		if err != nil || !reflect.DeepEqual(expectOriginHostNumMap, originHostNumMap) {
			t.Errorf("TestGetOriginHostNumMap() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				originHostNumMap, expectOriginHostNumMap)
		}
	}
	// case 2
	{
		sids := []uint{4}
		expectOriginHostNumMap := map[uint]int{}
		originHostNumMap, err := GetOriginHostNumMap(sids)
		if err != nil || !reflect.DeepEqual(expectOriginHostNumMap, originHostNumMap) {
			t.Errorf("TestGetOriginHostNumMap() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				originHostNumMap, expectOriginHostNumMap)
		}
	}
	// case 3
	{
		sids := []uint{1}
		expectOriginHostNumMap := map[uint]int{
			1: 3,
		}
		originHostNumMap, err := GetOriginHostNumMap(sids)
		if err != nil || !reflect.DeepEqual(expectOriginHostNumMap, originHostNumMap) {
			t.Errorf("TestGetOriginHostNumMap() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				originHostNumMap, expectOriginHostNumMap)
		}
	}

	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}

func TestGetAllocatedHostService(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./host_services_test.db"
	db := env.InitMysql(conf)
	SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)
	db.Create(&schema.Host{ID: 2, Ip: "127.1.0.2"})
	db.Create(&schema.Host{ID: 3, Ip: "127.2.0.3"})
	db.Create(&schema.Host{ID: 4, Ip: "127.3.0.4"})
	db.Create(&schema.Host{ID: 5, Ip: "127.1.0.5"})

	db.Create(&schema.Host{ID: 7, Ip: "127.3.0.7"})

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.1.0.2 -> service_1", LoadWeight: 120})
	db.Create(&schema.HostService{Hid: 5, Sid: 1, Desc: "127.1.0.5 -> service_1", LoadWeight: 130})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3", LoadWeight: 200})
	db.Create(&schema.HostService{Hid: 4, Sid: 2, Desc: "127.3.0.4 -> service_2", LoadWeight: 300})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.3.0.4 -> service_3", LoadWeight: 90})
	db.Create(&schema.HostService{Hid: 7, Sid: 1, Desc: "127.3.0.7 -> service_1", LoadWeight: 80})

	// case 1
	{
		expectAllocatedHostService := []util.SidHostNum{}
		allocatedHostService, err := GetAllocatedHostService()
		if err != nil || !reflect.DeepEqual(expectAllocatedHostService, allocatedHostService) {
			t.Errorf("TestGetAllocatedHostService() failed, err: %v, \n[Got]:\n%v,\n[Expect]:\n%v",
				err, allocatedHostService, expectAllocatedHostService)
		}
	}

	// case 2
	db.Create(&schema.Host{ID: 6, Ip: "127.1.0.6", Desc: "elastic_expansion"})
	db.Create(&schema.HostService{Hid: 6, Sid: 1, Desc: "127.1.0.6 -> service_1", LoadWeight: 140})
	{
		expectAllocatedHostService := []util.SidHostNum{
			util.SidHostNum{
				Sid:     1,
				HostNum: 1,
			},
		}
		allocatedHostService, err := GetAllocatedHostService()
		if err != nil || !reflect.DeepEqual(expectAllocatedHostService, allocatedHostService) {
			t.Errorf("TestGetAllocatedHostService() failed, err: %v, \n[Got]:\n%v,\n[Expect]:\n%v",
				err, allocatedHostService, expectAllocatedHostService)
		}
	}
	// case 3
	db.Create(&schema.Host{ID: 8, Ip: "127.3.0.8", Desc: "elastic_expansion"})
	db.Create(&schema.Host{ID: 9, Ip: "127.3.0.9", Desc: "elastic_expansion"})
	db.Create(&schema.Host{ID: 10, Ip: "127.3.0.10", Desc: "elastic_expansion"})
	db.Create(&schema.Host{ID: 11, Ip: "127.3.0.11", Desc: "elastic_expansion"})
	db.Create(&schema.HostService{Hid: 8, Sid: 1, Desc: "127.3.0.8 -> service_1", LoadWeight: 80})
	db.Create(&schema.HostService{Hid: 9, Sid: 2, Desc: "127.3.0.9 -> service_2", LoadWeight: 80})
	db.Create(&schema.HostService{Hid: 10, Sid: 3, Desc: "127.3.0.8 -> service_3", LoadWeight: 80})
	db.Create(&schema.HostService{Hid: 11, Sid: 2, Desc: "127.3.0.8 -> service_2", LoadWeight: 80})
	{
		expectAllocatedHostService := []util.SidHostNum{
			util.SidHostNum{
				Sid:     1,
				HostNum: 2,
			},
			util.SidHostNum{
				Sid:     2,
				HostNum: 2,
			},
			util.SidHostNum{
				Sid:     3,
				HostNum: 1,
			},
		}
		allocatedHostService, err := GetAllocatedHostService()
		mock.SortSidHostNumsBySid(allocatedHostService)
		mock.SortSidHostNumsBySid(expectAllocatedHostService)
		if err != nil || !reflect.DeepEqual(expectAllocatedHostService, allocatedHostService) {
			t.Errorf("TestGetAllocatedHostService() failed, err: %v, \n[Got]:\n%v,\n[Expect]:\n%v",
				err, allocatedHostService, expectAllocatedHostService)
		}
	}

	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}

func TestGetHostServiceMap(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./host_services_test.db"
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

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.1.0.2 -> service_1", LoadWeight: 120})
	db.Create(&schema.HostService{Hid: 5, Sid: 1, Desc: "127.1.0.5 -> service_1", LoadWeight: 130})
	db.Create(&schema.HostService{Hid: 6, Sid: 1, Desc: "127.1.0.6 -> service_1", LoadWeight: 140})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3", LoadWeight: 200})
	db.Create(&schema.HostService{Hid: 4, Sid: 2, Desc: "127.3.0.4 -> service_2", LoadWeight: 300})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.3.0.4 -> service_3", LoadWeight: 90})
	db.Create(&schema.HostService{Hid: 7, Sid: 1, Desc: "127.3.0.7 -> service_1", LoadWeight: 80})

	// case 1
	{
		expectHostServiceMap := map[string][]string{
			"127.1.0.2": []string{"service_1"},
			"127.2.0.3": []string{"service_2", "service_3"},
			"127.3.0.4": []string{"service_2", "service_3"},
			"127.1.0.5": []string{"service_1"},
			"127.1.0.6": []string{"service_1"},
			"127.3.0.7": []string{"service_1"},
		}
		hostServiceMap, err := GetHostServiceMap()
		if err != nil || !reflect.DeepEqual(expectHostServiceMap, hostServiceMap) {
			t.Errorf("TestGetHostServiceMap() failed, err:%v, \n[Got]:\n%v,\n[Expect]:\n%v",
				err, hostServiceMap, expectHostServiceMap)
		}
	}

	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}
