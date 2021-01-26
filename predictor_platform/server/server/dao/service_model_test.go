package dao

import (
	"reflect"
	"server/conf"
	"server/env"
	"server/mock"
	"server/schema"
	"testing"
)

func TestGetServiceModelMap(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./service_model_test.db"
	db := env.InitMysql(conf)
	SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)

	db.Create(&schema.Model{Name: "model_a", Path: "/data0/dummy/path"}) // Mid: 1
	db.Create(&schema.Model{Name: "model_b", Path: "/data0/dummy/path"}) // Mid: 2
	db.Create(&schema.Model{Name: "model_c", Path: "/data0/dummy/path"}) // Mid: 3
	db.Create(&schema.Model{Name: "model_d", Path: "/data0/dummy/path"}) // Mid: 4

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3

	db.Create(&schema.ServiceModel{Sid: 1, Mid: 1, Desc: "service_1 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 2, Desc: "service_1 -> model_b"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 3, Desc: "service_1 -> model_c"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 1, Desc: "service_2 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 4, Desc: "service_2 -> model_d"})

	// case 1
	{
		expectServiceModelMap := map[string][]string{
			"service_1": []string{
				"model_a",
				"model_b",
				"model_c",
			},
			"service_2": []string{
				"model_a",
				"model_d",
			},
		}
		serviceModelMap := GetServiceModelMap()
		if !reflect.DeepEqual(expectServiceModelMap, serviceModelMap) {
			t.Errorf("TestGetServiceModelMap() failed,\n[Got]:\n%v,\n[Expect]:\n%v",
				serviceModelMap, expectServiceModelMap)
		}
	}

	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}
