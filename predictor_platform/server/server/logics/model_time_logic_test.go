package logics

import (
	"fmt"
	"reflect"
	"server/conf"
	"server/env"
	"server/mock"
	"server/schema"
	"server/server/dao"
	"server/util"
	"testing"
	"time"
)

func TestModelUpdateTimeWithinWeek(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./model_time_logic_test.db"
	db := env.InitMysql(conf)
	dao.SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)

	db.Create(&schema.Model{Name: "model_a", Desc: "模型A", Path: "/data0/dummy/path"}) // Mid: 1
	db.Create(&schema.Model{Name: "model_b", Desc: "模型B", Path: "/data0/dummy/path"}) // Mid: 2
	db.Create(&schema.Model{Name: "model_c", Desc: "模型C", Path: "/data0/dummy/path"}) // Mid: 3
	db.Create(&schema.Model{Name: "model_d", Desc: "模型D", Path: "/data0/dummy/path"}) // Mid: 4

	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20201011_090000", Desc: "Validated"})              // ID: 1
	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20201020_090000", Desc: "Validated"})              // ID: 2
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "20201020_090000", Desc: "Validated", IsLocked: 1}) // ID: 3
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "20201020_010000", Desc: "Validated"})              // ID: 4
	db.Create(&schema.ModelHistory{ModelName: "model_c", Timestamp: "20201020_091000", Desc: "Validated"})              // ID: 5
	db.Create(&schema.ModelHistory{ModelName: "model_d", Timestamp: "20201001_010000", Desc: "Validated"})              // ID: 6

	// case 1
	{

		modelName := "model_a"
		conf.ModelTimingRange = -365 * 10
		seconds := util.GetTimestampInterval("20201020_090000", "20201011_090000") +
			util.GetTimestampInterval(time.Now().Format("20060102_150405"), "20201020_090000")
		expectModelUpdateTimingInfo := util.ModelUpdateTimingInfo{
			ModelName:             "model_a",
			ModelUpdateTimeWeekly: seconds / int64(2),
			LastestTimestampArray: []string{
				"20201020_090000",
				"20201011_090000",
			},
			ModelChannel: "模型A",
		}
		modelUpdateTimingInfo := ModelUpdateTimeWithinWeek(modelName, conf)

		if !reflect.DeepEqual(expectModelUpdateTimingInfo, modelUpdateTimingInfo) {
			t.Errorf("TestModelUpdateTimeWithinWeek() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				modelUpdateTimingInfo, expectModelUpdateTimingInfo)
		}
	}
	// case 2
	{
		modelName := "model_c"
		conf.ModelTimingRange = -365 * 10
		seconds := util.GetTimestampInterval(time.Now().Format("20060102_150405"), "20201020_091000")
		expectModelUpdateTimingInfo := util.ModelUpdateTimingInfo{
			ModelName:             "model_c",
			ModelUpdateTimeWeekly: seconds / int64(1),
			LastestTimestampArray: []string{
				"20201020_091000",
			},
			ModelChannel: "模型C",
		}
		modelUpdateTimingInfo := ModelUpdateTimeWithinWeek(modelName, conf)

		if !reflect.DeepEqual(expectModelUpdateTimingInfo, modelUpdateTimingInfo) {
			t.Errorf("TestModelUpdateTimeWithinWeek() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				modelUpdateTimingInfo, expectModelUpdateTimingInfo)
		}
	}
	// case 3
	{
		modelName := "model_e"
		conf.ModelTimingRange = -365 * 10
		expectModelUpdateTimingInfo := util.ModelUpdateTimingInfo{
			ModelName:             "model_e",
			ModelUpdateTimeWeekly: 0,
			LastestTimestampArray: nil,
			ModelChannel:          "",
		}
		modelUpdateTimingInfo := ModelUpdateTimeWithinWeek(modelName, conf)

		if !reflect.DeepEqual(expectModelUpdateTimingInfo, modelUpdateTimingInfo) {
			t.Errorf("TestModelUpdateTimeWithinWeek() failed, \n[Got]:\n%v,\n[Expect]:\n%v",
				modelUpdateTimingInfo, expectModelUpdateTimingInfo)
		}
	}
	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}

func TestModelListUpdateTimeWithinWeek(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./model_time_logic_test.db"
	db := env.InitMysql(conf)
	dao.SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)

	db.Create(&schema.Model{Name: "model_a", Desc: "模型A", Path: "/data0/dummy/path"}) // Mid: 1
	db.Create(&schema.Model{Name: "model_b", Desc: "模型B", Path: "/data0/dummy/path"}) // Mid: 2
	db.Create(&schema.Model{Name: "model_c", Desc: "模型C", Path: "/data0/dummy/path"}) // Mid: 3
	db.Create(&schema.Model{Name: "model_d", Desc: "模型D", Path: "/data0/dummy/path"}) // Mid: 4

	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20201011_090000", Desc: "Validated"})              // ID: 1
	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20201020_090000", Desc: "Validated"})              // ID: 2
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "20201020_090000", Desc: "Validated", IsLocked: 1}) // ID: 3
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "20201020_010000", Desc: "Validated"})              // ID: 4
	db.Create(&schema.ModelHistory{ModelName: "model_c", Timestamp: "20201020_091000", Desc: "Validated"})              // ID: 5
	db.Create(&schema.ModelHistory{ModelName: "model_d", Timestamp: "20201001_010000", Desc: "Validated"})              // ID: 6

	// case 1
	{
		modelList := []string{"model_a", "model_b"}
		conf.ModelTimingRange = -365 * 10
		seconds1 := util.GetTimestampInterval("20201020_090000", "20201011_090000") +
			util.GetTimestampInterval(time.Now().Format("20060102_150405"), "20201020_090000")
		seconds2 := util.GetTimestampInterval("20201020_090000", "20201020_090000") +
			util.GetTimestampInterval(time.Now().Format("20060102_150405"), "20201020_010000")
		expectModelUpdateTimingInfos := []util.ModelUpdateTimingInfo{
			util.ModelUpdateTimingInfo{
				ModelName:             "model_a",
				ModelUpdateTimeWeekly: seconds1 / int64(2),
				LastestTimestampArray: []string{
					"20201020_090000",
					"20201011_090000",
				},
				ModelChannel: "模型A",
			},
			util.ModelUpdateTimingInfo{
				ModelName:             "model_b",
				ModelUpdateTimeWeekly: seconds2 / int64(2),
				LastestTimestampArray: []string{
					"20201020_090000",
					"20201020_010000",
				},
				ModelChannel: "模型B",
			},
		}
		expectReportMap := map[string]interface{}{
			"Date":        time.Now().Format("2006-01-02 15:04:05"),
			"TableHeader": []string{"模型业务线", "模型名称", "7天平均更新频率", "最近一次更新时间"},
			"ModelData": []interface{}{
				[]string{"模型A", "model_a", util.SecondsToHM(fmt.Sprintf("%d", seconds1/int64(2))), "2020-10-20 09:00:00"},
				[]string{"模型B", "model_b", util.SecondsToHM(fmt.Sprintf("%d", seconds2/int64(2))), "2020-10-20 09:00:00"},
			},
		}

		modelUpdateTimingInfos, reportMap := ModelListUpdateTimeWithinWeek(modelList, conf)

		if !reflect.DeepEqual(expectModelUpdateTimingInfos, modelUpdateTimingInfos) ||
			!reflect.DeepEqual(expectReportMap, reportMap) {
			t.Errorf("TestModelListUpdateTimeWithinWeek() failed, \n[Got]:\n%v\n%v,\n[Expect]:\n%v\n%v",
				modelUpdateTimingInfos, reportMap, expectModelUpdateTimingInfos, expectReportMap)
		}
	}
	// case 1
	{
		modelList := []string{"model_a"}
		conf.ModelTimingRange = -365 * 10
		seconds1 := util.GetTimestampInterval("20201020_090000", "20201011_090000") +
			util.GetTimestampInterval(time.Now().Format("20060102_150405"), "20201020_090000")
		expectModelUpdateTimingInfos := []util.ModelUpdateTimingInfo{
			util.ModelUpdateTimingInfo{
				ModelName:             "model_a",
				ModelUpdateTimeWeekly: seconds1 / int64(2),
				LastestTimestampArray: []string{
					"20201020_090000",
					"20201011_090000",
				},
				ModelChannel: "模型A",
			},
		}
		expectReportMap := map[string]interface{}{
			"Date":        time.Now().Format("2006-01-02 15:04:05"),
			"TableHeader": []string{"模型业务线", "模型名称", "7天平均更新频率", "最近一次更新时间"},
			"ModelData": []interface{}{
				[]string{"模型A", "model_a", util.SecondsToHM(fmt.Sprintf("%d", seconds1/int64(2))), "2020-10-20 09:00:00"},
			},
		}

		modelUpdateTimingInfos, reportMap := ModelListUpdateTimeWithinWeek(modelList, conf)

		if !reflect.DeepEqual(expectModelUpdateTimingInfos, modelUpdateTimingInfos) ||
			!reflect.DeepEqual(expectReportMap, reportMap) {
			t.Errorf("TestModelListUpdateTimeWithinWeek() failed, \n[Got]:\n%v\n%v,\n[Expect]:\n%v\n%v",
				modelUpdateTimingInfos, reportMap, expectModelUpdateTimingInfos, expectReportMap)
		}
	}
	// case 1
	{
		modelList := []string{}
		conf.ModelTimingRange = -365 * 10
		var timingInfos []util.ModelUpdateTimingInfo
		var modelDatas []interface{}
		expectModelUpdateTimingInfos := timingInfos
		expectReportMap := map[string]interface{}{
			"Date":        time.Now().Format("2006-01-02 15:04:05"),
			"TableHeader": []string{"模型业务线", "模型名称", "7天平均更新频率", "最近一次更新时间"},
			"ModelData":   modelDatas,
		}

		modelUpdateTimingInfos, reportMap := ModelListUpdateTimeWithinWeek(modelList, conf)
		if !reflect.DeepEqual(expectModelUpdateTimingInfos, modelUpdateTimingInfos) ||
			!reflect.DeepEqual(expectReportMap, reportMap) {
			t.Errorf("TestModelListUpdateTimeWithinWeek() failed, \n[Got]:\n%v\n%+v,\n[Expect]:\n%v\n%+v",
				modelUpdateTimingInfos, reportMap, expectModelUpdateTimingInfos, expectReportMap)
		}
	}

	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}
