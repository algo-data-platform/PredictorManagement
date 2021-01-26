package logics

import (
	"content_service/conf"
	"content_service/env"
	"content_service/mock"
	"content_service/schema"
	//"reflect"
	"testing"
)

func TestFetchHdfsModelsFromDb(t *testing.T) {
	// create test conf, env, service
	conf := &conf.Conf{}
	conf.Db.Driver = "sqlite3"
	conf.Db.Name = "./hdfs_logic_test.db"
	env := env.New(conf)

	// insert test data in test db
	db := env.Db
	mock.AutoMigrateAll(env)
	mock.CleanUp(env)

	db.Create(&schema.Model{Name: "model_a", Path: "/data0/dummy/path"}) // Mid: 1
	db.Create(&schema.Model{Name: "model_b", Path: "/data0/dummy/path"}) // Mid: 2
	db.Create(&schema.Model{Name: "model_c", Path: "/data0/dummy/path"}) // Mid: 3
	db.Create(&schema.Model{Name: "model_d", Path: "/data0/dummy/path"}) // Mid: 4

	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "ts_1", Desc: "Validated"})              // ID: 1
	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "ts_2", Desc: "Validated"})              // ID: 2
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "ts_1", Desc: "Validated", IsLocked: 1}) // ID: 3
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "ts_2", Desc: "Validated"})              // ID: 4
	db.Create(&schema.ModelHistory{ModelName: "model_c", Timestamp: "ts_3", Desc: "Validated"})              // ID: 5
	db.Create(&schema.ModelHistory{ModelName: "model_d", Timestamp: "ts_4", Desc: "Validated"})              // ID: 6

	// case 1
	{
		res, err := FetchHdfsModelsFromDb(env)
		if err != nil || len(res) != 0 {
			t.Errorf("TestFetchHdfsModelsFromDb() failed: res:%v, err:%v",
				res, err)
		}
	}
	db.Create(&schema.ModelHistory{ModelName: "model_d", Timestamp: "ts_5", Desc: "hdfs://ns3-backup/user/adbot/wbl/mission_48/app_21/model_d/model_d-ts_5"}) // ID: 7
	db.Create(&schema.ModelHistory{ModelName: "model_d", Timestamp: "ts_6", Desc: "hdfs://ns3-backup/user/adbot/wbl/mission_48/app_21/model_d/model_d-ts_6"}) // ID: 8
	// case 2
	{
		res, err := FetchHdfsModelsFromDb(env)
		expect_res := []schema.ModelHistory{
			schema.ModelHistory{ID: 7, ModelName: "model_d", Timestamp: "ts_5", Desc: "hdfs://ns3-backup/user/adbot/wbl/mission_48/app_21/model_d/model_d-ts_5"},
			schema.ModelHistory{ID: 8, ModelName: "model_d", Timestamp: "ts_6", Desc: "hdfs://ns3-backup/user/adbot/wbl/mission_48/app_21/model_d/model_d-ts_6"},
		}
		if err != nil || !mock.IsEqualModelHistorySlice(res, expect_res) {
			t.Errorf("TestFetchHdfsModelsFromDb() failed: res:%v, expect_res: %v, err:%v",
				res, expect_res, err)
		}
	}
	db.Create(&schema.ModelHistory{ModelName: "model_d", Timestamp: "ts_7", Desc: "ddhdfs://ns3-backup/user/adbot/wbl/mission_48/app_21/model_d/model_d-ts_7"}) // ID: 9
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "ts_3", Desc: "hdfs://ns3-backup/user/adbot/wbl/mission_48/app_21/model_b/model_b-ts_3"})   // ID: 10
	// case 3
	{
		res, err := FetchHdfsModelsFromDb(env)
		expect_res := []schema.ModelHistory{
			schema.ModelHistory{ID: 7, ModelName: "model_d", Timestamp: "ts_5", Desc: "hdfs://ns3-backup/user/adbot/wbl/mission_48/app_21/model_d/model_d-ts_5"},
			schema.ModelHistory{ID: 8, ModelName: "model_d", Timestamp: "ts_6", Desc: "hdfs://ns3-backup/user/adbot/wbl/mission_48/app_21/model_d/model_d-ts_6"},
			schema.ModelHistory{ID: 10, ModelName: "model_b", Timestamp: "ts_3", Desc: "hdfs://ns3-backup/user/adbot/wbl/mission_48/app_21/model_b/model_b-ts_3"},
		}
		if err != nil || !mock.IsEqualModelHistorySlice(res, expect_res) {
			t.Errorf("TestFetchHdfsModelsFromDb() failed: res:%v, expect_res: %v, err:%v",
				res, expect_res, err)
		}
	}

	// clean up db
	mock.CleanUp(env)
}

func TestUpdateModelHistoryStatusById(t *testing.T) {
	// create test conf, env, service
	conf := &conf.Conf{}
	conf.Db.Driver = "sqlite3"
	conf.Db.Name = "./hdfs_logic_test.db"
	env := env.New(conf)

	// insert test data in test db
	db := env.Db
	mock.AutoMigrateAll(env)
	mock.CleanUp(env)

	db.Create(&schema.Model{Name: "model_a", Path: "/data0/dummy/path"}) // Mid: 1
	db.Create(&schema.Model{Name: "model_b", Path: "/data0/dummy/path"}) // Mid: 2
	db.Create(&schema.Model{Name: "model_c", Path: "/data0/dummy/path"}) // Mid: 3
	db.Create(&schema.Model{Name: "model_d", Path: "/data0/dummy/path"}) // Mid: 4

	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "ts_1", Desc: "Validated"})              // ID: 1
	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "ts_2", Desc: "Validated"})              // ID: 2
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "ts_1", Desc: "Validated", IsLocked: 1}) // ID: 3
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "ts_2", Desc: "Validated"})              // ID: 4
	db.Create(&schema.ModelHistory{ModelName: "model_c", Timestamp: "ts_3", Desc: "Validated"})              // ID: 5
	db.Create(&schema.ModelHistory{ModelName: "model_d", Timestamp: "ts_4", Desc: "Validated"})              // ID: 6
	var mhid uint
	// case 1
	{
		mhid = 2
		desc := ""
		err := UpdateModelHistoryStatusById(env, mhid, desc)
		if err != nil {
			t.Errorf("TestUpdateModelHistoryStatusById() failed: err:%v",
				err)
		}
		var updated_mh schema.ModelHistory
		db.Where(schema.ModelHistory{ID: mhid}).Find(&updated_mh)
		errs := db.GetErrors()
		expected_mh := schema.ModelHistory{ID: 2, ModelName: "model_a", Timestamp: "ts_2", Desc: desc}
		if len(errs) > 0 || !mock.IsEqualModelHistory(updated_mh, expected_mh) {
			t.Errorf("TestUpdateModelHistoryStatusById() failed: updated_mh:%v, expected_mh:%v, errs:%v",
				updated_mh, expected_mh, errs)
		}
	}
	// case 2
	{
		mhid = 2
		desc := "Validated"
		err := UpdateModelHistoryStatusById(env, mhid, desc)
		if err != nil {
			t.Errorf("TestUpdateModelHistoryStatusById() failed: err:%v",
				err)
		}
		var updated_mh schema.ModelHistory
		db.Where(schema.ModelHistory{ID: mhid}).Find(&updated_mh)
		errs := db.GetErrors()
		expected_mh := schema.ModelHistory{ID: 2, ModelName: "model_a", Timestamp: "ts_2", Desc: desc}
		if len(errs) > 0 || !mock.IsEqualModelHistory(updated_mh, expected_mh) {
			t.Errorf("TestUpdateModelHistoryStatusById() failed: updated_mh:%v, expected_mh:%v, errs:%v",
				updated_mh, expected_mh, errs)
		}
	}

	// clean up db
	mock.CleanUp(env)
}
