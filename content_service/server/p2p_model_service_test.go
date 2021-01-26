package server

import (
	"content_service/conf"
	"content_service/env"
	"content_service/mock"
	"content_service/schema"
	"reflect"
	"sort"
	"sync"
	"testing"
	"fmt"
)

func TestFindModelsByServiceToPull(t *testing.T) {
	tables := []struct {
		db_service_models map[string]schema.ModelHistory
		disk_models       map[string]bool
		res               []string
	}{
		// case 1
		{
			// db data
			map[string]schema.ModelHistory{
				"model_a": schema.ModelHistory{ModelName: "model_a", Timestamp: "20200101_000001", IsLocked: 0},
				"model_b": schema.ModelHistory{ModelName: "model_b", Timestamp: "20200101_000002", IsLocked: 0},
				"model_c": schema.ModelHistory{ModelName: "model_c", Timestamp: "20200101_000003", IsLocked: 0},
			},
			// disk data
			map[string]bool{
				"model_a-20200101_000001": true,
				"model_b-20200101_000002": true,
				"model_c-20200101_000003": true,
			},
			// disk data same as db data, expected empty result
			[]string{},
		},
		// case 2
		{
			// db data
			map[string]schema.ModelHistory{
				"model_a": schema.ModelHistory{ModelName: "model_a", Timestamp: "20200101_000001", IsLocked: 0},
				"model_b": schema.ModelHistory{ModelName: "model_b", Timestamp: "20200101_000002", IsLocked: 0},
				"model_c": schema.ModelHistory{ModelName: "model_c", Timestamp: "20200101_000003", IsLocked: 0},
			},
			// disk data
			map[string]bool{
				"model_a-20200101_000001": true,
				"model_b-20200101_000002": true,
				"model_c-20200101_000003": true,
				"model_c-20200101_000002": true,
			},
			// disk data covers db data, expected empty result
			[]string{},
		},
		// case 3
		{
			// db data
			map[string]schema.ModelHistory{
				"model_a": schema.ModelHistory{ModelName: "model_a", Timestamp: "20200101_000001", IsLocked: 0},
				"model_b": schema.ModelHistory{ModelName: "model_b", Timestamp: "20200101_000002", IsLocked: 0},
				"model_c": schema.ModelHistory{ModelName: "model_c", Timestamp: "20200101_000003", IsLocked: 0},
				"model_d": schema.ModelHistory{ModelName: "model_d", Timestamp: "20200101_000004", IsLocked: 0},
				"model_e": schema.ModelHistory{ModelName: "model_e", Timestamp: "20200101_000001", IsLocked: 0},
			},
			// disk data
			map[string]bool{
				"model_a-20200101_000001": true,
				"model_b-20200101_000002": true,
				"model_c-20200101_000003": true,
				"model_c-20200101_000002": true,
			},
			// db has more data, expect complement data from db
			[]string{"model_d-20200101_000004", "model_e-20200101_000001"},
		},
		// case 4
		{
			// db data
			map[string]schema.ModelHistory{},
			// disk data
			map[string]bool{
				"model_a-20200101_000001": true,
				"model_b-20200101_000002": true,
				"model_c-20200101_000003": true,
			},
			// db has none, expect empty result
			[]string{},
		},
		// case 5
		{
			// db data
			map[string]schema.ModelHistory{
				"model_a": schema.ModelHistory{ModelName: "model_a", Timestamp: "20200101_000001", IsLocked: 0},
				"model_b": schema.ModelHistory{ModelName: "model_b", Timestamp: "20200101_000002", IsLocked: 0},
				"model_c": schema.ModelHistory{ModelName: "model_c", Timestamp: "20200101_000003", IsLocked: 0},
				"model_d": schema.ModelHistory{ModelName: "model_d", Timestamp: "20200101_000004", IsLocked: 0},
				"model_e": schema.ModelHistory{ModelName: "model_e", Timestamp: "20200101_000001", IsLocked: 0},
			},
			// disk data
			map[string]bool{
				"model_b-20200101_000002": true,
				"model_c-20200101_000003": true,
				"model_c-20200101_000002": true,
				"model_e-20200101_000001": true,
			},
			// test deduplication from db data
			[]string{"model_a-20200101_000001", "model_d-20200101_000004"},
		},
	}

	test_service := NewP2PModelService()
	for _, table := range tables {
		res := test_service.findModelsByServiceToPull(table.db_service_models, table.disk_models)
		// must sort the results here as we don't care about the order
		sort.Strings(res)
		sort.Strings(table.res)
		if !reflect.DeepEqual(res, table.res) {
			t.Errorf("TestFindModelsBySereviceToPull(%v, %v) failed, got: %v, want: %v\n",
				table.db_service_models, table.disk_models, res, table.res)
		}
	}
}

func TestFetchDbModelsByService(t *testing.T) {
	// create test conf, env, service
	conf := &conf.Conf{}
	conf.Db.Driver = "sqlite3"
	conf.Db.Name = "./model_service_test.db"
	env := env.New(conf)
	service := NewP2PModelService()

	// insert test data in test db
	db := env.Db
	mock.AutoMigrateAll(env)
	mock.CleanUp(env)

	db.Create(&schema.Host{Ip: "127.1.0.1"})  // Hid: 1
	db.Create(&schema.Host{Ip: "127.3.0.2"}) // Hid: 2
	db.Create(&schema.Host{Ip: "127.2.0.3"})  // Hid: 3
	db.Create(&schema.Host{Ip: "127.4.0.4"})  // Hid: 4

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3

	db.Create(&schema.Model{Name: "model_a", Path: "/data0/dummy/path"}) // Mid: 1
	db.Create(&schema.Model{Name: "model_b", Path: "/data0/dummy/path"}) // Mid: 2
	db.Create(&schema.Model{Name: "model_c", Path: "/data0/dummy/path"}) // Mid: 3
	db.Create(&schema.Model{Name: "model_d", Path: "/data0/dummy/path"}) // Mid: 4

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.3.0.2 -> service_1"})
	db.Create(&schema.HostService{Hid: 3, Sid: 1, Desc: "127.2.0.3 -> service_1"})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2"})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3"})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.4.0.4 -> service_3"})

	db.Create(&schema.ServiceModel{Sid: 1, Mid: 1, Desc: "service_1 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 2, Desc: "service_1 -> model_b"})
	db.Create(&schema.ServiceModel{Sid: 1, Mid: 3, Desc: "service_1 -> model_c"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 1, Desc: "service_2 -> model_a"})
	db.Create(&schema.ServiceModel{Sid: 2, Mid: 4, Desc: "service_2 -> model_d"})

	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20200101_000001", Desc: "Validated"})              // ID: 1
	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20200101_000002", Desc: "Validated"})              // ID: 2
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "20200101_000001", Desc: "Validated", IsLocked: 1}) // ID: 3
	db.Create(&schema.ModelHistory{ModelName: "model_b", Timestamp: "20200101_000002", Desc: "Validated"})              // ID: 4
	db.Create(&schema.ModelHistory{ModelName: "model_c", Timestamp: "20200101_000003", Desc: "Validated"})              // ID: 5
	db.Create(&schema.ModelHistory{ModelName: "model_d", Timestamp: "20200101_000004", Desc: "Validated"})              // ID: 6

	env.LocalIp = "127.3.0.2" // just not equal to env.Conf.ValidateService.Host
	env.Conf.ValidateService.Host = "127.1.57.206"
	var sid uint
	// case 1
	{
		sid = 1
		res, err := service.fetchDbModelsByService(env, sid)
		expected_res := map[string]schema.ModelHistory{
			"model_a": schema.ModelHistory{ModelName: "model_a", Timestamp: "20200101_000002", IsLocked: 0},
			// expect getting back model_a with the latest timestamp
			"model_b": schema.ModelHistory{ModelName: "model_b", Timestamp: "20200101_000001", IsLocked: 1},
			// expect getting back model_b with the locked older timestamp, even when there is a newer timestamp
			"model_c": schema.ModelHistory{ModelName: "model_c", Timestamp: "20200101_000003", IsLocked: 0},
		}
		if err != nil || !mock.IsEqualModelHistoryMap(res, expected_res) {
			t.Errorf("TestFetchDbModelsByService() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}
	// case 2
	{
		sid = 2
		res, err := service.fetchDbModelsByService(env, sid)
		expected_res := map[string]schema.ModelHistory{
			"model_a": schema.ModelHistory{ModelName: "model_a", Timestamp: "20200101_000002", IsLocked: 0},
			// expect getting back model_b with the locked older timestamp, even when there is a newer timestamp
			"model_d": schema.ModelHistory{ModelName: "model_d", Timestamp: "20200101_000004", IsLocked: 0},
		}
		if err != nil || !mock.IsEqualModelHistoryMap(res, expected_res) {
			t.Errorf("TestFetchDbModelsByService() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}
	// case 3
	{
		sid = 3
		res, err := service.fetchDbModelsByService(env, sid)
		if err != nil || len(res) != 0 {
			t.Errorf("TestFetchDbModelsByService() failed:\ngot:\n\t%v\nwant:empty map\nerr:%v\n", res, err)
		}
	}

	// db updated
	db.Create(&schema.ServiceModel{Sid: 3, Mid: 1, Desc: "service_3 -> model_a"})
	// case 4
	{
		sid = 3
		res, err := service.fetchDbModelsByService(env, sid)
		expected_res := map[string]schema.ModelHistory{
			"model_a": schema.ModelHistory{ModelName: "model_a", Timestamp: "20200101_000002", IsLocked: 0},
		}
		if err != nil || !mock.IsEqualModelHistoryMap(res, expected_res) {
			t.Errorf("TestFetchDbModelsByService() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}
	// db updated with a newer timestamp for model_a
	db.Create(&schema.ModelHistory{ModelName: "model_a", Timestamp: "20200101_000003", Desc: "Validated"}) // ID: 7
	// case 4.2: "127.4.0.4" after db update
	{
		sid = 3
		res, err := service.fetchDbModelsByService(env, sid)
		expected_res := map[string]schema.ModelHistory{
			"model_a": schema.ModelHistory{ModelName: "model_a", Timestamp: "20200101_000003", IsLocked: 0},
		}
		if err != nil || !mock.IsEqualModelHistoryMap(res, expected_res) {
			t.Errorf("TestFetchDbModelsByService() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}

	// clean up db
	mock.CleanUp(env)
}

func TestCleanServiceCheckTimesMap(t *testing.T) {
	service := NewP2PModelService()
	var expectedServiceCheTimesMap map[string]int32
	var resServiceCheckTimesMap map[string]int32
	// case 1
	{
		service.ServiceCheckTimesMap = sync.Map{}
		service.ServiceCheckTimesMap.Store("service_1", int32(2))
		service.ServiceCheckTimesMap.Store("service_2", int32(3))
		dbServices := []*schema.Service{
			&schema.Service{Name: "service_1"},
			&schema.Service{Name: "service_2"},
		}
		service.cleanServiceCheckTimesMap(dbServices)
		resServiceCheckTimesMap = make(map[string]int32)
		service.ServiceCheckTimesMap.Range(func(k, v interface{}) bool {
			service_name, _ := k.(string)
			num, _ := v.(int32)
			resServiceCheckTimesMap[service_name] = num
			return true
		})
		expectedServiceCheTimesMap = map[string]int32{
			"service_1": 2,
			"service_2": 3,
		}

		if !reflect.DeepEqual(resServiceCheckTimesMap, expectedServiceCheTimesMap) {
			t.Errorf("TestCleanServiceCheckTimesMap() failed: resServiceCheckTimesMap: %v, expectedServiceCheTimesMap:%v",
				resServiceCheckTimesMap, expectedServiceCheTimesMap)
		}
	}
	// case 2
	{
		service.ServiceCheckTimesMap = sync.Map{}
		service.ServiceCheckTimesMap.Store("service_1", int32(2))
		service.ServiceCheckTimesMap.Store("service_2", int32(3))
		service.ServiceCheckTimesMap.Store("service_3", int32(1))
		dbServices := []*schema.Service{
			&schema.Service{Name: "service_1"},
			&schema.Service{Name: "service_2"},
		}
		service.cleanServiceCheckTimesMap(dbServices)
		resServiceCheckTimesMap = make(map[string]int32)
		service.ServiceCheckTimesMap.Range(func(k, v interface{}) bool {
			service_name, _ := k.(string)
			num, _ := v.(int32)
			resServiceCheckTimesMap[service_name] = num
			return true
		})
		expectedServiceCheTimesMap = map[string]int32{
			"service_1": 2,
			"service_2": 3,
		}

		if !reflect.DeepEqual(resServiceCheckTimesMap, expectedServiceCheTimesMap) {
			t.Errorf("TestCleanServiceCheckTimesMap() failed: resServiceCheckTimesMap: %v, expectedServiceCheTimesMap:%v",
				resServiceCheckTimesMap, expectedServiceCheTimesMap)
		}
	}
	// case 3
	{
		service.ServiceCheckTimesMap = sync.Map{}
		service.ServiceCheckTimesMap.Store("service_1", int32(2))
		service.ServiceCheckTimesMap.Store("service_2", int32(3))
		service.ServiceCheckTimesMap.Store("service_3", int32(1))
		dbServices := []*schema.Service{}
		service.cleanServiceCheckTimesMap(dbServices)
		resServiceCheckTimesMap = make(map[string]int32)
		service.ServiceCheckTimesMap.Range(func(k, v interface{}) bool {
			service_name, _ := k.(string)
			num, _ := v.(int32)
			resServiceCheckTimesMap[service_name] = num
			return true
		})
		expectedServiceCheTimesMap = map[string]int32{}

		if !reflect.DeepEqual(resServiceCheckTimesMap, expectedServiceCheTimesMap) {
			t.Errorf("TestCleanServiceCheckTimesMap() failed: resServiceCheckTimesMap: %v, expectedServiceCheTimesMap:%v",
				resServiceCheckTimesMap, expectedServiceCheTimesMap)
		}
	}

	// case 4
	{
		service.ServiceCheckTimesMap = sync.Map{}
		dbServices := []*schema.Service{
			&schema.Service{Name: "service_1"},
		}
		service.cleanServiceCheckTimesMap(dbServices)
		resServiceCheckTimesMap = make(map[string]int32)
		service.ServiceCheckTimesMap.Range(func(k, v interface{}) bool {
			service_name, _ := k.(string)
			num, _ := v.(int32)
			resServiceCheckTimesMap[service_name] = num
			return true
		})
		expectedServiceCheTimesMap = map[string]int32{}

		if !reflect.DeepEqual(resServiceCheckTimesMap, expectedServiceCheTimesMap) {
			t.Errorf("TestCleanServiceCheckTimesMap() failed: resServiceCheckTimesMap: %v, expectedServiceCheTimesMap:%v",
				resServiceCheckTimesMap, expectedServiceCheTimesMap)
		}
	}

}

func TestFetchStressInfo(t *testing.T) {
	conf := &conf.Conf{}
	conf.Db.Driver = "sqlite3"
	conf.Db.Name = "./model_service_test.db"
	env := env.New(conf)
	test_service := NewP2PModelService()

	db := env.Db
	mock.AutoMigrateAll(env)
	mock.CleanUp(env)

	db.Create(&schema.Host{Ip: "127.1.0.1"})  // Hid: 1
	db.Create(&schema.Host{Ip: "127.3.0.2"}) // Hid: 2
	db.Create(&schema.Host{Ip: "127.2.0.3"})  // Hid: 3
	db.Create(&schema.Host{Ip: "127.4.0.4"})  // Hid: 4

	db.Create(&schema.Model{Name: "model-1"}) // Sid: 1
	db.Create(&schema.Model{Name: "model-2"}) // Sid: 2
	db.Create(&schema.Model{Name: "model-3"}) // Sid: 3

	db.Create(&schema.StressInfo{Hid: 1, Mids: "1,2,3", Qps: "100,200", IsEnable: 1})
	db.Create(&schema.StressInfo{Hid: 2, Mids: "1,2,3", Qps: "100,200,300", IsEnable: 0})
	db.Create(&schema.StressInfo{Hid: 3, Mids: "1", Qps: "100", IsEnable: 1})
	db.Create(&schema.StressInfo{Hid: 4, Mids: "1", Qps: "100", IsEnable: 1})

	// case 1
	{
		env.LocalIp = "127.1.0.1"
		_, _, err := test_service.FetchStressInfo(env)
		if err.Error() != "model and qps is not fit" {
			t.Errorf("FetchStressInfo failed, %v\n", err)
		}
	}

	// case 2
	{
		env.LocalIp = "127.3.0.2"
		model_names_list, qps, err := test_service.FetchStressInfo(env)
		if len(model_names_list) != 0 && len(qps) != 0 {
			t.Errorf("FetchStressInfo failed, %v\n", err)
			fmt.Printf("model is %v, qps is %v", model_names_list, qps)
		}
	}

	// case 3
	{
		env.LocalIp = "127.2.0.3"
		_, _, err := test_service.FetchStressInfo(env)
		if err != nil {
			t.Errorf("FetchStressInfo failed, %v\n", err)
		}
	}

	// case 4
	{
		env.LocalIp = "127.4.0.5"
		model_names_list, qps, err := test_service.FetchStressInfo(env)
		if len(model_names_list) != 0 && len(qps) != 0 {
			t.Errorf("FetchStressInfo failed, %v\n", err)
			fmt.Printf("model is %v, qps is %v", model_names_list, qps)
		}
	}

	// case 5
	{
		env.LocalIp = "127.4.0.4"
		model_names_list, qps, err := test_service.FetchStressInfo(env)
		if err != nil {
			t.Errorf("FetchStressInfo failed, %v\n", err)
			fmt.Printf("model is %v, qps is %v", model_names_list, qps)
		}

		if model_names_list !="model-1" || qps != "100" {
			t.Errorf("FetchStressInfo failed, %v\n", err)
			fmt.Printf("model is %v, qps is %v", model_names_list, qps)
		}
	}
}
