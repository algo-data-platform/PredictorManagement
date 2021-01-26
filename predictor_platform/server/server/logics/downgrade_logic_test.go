package logics

import (
	"reflect"
	"server/conf"
	"server/env"
	"server/mock"
	"server/schema"
	"server/server/dao"
	"testing"
)

func TestGetIpModelPercentMap(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./downgrade_logic_test.db"
	db := env.InitMysql(conf)
	dao.SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3

	db.Create(&schema.Model{Name: "model_a", Desc: "模型A", Path: "/data0/dummy/path"}) // Mid: 1
	db.Create(&schema.Model{Name: "model_b", Desc: "模型B", Path: "/data0/dummy/path"}) // Mid: 2
	db.Create(&schema.Model{Name: "model_c", Desc: "模型C", Path: "/data0/dummy/path"}) // Mid: 3
	db.Create(&schema.Model{Name: "model_d", Desc: "模型D", Path: "/data0/dummy/path"}) // Mid: 4

	db.Create(&schema.ServiceModel{Sid: 1, Mid: 1, Desc: "service_1 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 2, Desc: "service_1 -> model_b"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 3, Desc: "service_1 -> model_c"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 1, Desc: "service_2 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 4, Desc: "service_2 -> model_d"})

	db.Create(&schema.Host{ID: 2, Ip: "127.1.0.2"})
	db.Create(&schema.Host{ID: 3, Ip: "127.2.0.3"})
	db.Create(&schema.Host{ID: 4, Ip: "127.3.0.4"})
	db.Create(&schema.Host{ID: 5, Ip: "127.1.0.5"})
	db.Create(&schema.Host{ID: 7, Ip: "127.3.0.7"})

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.1.0.2 -> service_1", LoadWeight: 120})
	db.Create(&schema.HostService{Hid: 5, Sid: 1, Desc: "127.1.0.5 -> service_1", LoadWeight: 130})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3", LoadWeight: 200})
	db.Create(&schema.HostService{Hid: 4, Sid: 2, Desc: "127.3.0.4 -> service_2", LoadWeight: 300})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.3.0.4 -> service_3", LoadWeight: 90})
	db.Create(&schema.HostService{Hid: 7, Sid: 1, Desc: "127.3.0.7 -> service_1", LoadWeight: 80})

	// case 1
	{

		percent := 3
		ipList := []string{
			"127.1.0.2",
			"127.1.0.5",
			"127.2.0.9",
		}
		expectIpModelPercentMap := map[string]modelPercentMap{
			"127.1.0.2": modelPercentMap{
				"model_a": 3,
				"model_b": 3,
				"model_c": 3,
			},
			"127.1.0.5": modelPercentMap{
				"model_a": 3,
				"model_b": 3,
				"model_c": 3,
			},
		}
		ipModelPercentMap, err := getIpModelPercentMap(ipList, percent)

		if err != nil || !reflect.DeepEqual(expectIpModelPercentMap, ipModelPercentMap) {
			t.Errorf("TestGetIpModelPercentMap() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				ipModelPercentMap, expectIpModelPercentMap)
		}
	}
	// case 2
	{

		percent := 3
		ipList := []string{
			"127.1.0.2",
			"127.1.0.5",
			"127.2.0.3",
		}
		expectIpModelPercentMap := map[string]modelPercentMap{
			"127.1.0.2": modelPercentMap{
				"model_a": 3,
				"model_b": 3,
				"model_c": 3,
			},
			"127.1.0.5": modelPercentMap{
				"model_a": 3,
				"model_b": 3,
				"model_c": 3,
			},
			"127.2.0.3": modelPercentMap{
				"model_a": 3,
				"model_d": 3,
			},
		}
		ipModelPercentMap, err := getIpModelPercentMap(ipList, percent)

		if err != nil || !reflect.DeepEqual(expectIpModelPercentMap, ipModelPercentMap) {
			t.Errorf("TestGetIpModelPercentMap() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				ipModelPercentMap, expectIpModelPercentMap)
		}
	}

	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}
