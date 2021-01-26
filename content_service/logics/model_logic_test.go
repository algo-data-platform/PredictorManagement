package logics

import (
	"content_service/api"
	"content_service/common"
	"content_service/conf"
	"content_service/env"
	"content_service/mock"
	"content_service/schema"
	"reflect"
	"testing"
	"sync"
)

func TestGetPayloadByServiceModels(t *testing.T) {
	var config_map map[string]interface{}
	tables := []struct {
		db_service_models  map[string]map[string]schema.ModelHistory
		expect_api_payload api.PredictorPayload
	}{
		// case 1 测试传入为空的modelHistory,期望返回payload为空，而不会异常
		{
			// db data
			map[string]map[string]schema.ModelHistory{},
			api.PredictorPayload{
				[]api.PredictorService{},
			},
		},
		// case 2 测试正常的一个service,期望得到正确结构
		{
			// db data
			map[string]map[string]schema.ModelHistory{
				"predictor_service_a": map[string]schema.ModelHistory{
					"model_a": schema.ModelHistory{ModelName: "model_a", Timestamp: "ts_1", IsLocked: 0},
					"model_b": schema.ModelHistory{ModelName: "model_b", Timestamp: "ts_2", IsLocked: 0},
					"model_c": schema.ModelHistory{ModelName: "model_c", Timestamp: "ts_3", IsLocked: 0},
				},
			},
			api.PredictorPayload{
				[]api.PredictorService{
					{
						"predictor_service_a",
						16,
						config_map,
						[]api.PredictorModelRecord{
							{
								Name:       "model_a",
								Timestamp:  "ts_1",
								FullName:   "model_a-ts_1",
								ConfigName: "model_a-ts_1/model_a.json",
								IsLocked:   0,
								Md5:        "",
							},
							{
								Name:       "model_b",
								Timestamp:  "ts_2",
								FullName:   "model_b-ts_2",
								ConfigName: "model_b-ts_2/model_b.json",
								IsLocked:   0,
								Md5:        "",
							},
							{
								Name:       "model_c",
								Timestamp:  "ts_3",
								FullName:   "model_c-ts_3",
								ConfigName: "model_c-ts_3/model_c.json",
								IsLocked:   0,
								Md5:        "",
							},
						},
					},
				},
			},
		},
		// case 3 测试多个service,以及service 里面model_history 为空的极端情况
		{
			// db data
			map[string]map[string]schema.ModelHistory{
				"predictor_service_a": map[string]schema.ModelHistory{
					"model_a": schema.ModelHistory{ModelName: "model_a", Timestamp: "ts_a_1", IsLocked: 0},
					"model_b": schema.ModelHistory{ModelName: "model_b", Timestamp: "ts_b_1", IsLocked: 0},
					"model_c": schema.ModelHistory{ModelName: "model_c", Timestamp: "ts_c_1", IsLocked: 0},
				},
				"predictor_service_b": map[string]schema.ModelHistory{
					"model_d": schema.ModelHistory{ModelName: "model_d", Timestamp: "ts_d_1", IsLocked: 0},
					"model_e": schema.ModelHistory{ModelName: "model_e", Timestamp: "ts_e_1", IsLocked: 0},
					"model_f": schema.ModelHistory{ModelName: "model_f", Timestamp: "ts_f_1", IsLocked: 0},
				},
				"predictor_service_c": map[string]schema.ModelHistory{},
			},
			api.PredictorPayload{
				[]api.PredictorService{
					{
						"predictor_service_a",
						16,
						config_map,
						[]api.PredictorModelRecord{
							{
								Name:       "model_a",
								Timestamp:  "ts_a_1",
								FullName:   "model_a-ts_a_1",
								ConfigName: "model_a-ts_a_1/model_a.json",
								IsLocked:   0,
								Md5:        "",
							},
							{
								Name:       "model_b",
								Timestamp:  "ts_b_1",
								FullName:   "model_b-ts_b_1",
								ConfigName: "model_b-ts_b_1/model_b.json",
								IsLocked:   0,
								Md5:        "",
							},
							{
								Name:       "model_c",
								Timestamp:  "ts_c_1",
								FullName:   "model_c-ts_c_1",
								ConfigName: "model_c-ts_c_1/model_c.json",
								IsLocked:   0,
								Md5:        "",
							},
						},
					},
					{
						"predictor_service_b",
						8,
						config_map,
						[]api.PredictorModelRecord{
							{
								Name:       "model_d",
								Timestamp:  "ts_d_1",
								FullName:   "model_d-ts_d_1",
								ConfigName: "model_d-ts_d_1/model_d.json",
								IsLocked:   0,
								Md5:        "",
							},
							{
								Name:       "model_e",
								Timestamp:  "ts_e_1",
								FullName:   "model_e-ts_e_1",
								ConfigName: "model_e-ts_e_1/model_e.json",
								IsLocked:   0,
								Md5:        "",
							},
							{
								Name:       "model_f",
								Timestamp:  "ts_f_1",
								FullName:   "model_f-ts_f_1",
								ConfigName: "model_f-ts_f_1/model_f.json",
								IsLocked:   0,
								Md5:        "",
							},
						},
					},
					{
						"predictor_service_c",
						32,
						config_map,
						[]api.PredictorModelRecord{},
					},
				},
			},
		},
		// case 4 service_model中，service正常，但某些model_history不正常
		// 如ModelName或Timestamp为空，期望跳过invalid的model_history
		{
			// db data
			map[string]map[string]schema.ModelHistory{
				"predictor_service_a": map[string]schema.ModelHistory{
					"model_a":                         schema.ModelHistory{ModelName: "model_a", Timestamp: "ts_1", IsLocked: 0},
					"invalid_model_b_empty_name":      schema.ModelHistory{ModelName: "", Timestamp: "ts_2", IsLocked: 0},
					"invalid_model_c_empty_timestamp": schema.ModelHistory{ModelName: "invalid_model_c_empty_timestamp", Timestamp: "", IsLocked: 0},
				},
			},
			api.PredictorPayload{
				[]api.PredictorService{
					{
						"predictor_service_a",
						16,
						config_map,
						[]api.PredictorModelRecord{
							{
								Name:       "model_a",
								Timestamp:  "ts_1",
								FullName:   "model_a-ts_1",
								ConfigName: "model_a-ts_1/model_a.json",
								IsLocked:   0,
								Md5:        "",
							},
						},
					},
				},
			},
		},
	}
	confPtr := &conf.Conf{}
	envPtr := &env.Env{
		Conf: confPtr,
	}
	service_weight := map[string]int{"predictor_service_a": 16, "predictor_service_b": 8, "predictor_service_c": 32}
	service_config := make(map[string]string)
	for _, table := range tables {
		expectApiPayload := table.expect_api_payload
		apiPayload := getPayloadByServiceModels(envPtr, table.db_service_models, service_weight, service_config)
		common.SortPayload(&apiPayload)
		common.SortPayload(&expectApiPayload)
		// 比较请求payload 和 期望是否相等
		if !reflect.DeepEqual(apiPayload, expectApiPayload) {
			t.Errorf("TestGetPayloadByServiceModels DeepEqual fail, \n[Expect]:\n%+v, \n[Got]:\n%+v", common.Pretty(expectApiPayload), common.Pretty(apiPayload))
		}
	}
}

func TestGetServiceWeight(t *testing.T) {
	// create test conf, env, service
	conf := &conf.Conf{}
	conf.Db.Driver = "sqlite3"
	conf.Db.Name = "./model_logic_test.db"
	env := env.New(conf)

	// insert test data in test db
	db := env.Db
	mock.AutoMigrateAll(env)
	mock.CleanUp(env)

	db.Create(&schema.Host{Ip: "127.1.0.1"})  // Hid: 1
	db.Create(&schema.Host{Ip: "127.3.0.2"}) // Hid: 2
	db.Create(&schema.Host{Ip: "127.2.0.3"})  // Hid: 3
	db.Create(&schema.Host{Ip: "127.4.0.4"})  // Hid: 4
	db.Create(&schema.Host{Ip: "127.2.3.1"})  // Hid: 5
	db.Create(&schema.Host{Ip: "127.2.5.2"})  // Hid: 6
	db.Create(&schema.Host{Ip: "127.6.0.1"}) // Hid: 7
	db.Create(&schema.Host{Ip: "127.6.0.2"}) // Hid: 8

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.3.0.2 -> service_1", LoadWeight: 1})
	db.Create(&schema.HostService{Hid: 3, Sid: 1, Desc: "127.2.0.3 -> service_1", LoadWeight: 2})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2", LoadWeight: 3})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3", LoadWeight: 4})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.4.0.4 -> service_3", LoadWeight: 5})

	var ip string
	var expectedServiceWeight map[string]int
	expectedServiceWeight = make(map[string]int)
	// case 1
	{
		ip = "127.1.57.206"
		resServiceWeight, err := GetServiceWeight(env, ip)

		if err != nil || !reflect.DeepEqual(expectedServiceWeight, resServiceWeight) {
			t.Errorf("TestGetServiceWeight() failed: resServiceWeight:%v,expectedServiceWeight:%v,err:%v",
				resServiceWeight, expectedServiceWeight, err)
		}
	}
	// case 2
	{
		ip = "127.3.0.2"
		resServiceWeight, err := GetServiceWeight(env, ip)
		expectedServiceWeight = map[string]int{
			"service_1": 1,
		}
		if err != nil || !reflect.DeepEqual(expectedServiceWeight, resServiceWeight) {
			t.Errorf("TestGetServiceWeight() failed: resServiceWeight:%v,expectedServiceWeight:%v,err:%v",
				resServiceWeight, expectedServiceWeight, err)
		}
	}
	// case 3
	{
		ip = "127.2.0.3"
		resServiceWeight, err := GetServiceWeight(env, ip)
		expectedServiceWeight = map[string]int{
			"service_1": 2,
			"service_2": 3,
			"service_3": 4,
		}
		if err != nil || !reflect.DeepEqual(expectedServiceWeight, resServiceWeight) {
			t.Errorf("TestGetServiceWeight() failed: resServiceWeight:%v,expectedServiceWeight:%v,err:%v",
				resServiceWeight, expectedServiceWeight, err)
		}
	}
	db.Create(&schema.HostService{Hid: 2, Sid: 2, Desc: "127.3.0.2 -> service_2", LoadWeight: 6})

	// case 2
	{
		ip = "127.3.0.2"
		resServiceWeight, err := GetServiceWeight(env, ip)
		expectedServiceWeight = map[string]int{
			"service_1": 1,
			"service_2": 6,
		}
		if err != nil || !reflect.DeepEqual(expectedServiceWeight, resServiceWeight) {
			t.Errorf("TestGetServiceWeight() failed: resServiceWeight:%v,expectedServiceWeight:%v,err:%v",
				resServiceWeight, expectedServiceWeight, err)
		}
	}

	// clean up db
	mock.CleanUp(env)
}

func TestGetServiceConfig(t *testing.T) {
	// create test conf, env, service
	conf := &conf.Conf{}
	conf.Db.Driver = "sqlite3"
	conf.Db.Name = "./model_logic_test.db"
	env := env.New(conf)

	// insert test data in test db
	db := env.Db
	mock.AutoMigrateAll(env)
	mock.CleanUp(env)

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3

	db.Create(&schema.Config{Description: "config_1", Config: "{\"thread_num\": 1}"}) // cid: 1
	db.Create(&schema.Config{Description: "config_2", Config: "{\"thread_num\": 2}"}) // cid: 2

	db.Create(&schema.ServiceConfig{Description: "service_1_config", Sid: 1, Cid: 1})
	db.Create(&schema.ServiceConfig{Description: "service_2_config", Sid: 2, Cid: 2})

	// case 1
	{
		serviceNames := []string{"service_1"}
		resServiceConfig, err := GetServiceConfig(env, serviceNames)
		expectedServiceConfig := map[string]string{"service_1":"{\"thread_num\": 1}"}

		if err != nil || !reflect.DeepEqual(expectedServiceConfig, resServiceConfig) {
			t.Errorf("TestGetServiceConfig() failed: resServiceConfig:%v,expectedServiceConfig:%v,err:%v",
				resServiceConfig, expectedServiceConfig, err)
		}
	}
	
	// case 2
	{
		serviceNames := []string{"service_1", "service_2"}
		resServiceConfig, err := GetServiceConfig(env, serviceNames)
		expectedServiceConfig := map[string]string{
			"service_1":"{\"thread_num\": 1}", 
			"service_2":"{\"thread_num\": 2}",
		}

		if err != nil || !reflect.DeepEqual(expectedServiceConfig, resServiceConfig) {
			t.Errorf("TestGetServiceConfig() failed: resServiceConfig:%v,expectedServiceConfig:%v,err:%v",
				resServiceConfig, expectedServiceConfig, err)
		}
	}

	// case 3
	{
		serviceNames := []string{"service_3"}
		resServiceConfig, err := GetServiceConfig(env, serviceNames)
		expectedServiceConfig := map[string]string{}

		if err != nil || !reflect.DeepEqual(expectedServiceConfig, resServiceConfig) {
			t.Errorf("TestGetServiceConfig() failed: resServiceConfig:%v,expectedServiceConfig:%v,err:%v",
				resServiceConfig, expectedServiceConfig, err)
		}
	}

	// clean up db
	mock.CleanUp(env)
}

func TestFetchDbServices(t *testing.T) {
	// create test conf, env, service
	conf := &conf.Conf{}
	conf.Db.Driver = "sqlite3"
	conf.Db.Name = "./model_service_test.db"
	env := env.New(conf)

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

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.3.0.2 -> service_1"})
	db.Create(&schema.HostService{Hid: 3, Sid: 1, Desc: "127.2.0.3 -> service_1"})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2"})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3"})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.4.0.4 -> service_3"})

	// case 1.1: "127.1.0.1": no service on this host, expect empty result
	{
		env.LocalIp = "127.1.0.1"
		res, err := FetchDbServices(env)
		if err != nil || len(res) != 0 {
			mock.PrintModelsByServiceMap(res)
			t.Errorf("TestFetchDbServices() failed, got: %v, want: empty map\n", res)
		}
	}
	// case 1.2: "127.4.0.4" has "service_3"
	{
		env.LocalIp = "127.4.0.4"
		res, err := FetchDbServices(env)
		expected_res := []*schema.Service{
			&schema.Service{ID: 3, Name: "service_3"},
		}

		if err != nil || !reflect.DeepEqual(res, expected_res) {
			mock.PrintModelsByServiceMap(res)
			t.Errorf("TestFetchDbServices() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}
	// case 2: "127.3.0.2" has "service_1"
	{
		env.LocalIp = "127.3.0.2"
		res, err := FetchDbServices(env)
		expected_res := []*schema.Service{
			&schema.Service{ID: 1, Name: "service_1"},
		}

		if err != nil || !reflect.DeepEqual(res, expected_res) {
			mock.PrintModelsByServiceMap(res)
			t.Errorf("TestFetchDbServices() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}
	// case 3: "127.2.0.3"
	{
		env.LocalIp = "127.2.0.3"
		res, err := FetchDbServices(env)
		expected_res := []*schema.Service{
			&schema.Service{ID: 1, Name: "service_1"},
			&schema.Service{ID: 2, Name: "service_2"},
			&schema.Service{ID: 3, Name: "service_3"},
		}
		if err != nil || !reflect.DeepEqual(res, expected_res) {
			mock.PrintModelsByServiceMap(res)
			t.Errorf("TestFetchDbServices() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}
	// db updated
	db.Create(&schema.HostService{Hid: 1, Sid: 2, Desc: "127.1.0.1 -> service_2"})
	// case 4: "127.1.0.1" after db update
	{
		env.LocalIp = "127.1.0.1"
		res, err := FetchDbServices(env)
		expected_res := []*schema.Service{
			&schema.Service{ID: 2, Name: "service_2"},
		}

		if err != nil || !reflect.DeepEqual(res, expected_res) {
			mock.PrintModelsByServiceMap(res)
			t.Errorf("TestFetchDbServices() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}

	// clean up db
	mock.CleanUp(env)
}

func TestFetchDbHostsByService(t *testing.T) {
	// create test conf, env, service
	conf := &conf.Conf{}
	conf.Db.Driver = "sqlite3"
	conf.Db.Name = "./model_service_test.db"
	env := env.New(conf)

	// insert test data in test db
	db := env.Db
	mock.AutoMigrateAll(env)
	mock.CleanUp(env)

	db.Create(&schema.Host{Ip: "127.1.0.1"})  // Hid: 1
	db.Create(&schema.Host{Ip: "127.3.0.2"}) // Hid: 2
	db.Create(&schema.Host{Ip: "127.2.0.3"})  // Hid: 3
	db.Create(&schema.Host{Ip: "127.4.0.4"})  // Hid: 4
	db.Create(&schema.Host{Ip: "127.2.3.1"})  // Hid: 5

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.3.0.2 -> service_1"})
	db.Create(&schema.HostService{Hid: 3, Sid: 1, Desc: "127.2.0.3 -> service_1"})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2"})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3"})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.4.0.4 -> service_3"})

	var sid uint
	var prefixIp string
	// case 1
	{
		sid = 1
		prefixIp = "127.3."
		res, err := FetchDbHostsByService(env, sid, prefixIp)
		expected_res := []string{
			"127.3.0.2",
		}
		if err != nil || !reflect.DeepEqual(res, expected_res) {
			t.Errorf("TestFetchDbModelsByService() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}
	// case 2
	{
		sid = 1
		prefixIp = "127"
		res, err := FetchDbHostsByService(env, sid, prefixIp)
		expected_res := []string{
			"127.3.0.2",
			"127.2.0.3",
		}
		if err != nil || !reflect.DeepEqual(res, expected_res) {
			t.Errorf("TestFetchDbModelsByService() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}
	db.Create(&schema.HostService{Hid: 5, Sid: 1, Desc: "127.2.3.1 -> service_1"})
	// case 3
	{
		sid = 1
		prefixIp = "127.2."
		res, err := FetchDbHostsByService(env, sid, prefixIp)
		expected_res := []string{
			"127.2.0.3",
			"127.2.3.1",
		}
		if err != nil || !reflect.DeepEqual(res, expected_res) {
			t.Errorf("TestFetchDbModelsByService() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}
	// case 4
	{
		sid = 3
		prefixIp = ""
		res, err := FetchDbHostsByService(env, sid, prefixIp)
		expected_res := []string{
			"127.2.0.3",
			"127.4.0.4",
		}
		if err != nil || !reflect.DeepEqual(res, expected_res) {
			t.Errorf("TestFetchDbModelsByService() failed:\ngot:\n\t%v\nwant:\n\t%v\n", res, expected_res)
		}
	}
	// case 5
	{
		sid = 3
		prefixIp = "127.3"
		res, err := FetchDbHostsByService(env, sid, prefixIp)
		expected_res := []string{}
		if err != nil || !reflect.DeepEqual(res, expected_res) {
			t.Errorf("TestFetchDbModelsByService() failed:\ngot:\n\t%v\nwant:\n\t%v\n\terr:%v\n", res, expected_res, err)
		}
	}
	// clean up db
	mock.CleanUp(env)
}

func TestGetParentIP(t *testing.T) {
	// create test conf, env, service
	conf := &conf.Conf{}
	conf.Db.Driver = "sqlite3"
	conf.Db.Name = "./model_service_test.db"
	env := env.New(conf)

	// insert test data in test db
	db := env.Db
	mock.AutoMigrateAll(env)
	mock.CleanUp(env)

	db.Create(&schema.Host{Ip: "127.1.0.1"})  // Hid: 1
	db.Create(&schema.Host{Ip: "127.3.0.2"}) // Hid: 2
	db.Create(&schema.Host{Ip: "127.2.0.3"})  // Hid: 3
	db.Create(&schema.Host{Ip: "127.4.0.4"})  // Hid: 4
	db.Create(&schema.Host{Ip: "127.2.3.1"})  // Hid: 5
	db.Create(&schema.Host{Ip: "127.2.5.2"})  // Hid: 6
	db.Create(&schema.Host{Ip: "127.6.0.1"}) // Hid: 7
	db.Create(&schema.Host{Ip: "127.6.0.2"}) // Hid: 8

	db.Create(&schema.Service{Name: "service_1"}) // Sid: 1
	db.Create(&schema.Service{Name: "service_2"}) // Sid: 2
	db.Create(&schema.Service{Name: "service_3"}) // Sid: 3

	db.Create(&schema.HostService{Hid: 2, Sid: 1, Desc: "127.3.0.2 -> service_1"})
	db.Create(&schema.HostService{Hid: 3, Sid: 1, Desc: "127.2.0.3 -> service_1"})
	db.Create(&schema.HostService{Hid: 3, Sid: 2, Desc: "127.2.0.3 -> service_2"})
	db.Create(&schema.HostService{Hid: 3, Sid: 3, Desc: "127.2.0.3 -> service_3"})
	db.Create(&schema.HostService{Hid: 4, Sid: 3, Desc: "127.4.0.4 -> service_3"})

	env.Conf.P2PModelService.SrcHost = "127.2.12.125"
	env.Conf.ValidateService.Host = "127.1.57.206"
	env.Conf.P2PModelService.PeerLimit = 5
	var expectedIp string
	var expectedPeerNum int
	var dbService *schema.Service
	var ServiceCheckTimesMap sync.Map
	// case 1
	{
		env.LocalIp = "127.1.57.206"
		dbService = &schema.Service{
			ID:   1,
			Name: "service_1",
		}
		
		resIp, resPeerNum, err := GetParentIP(env, dbService, ServiceCheckTimesMap, env.Conf.P2PModelService.ServicePullAlertLimit, env.Conf.P2PModelService.SrcHost)
		expectedIp = "127.2.12.125"
		expectedPeerNum = 0
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIP() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	// case 2
	{
		env.LocalIp = "127.1.57.207"
		env.Conf.P2PModelService.ServicePullAlertLimit = 5
		dbService = &schema.Service{
			ID:   1,
			Name: "service_1",
		}
		ServiceCheckTimesMap.Store(dbService.Name, int32(6))

		resIp, resPeerNum, err := GetParentIP(env, dbService, ServiceCheckTimesMap, env.Conf.P2PModelService.ServicePullAlertLimit, env.Conf.P2PModelService.SrcHost)
		expectedIp = "127.2.12.125"
		expectedPeerNum = 0
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIP() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}

	// case 3
	{
		env.LocalIp = "127.1.57.207"
		env.Conf.P2PModelService.ServicePullAlertLimit = 5
		dbService = &schema.Service{
			ID:   1,
			Name: "service_1",
		}
		ServiceCheckTimesMap.Store(dbService.Name, int32(3))

		resIp, resPeerNum, err := GetParentIP(env, dbService, ServiceCheckTimesMap, env.Conf.P2PModelService.ServicePullAlertLimit, env.Conf.P2PModelService.SrcHost)
		expectedIp = ""
		expectedPeerNum = 0
		if err == nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIP() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	// case 4
	{
		env.LocalIp = "127.2.0.3"
		env.Conf.P2PModelService.ServicePullAlertLimit = 5
		dbService = &schema.Service{
			ID:   1,
			Name: "service_1",
		}
		ServiceCheckTimesMap.Store(dbService.Name, int32(3))
		resIp, resPeerNum, err := GetParentIP(env, dbService, ServiceCheckTimesMap, env.Conf.P2PModelService.ServicePullAlertLimit, env.Conf.P2PModelService.SrcHost)
		expectedIp = "127.2.12.125"
		expectedPeerNum = 0
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIP() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	db.Create(&schema.HostService{Hid: 5, Sid: 1, Desc: "127.2.3.1 -> service_1"})
	db.Create(&schema.HostService{Hid: 6, Sid: 1, Desc: "127.2.5.2 -> service_1"})
	db.Create(&schema.HostService{Hid: 7, Sid: 1, Desc: "127.6.0.1 -> service_1"})
	db.Create(&schema.HostService{Hid: 8, Sid: 1, Desc: "127.6.0.2 -> service_1"})
	{
		env.LocalIp = "127.2.3.1"
		env.Conf.P2PModelService.ServicePullAlertLimit = 5
		dbService = &schema.Service{
			ID:   1,
			Name: "service_1",
		}
		ServiceCheckTimesMap.Store(dbService.Name, int32(3))
		resIp, resPeerNum, err := GetParentIP(env, dbService, ServiceCheckTimesMap, env.Conf.P2PModelService.ServicePullAlertLimit, env.Conf.P2PModelService.SrcHost)
		expectedIp = "127.2.0.3"
		expectedPeerNum = 2
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIP() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	{
		env.LocalIp = "127.2.0.3"
		env.Conf.P2PModelService.ServicePullAlertLimit = 5
		dbService = &schema.Service{
			ID:   1,
			Name: "service_1",
		}
		ServiceCheckTimesMap.Store(dbService.Name, int32(3))
		resIp, resPeerNum, err := GetParentIP(env, dbService, ServiceCheckTimesMap, env.Conf.P2PModelService.ServicePullAlertLimit, env.Conf.P2PModelService.SrcHost)
		expectedIp = "127.2.12.125"
		expectedPeerNum = 0
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIP() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	{
		env.LocalIp = "127.2.5.2"
		env.Conf.P2PModelService.ServicePullAlertLimit = 5
		dbService = &schema.Service{
			ID:   1,
			Name: "service_1",
		}
		ServiceCheckTimesMap.Store(dbService.Name, int32(5))
		resIp, resPeerNum, err := GetParentIP(env, dbService, ServiceCheckTimesMap, env.Conf.P2PModelService.ServicePullAlertLimit, env.Conf.P2PModelService.SrcHost)
		expectedIp = "127.2.12.125"
		expectedPeerNum = 0
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIP() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	{
		env.LocalIp = "127.6.0.2"
		env.Conf.P2PModelService.ServicePullAlertLimit = 5
		dbService = &schema.Service{
			ID:   1,
			Name: "service_1",
		}
		ServiceCheckTimesMap.Store(dbService.Name, int32(1))
		resIp, resPeerNum, err := GetParentIP(env, dbService, ServiceCheckTimesMap, env.Conf.P2PModelService.ServicePullAlertLimit, env.Conf.P2PModelService.SrcHost)
		expectedIp = "127.6.0.1"
		expectedPeerNum = 1
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIP() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	{
		env.LocalIp = "127.6.0.1"
		env.Conf.P2PModelService.ServicePullAlertLimit = 5
		dbService = &schema.Service{
			ID:   1,
			Name: "service_1",
		}
		ServiceCheckTimesMap.Store(dbService.Name, int32(0))
		resIp, resPeerNum, err := GetParentIP(env, dbService, ServiceCheckTimesMap, env.Conf.P2PModelService.ServicePullAlertLimit, env.Conf.P2PModelService.SrcHost)
		expectedIp = "127.2.12.125"
		expectedPeerNum = 0
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIP() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}

	// clean up db
	mock.CleanUp(env)
}

func TestGetParentIpFromABIps(t *testing.T) {
	confPtr := &conf.Conf{}
	env := &env.Env{
		Conf: confPtr,
	}
	env.Conf.P2PModelService.SrcHost = "127.2.12.125"
	env.Conf.P2PModelService.PeerLimit = 5
	var expectedIp string
	var expectedPeerNum int
	// case 1
	{
		env.LocalIp = "127.6.0.1"
		abIps := []string{
			"127.6.0.1",
			"127.6.0.2",
			"127.6.0.3",
		}
		resIp, resPeerNum, err := GetParentIpFromABIps(env, abIps)
		expectedIp = "127.2.12.125"
		expectedPeerNum = 0
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIpFromABIps() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	// case 2
	{
		env.LocalIp = "127.6.0.2"
		abIps := []string{
			"127.6.0.1",
			"127.6.0.2",
			"127.6.0.3",
		}
		resIp, resPeerNum, err := GetParentIpFromABIps(env, abIps)
		expectedIp = "127.6.0.1"
		expectedPeerNum = 2
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIpFromABIps() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	// case 3
	{
		env.LocalIp = "127.6.0.3"
		abIps := []string{}
		resIp, resPeerNum, err := GetParentIpFromABIps(env, abIps)
		expectedIp = ""
		expectedPeerNum = 0
		if err == nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIpFromABIps() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	// case 4
	{
		env.LocalIp = "127.6.0.2"
		abIps := []string{
			"127.6.0.1",
			"127.6.0.2",
			"127.6.0.3",
			"127.6.0.4",
			"127.6.0.5",
			"127.6.0.6",
		}
		resIp, resPeerNum, err := GetParentIpFromABIps(env, abIps)
		expectedIp = "127.6.0.1"
		expectedPeerNum = 5
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIpFromABIps() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	// case 5
	{
		env.LocalIp = "127.6.0.2"
		abIps := []string{
			"127.6.0.1",
			"127.6.0.2",
			"127.6.0.3",
			"127.6.0.4",
			"127.6.0.5",
			"127.6.0.6",
			"127.6.0.7",
		}
		resIp, resPeerNum, err := GetParentIpFromABIps(env, abIps)
		expectedIp = "127.2.12.125"
		expectedPeerNum = 0
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIpFromABIps() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}
	// case 6
	{
		abIps := []string{
			"127.6.0.1",
			"127.6.0.2",
			"127.6.0.3",
			"127.6.0.4",
			"127.6.0.5",
			"127.6.0.6",
			"127.6.0.7",
		}
		env.LocalIp = "127.6.0.6"
		resIp, resPeerNum, err := GetParentIpFromABIps(env, abIps)
		expectedIp = "127.6.0.2"
		expectedPeerNum = 2
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIpFromABIps() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}

		env.LocalIp = "127.6.0.5"
		resIp, resPeerNum, err = GetParentIpFromABIps(env, abIps)
		expectedIp = "127.6.0.1"
		expectedPeerNum = 3
		if err != nil || expectedIp != resIp || expectedPeerNum != resPeerNum {
			t.Errorf("TestGetParentIpFromABIps() failed: resIp:%s,resPeerNum:%d,err:%v,expectedIp:%s,expectedPeerNum:%d",
				resIp, resPeerNum, err, expectedIp, expectedPeerNum)
		}
	}

}