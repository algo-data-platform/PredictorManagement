package logics

import (
	"reflect"
	"server/conf"
	"server/env"
	"server/metrics"
	"server/mock"
	"server/schema"
	"server/server/dao"
	"testing"
	"time"
)

func TestGetDBModelFullNameMap(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./model_time_logic_test.db"
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

	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20201011_090000", Desc: "Validated"})              // ID: 1
	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20201020_090000", Desc: "Validated"})              // ID: 2
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "20201020_090000", Desc: "Validated"})              // ID: 3
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "20201020_010000", Desc: "Validated", IsLocked: 1}) // ID: 4
	db.Create(&schema.ModelHistory{ModelName: "model_c", Timestamp: "20201020_091000", Desc: "Validated"})              // ID: 5
	db.Create(&schema.ModelHistory{ModelName: "model_d", Timestamp: "20201001_010000", Desc: "Validated"})              // ID: 6

	// case 1
	{
		expectServiceModelFullNameMap := map[string][]string{
			"service_1": []string{
				"model_a-20201020_090000",
				"model_b-20201020_010000",
				"model_c-20201020_091000",
			},
			"service_2": []string{
				"model_a-20201020_090000",
				"model_d-20201001_010000",
			},
		}
		serviceModelFullNameMap, err := GetDBServiceModelFullNameMap()

		if err != nil || !reflect.DeepEqual(expectServiceModelFullNameMap, serviceModelFullNameMap) {
			t.Errorf("TestGetDBServiceModelFullNameMap() failed, err: %v, \n[Got]:\n%v,\n[Expect]:\n%v",
				err, serviceModelFullNameMap, expectServiceModelFullNameMap)
		}
	}
	// case 2
	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20201023_110000", Desc: "Abandoned"})
	{
		expectServiceModelFullNameMap := map[string][]string{
			"service_1": []string{
				"model_a-20201020_090000",
				"model_b-20201020_010000",
				"model_c-20201020_091000",
			},
			"service_2": []string{
				"model_a-20201020_090000",
				"model_d-20201001_010000",
			},
		}
		serviceModelFullNameMap, err := GetDBServiceModelFullNameMap()

		if err != nil || !reflect.DeepEqual(expectServiceModelFullNameMap, serviceModelFullNameMap) {
			t.Errorf("TestGetDBServiceModelFullNameMap() failed, err: %v, \n[Got]:\n%v,\n[Expect]:\n%v",
				err, serviceModelFullNameMap, expectServiceModelFullNameMap)
		}
	}
	// case 3
	db.Create(&schema.Model{Name: "model_e", Desc: "模型E", Path: "/data0/dummy/path"}) // Mid: 5
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 5, Desc: "service_2 -> model_e"})
	{
		expectServiceModelFullNameMap := map[string][]string{
			"service_1": []string{
				"model_a-20201020_090000",
				"model_b-20201020_010000",
				"model_c-20201020_091000",
			},
			"service_2": []string{
				"model_a-20201020_090000",
				"model_d-20201001_010000",
			},
		}
		serviceModelFullNameMap, err := GetDBServiceModelFullNameMap()

		if err != nil || !reflect.DeepEqual(expectServiceModelFullNameMap, serviceModelFullNameMap) {
			t.Errorf("TestGetDBServiceModelFullNameMap() failed, err: %v, \n[Got]:\n%v,\n[Expect]:\n%v",
				err, serviceModelFullNameMap, expectServiceModelFullNameMap)
		}
	}
	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}

func TestGetMaxLoadInterval(t *testing.T) {
	metrics.InitMetrics()
	lastFourMinVersion := time.Unix(time.Now().Unix()-240, 0).Format("20060102_150405")
	lastFiveMinVersion := time.Unix(time.Now().Unix()-300, 0).Format("20060102_150405")
	lastSixMinVersion := time.Unix(time.Now().Unix()-360, 0).Format("20060102_150405")

	// case 1
	{
		fullModelList := []string{
			"model_a-20201020_090000",
			"model_b-20201020_010000",
			"model_c-20201020_091000",
		}
		versionTime, _ := time.ParseInLocation("20060102_150405", "20201020_010000", time.Local)
		expectMaxLoadInterval := time.Now().Sub(versionTime).Seconds()
		resMaxLoadInterval := GetMaxLoadInterval(fullModelList)

		if int(expectMaxLoadInterval) != resMaxLoadInterval {
			t.Errorf("TestGetMaxLoadInterval() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				resMaxLoadInterval, expectMaxLoadInterval)
		}
	}
	// case 2
	{
		fullModelList := []string{
			"model_b-" + lastFourMinVersion,
			"model_c-" + lastFiveMinVersion,
			"model_c-" + lastSixMinVersion,
		}
		expectMaxLoadInterval := 360
		resMaxLoadInterval := GetMaxLoadInterval(fullModelList)

		if expectMaxLoadInterval != resMaxLoadInterval {
			t.Errorf("TestGetMaxLoadInterval() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				resMaxLoadInterval, expectMaxLoadInterval)
		}
	}
	// case 3
	{
		fullModelList := []string{
			"model_a-20201024",
			"model_b-" + lastFourMinVersion,
			"model_c-" + lastFiveMinVersion,
			"model_c-" + lastSixMinVersion,
		}
		expectMaxLoadInterval := 3600
		resMaxLoadInterval := GetMaxLoadInterval(fullModelList)

		if expectMaxLoadInterval != resMaxLoadInterval {
			t.Errorf("TestGetMaxLoadInterval() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				resMaxLoadInterval, expectMaxLoadInterval)
		}
	}
	// case 4
	{
		fullModelList := []string{
			"model_a-20201020",
			"model_a",
			"model_b-" + lastFourMinVersion,
			"model_c-" + lastFiveMinVersion,
			"model_c-" + lastSixMinVersion,
		}
		expectMaxLoadInterval := 3600
		resMaxLoadInterval := GetMaxLoadInterval(fullModelList)

		if expectMaxLoadInterval != resMaxLoadInterval {
			t.Errorf("TestGetMaxLoadInterval() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				resMaxLoadInterval, expectMaxLoadInterval)
		}
	}
}
