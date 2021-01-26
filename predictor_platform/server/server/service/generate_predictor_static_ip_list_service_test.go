package service

import (
	"fmt"
	"reflect"
	"server/conf"
	"server/env"
	"server/mock"
	"server/schema"
	"server/server/dao"
	"testing"
)

func TestGetServiceIpList(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./generate_predictor_static_ip_list_service_test.db"
	db := env.InitMysql(conf)
	dao.SetMysqlDB(db)
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
		service_map := map[string]uint{
			"service_1": 1,
		}
		expectServiceIpListMap := map[string][]string{
			"service_1": []string{
				fmt.Sprintf("thrift,%v:9537,default,%v", "127.1.0.2", 120),
				fmt.Sprintf("thrift,%v:9537,default,%v", "127.1.0.5", 130),
				fmt.Sprintf("thrift,%v:9537,default,%v", "127.3.0.7", 100),
			},
		}
		serviceIpListMap := GetServiceIpList(service_map)
		if !reflect.DeepEqual(expectServiceIpListMap, serviceIpListMap) {
			t.Errorf("TestGetServiceIpList() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				serviceIpListMap, expectServiceIpListMap)
		}
	}
	// case 2
	{
		service_map := map[string]uint{}
		expectServiceIpListMap := map[string][]string{}
		serviceIpListMap := GetServiceIpList(service_map)
		if !reflect.DeepEqual(expectServiceIpListMap, serviceIpListMap) {
			t.Errorf("TestGetServiceIpList() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				serviceIpListMap, expectServiceIpListMap)
		}
	}
	// case 3
	{
		service_map := map[string]uint{
			"service_1": 1,
			"service_2": 2,
		}
		expectServiceIpListMap := map[string][]string{
			"service_1": []string{
				fmt.Sprintf("thrift,%v:9537,default,%v", "127.1.0.2", 120),
				fmt.Sprintf("thrift,%v:9537,default,%v", "127.1.0.5", 130),
				fmt.Sprintf("thrift,%v:9537,default,%v", "127.3.0.7", 100),
			},
			"service_2": []string{
				fmt.Sprintf("thrift,%v:9537,default,%v", "127.2.0.3", 100),
				fmt.Sprintf("thrift,%v:9537,default,%v", "127.3.0.4", 300),
			},
		}
		serviceIpListMap := GetServiceIpList(service_map)
		if !reflect.DeepEqual(expectServiceIpListMap, serviceIpListMap) {
			t.Errorf("TestGetServiceIpList() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				serviceIpListMap, expectServiceIpListMap)
		}
	}
}
