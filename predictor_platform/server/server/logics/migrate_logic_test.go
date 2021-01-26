package logics

import (
	"reflect"
	"server/conf"
	"server/env"
	"server/mock"
	"server/schema"
	"server/server/dao"
	"server/util"
	"testing"
)

func TestGetGroupServices(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./migrate_logic_test.db"
	db := env.InitMysql(conf)
	dao.SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3
	db.Create(&schema.Service{Name: "service_4"}) // Sid: 4
	db.Create(&schema.Service{Name: "service_5"}) // Sid: 5

	db.Create(&schema.Host{ID: 2, Ip: "127.1.0.2"})
	db.Create(&schema.Host{ID: 3, Ip: "127.2.0.3"})
	db.Create(&schema.Host{ID: 4, Ip: "127.3.0.4"})
	db.Create(&schema.Host{ID: 5, Ip: "127.1.0.5"})
	db.Create(&schema.Host{ID: 7, Ip: "127.3.0.7"})
	db.Create(&schema.Host{ID: 8, Ip: "127.3.0.8"})
	db.Create(&schema.Host{ID: 9, Ip: "127.3.0.9"})
	db.Create(&schema.Host{ID: 10, Ip: "127.3.1.0"})

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.1.0.2 -> service_1", LoadWeight: 120})
	db.Create(&schema.HostService{Hid: 5, Sid: 1, Desc: "127.1.0.5 -> service_1", LoadWeight: 130})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2", LoadWeight: 100})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3", LoadWeight: 200})
	db.Create(&schema.HostService{Hid: 4, Sid: 2, Desc: "127.3.0.4 -> service_2", LoadWeight: 300})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.3.0.4 -> service_3", LoadWeight: 90})
	db.Create(&schema.HostService{Hid: 7, Sid: 1, Desc: "127.3.0.7 -> service_1", LoadWeight: 80})

	// case 1
	{
		expectGroupServices := []util.GroupSid{
			util.GroupSid{
				2, 3,
			},
		}
		groupServices, err := getGroupServices()
		if err != nil || !reflect.DeepEqual(expectGroupServices, groupServices) {
			t.Errorf("TestGetIpModelPercentMap() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				groupServices, expectGroupServices)
		}
	}
	// case 2
	{
		db.Create(&schema.HostService{Hid: 8, Sid: 5, Desc: "127.3.0.8 -> service_5", LoadWeight: 80})
		db.Create(&schema.HostService{Hid: 8, Sid: 3, Desc: "127.3.0.8 -> service_3", LoadWeight: 80})
		db.Create(&schema.HostService{Hid: 8, Sid: 1, Desc: "127.3.0.8 -> service_1", LoadWeight: 80})
		db.Create(&schema.HostService{Hid: 9, Sid: 3, Desc: "127.3.0.9 -> service_3", LoadWeight: 80})
		db.Create(&schema.HostService{Hid: 9, Sid: 1, Desc: "127.3.0.9 -> service_1", LoadWeight: 80})
		db.Create(&schema.HostService{Hid: 10, Sid: 1, Desc: "127.3.1.0 -> service_1", LoadWeight: 80})
		db.Create(&schema.HostService{Hid: 10, Sid: 3, Desc: "127.3.1.0 -> service_3", LoadWeight: 80})

		expectGroupServices := []util.GroupSid{
			util.GroupSid{
				2, 3,
			},
			util.GroupSid{
				5, 3, 1,
			},
			util.GroupSid{
				3, 1,
			},
		}
		groupServices, err := getGroupServices()
		expectGroupMap := mock.GroupSidsToMap(expectGroupServices)
		GroupMap := mock.GroupSidsToMap(groupServices)
		if err != nil || !reflect.DeepEqual(expectGroupMap, GroupMap) {
			t.Errorf("TestGetGroupServices() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				GroupMap, expectGroupMap)
		}
	}
	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}

func TestGetToMigratedIpsByIDC(t *testing.T) {
	env.Env = &env.Environment{
		Conf: &conf.Conf{
			StressTestService:  "predicter_service_QA",
			ConsistenceService: "algo_service_consistence",
			MigrateHosts: conf.MigrateHosts{
				ExcludeHosts: []string{
					"127.2.12.125",
					"127.4.192.222",
				},
			},
		},
	}
	// case 1
	{
		fromIDCIpsMap := map[string][]string{
			"huawei": []string{
				"127.1.0.1",
				"127.1.0.2",
				"127.1.0.3",
			},
			"aliyun": []string{
				"127.4.0.1",
				"127.4.0.2",
				"127.4.0.3",
			},
		}
		toIDCIpsMap := map[string][]string{
			"huawei": []string{
				"127.1.0.4",
				"127.1.0.5",
			},
			"aliyun": []string{
				"127.4.0.4",
			},
		}
		num := 3
		expectToMigratedIpsMap := map[string][]string{
			"huawei": []string{
				"127.1.0.1",
				"127.1.0.2",
			},
			"aliyun": []string{
				"127.4.0.1",
			},
		}
		toMigratedIpsMap := getToMigratedIpsByIDC(fromIDCIpsMap, toIDCIpsMap, num)

		if !reflect.DeepEqual(expectToMigratedIpsMap, toMigratedIpsMap) {
			t.Errorf("TestGetToMigratedIpsByIDC() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				toMigratedIpsMap, expectToMigratedIpsMap)
		}
	}
	// case 2
	{
		fromIDCIpsMap := map[string][]string{
			"huawei": []string{
				"127.1.0.1",
				"127.1.0.2",
				"127.1.0.3",
			},
		}
		toIDCIpsMap := map[string][]string{
			"huawei": []string{
				"127.1.0.4",
				"127.1.0.5",
			},
			"bx": []string{
				"127.3.0.4",
			},
		}
		num := 3
		expectToMigratedIpsMap := map[string][]string{
			"huawei": []string{
				"127.1.0.1",
				"127.1.0.2",
				"127.1.0.3",
			},
		}
		toMigratedIpsMap := getToMigratedIpsByIDC(fromIDCIpsMap, toIDCIpsMap, num)

		if !reflect.DeepEqual(expectToMigratedIpsMap, toMigratedIpsMap) {
			t.Errorf("TestGetToMigratedIpsByIDC() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				toMigratedIpsMap, expectToMigratedIpsMap)
		}
	}
	// case 3
	{
		fromIDCIpsMap := map[string][]string{
			"huawei": []string{
				"127.1.0.1",
				"127.1.0.2",
				"127.1.0.3",
			},
			"aliyun": []string{
				"127.4.0.1",
				"127.4.0.2",
				"127.4.0.3",
			},
		}
		toIDCIpsMap := map[string][]string{
			"huawei": []string{
				"127.1.0.4",
			},
			"bx": []string{
				"127.3.0.4",
			},
		}
		num := 4
		expectToMigratedIpsMap1 := map[string][]string{
			"huawei": []string{
				"127.1.0.1",
				"127.1.0.2",
				"127.1.0.3",
			},
			"aliyun": []string{
				"127.4.0.1",
			},
		}
		expectToMigratedIpsMap2 := map[string][]string{
			"huawei": []string{
				"127.1.0.1",
				"127.1.0.2",
			},
			"aliyun": []string{
				"127.4.0.1",
				"127.4.0.2",
			},
		}
		toMigratedIpsMap := getToMigratedIpsByIDC(fromIDCIpsMap, toIDCIpsMap, num)

		if !reflect.DeepEqual(expectToMigratedIpsMap1, toMigratedIpsMap) &&
			!reflect.DeepEqual(expectToMigratedIpsMap2, toMigratedIpsMap) {
			t.Errorf("TestGetToMigratedIpsByIDC() failed, \n[Got]:\n%v,\n[Expect1]:\n%v\n[Expect2]:\n%v",
				toMigratedIpsMap, expectToMigratedIpsMap1, expectToMigratedIpsMap2)
		}
	}
	// case 4
	{
		fromIDCIpsMap := map[string][]string{
			"aliyun": []string{
				"127.4.0.1",
				"127.4.0.2",
				"127.4.0.3",
			},
		}
		toIDCIpsMap := map[string][]string{
			"huawei": []string{
				"127.1.0.4",
			},
			"bx": []string{
				"127.3.0.4",
			},
		}
		num := 5
		expectToMigratedIpsMap := map[string][]string{
			"aliyun": []string{
				"127.4.0.1",
				"127.4.0.2",
				"127.4.0.3",
			},
		}
		toMigratedIpsMap := getToMigratedIpsByIDC(fromIDCIpsMap, toIDCIpsMap, num)

		if !reflect.DeepEqual(expectToMigratedIpsMap, toMigratedIpsMap) {
			t.Errorf("TestGetToMigratedIpsByIDC() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				toMigratedIpsMap, expectToMigratedIpsMap)
		}
	}
	// case 5
	{
		fromIDCIpsMap := map[string][]string{
			"huawei": []string{
				"127.1.0.1",
				"127.1.0.2",
				"127.1.0.3",
				"127.1.0.4",
			},
			"aliyun": []string{
				"127.4.0.1",
				"127.4.0.2",
				"127.4.0.3",
				"127.4.0.4",
			},
			"bx": []string{
				"127.3.0.1",
				"127.3.0.2",
			},
		}
		toIDCIpsMap := map[string][]string{}
		num := 5
		expectToMigratedIpsMap := map[string][]string{
			"huawei": []string{
				"127.1.0.1",
				"127.1.0.2",
			},
			"aliyun": []string{
				"127.4.0.1",
				"127.4.0.2",
			},
			"bx": []string{
				"127.3.0.1",
			},
		}
		toMigratedIpsMap := getToMigratedIpsByIDC(fromIDCIpsMap, toIDCIpsMap, num)

		if !reflect.DeepEqual(expectToMigratedIpsMap, toMigratedIpsMap) {
			t.Errorf("TestGetToMigratedIpsByIDC() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				toMigratedIpsMap, expectToMigratedIpsMap)
		}
	}
}
