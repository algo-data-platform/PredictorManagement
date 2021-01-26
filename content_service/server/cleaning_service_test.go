package server

import (
	"content_service/conf"
	"content_service/env"
	"content_service/mock"
	"content_service/schema"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestGetDiskModelsByFiles(t *testing.T) {
	// init test case
	verifyStructs := []struct {
		files              []string
		expectDiskModelMap map[string]DiskModel
	}{
		// case 1 正常测试结构转化
		{
			files: []string{
				"lr_model_1-20191217_144035",
				"lr_model_3-20191217_153345",
				"lr_model_v0_2-20191205_135500",
				"lr_model_v1_2-20191205_140000",
				"lr_model_v1_2-20191210_115019",
				"tf_model_1-20191203_091907",
				"tf_model_1-20191204_140757",
				"tf_model_1-20191211_192507",
			},
			expectDiskModelMap: map[string]DiskModel{
				"lr_model_1": DiskModel{
					ModelName: "lr_model_1",
					Timestamps: []string{
						"20191217_144035",
					},
				},
				"lr_model_3": DiskModel{
					ModelName: "lr_model_3",
					Timestamps: []string{
						"20191217_153345",
					},
				},
				"lr_model_v0_2": DiskModel{
					ModelName: "lr_model_v0_2",
					Timestamps: []string{
						"20191205_135500",
					},
				},
				"lr_model_v1_2": DiskModel{
					ModelName: "lr_model_v1_2",
					Timestamps: []string{
						"20191205_140000",
						"20191210_115019",
					},
				},
				"tf_model_1": DiskModel{
					ModelName: "tf_model_1",
					Timestamps: []string{
						"20191203_091907",
						"20191204_140757",
						"20191211_192507",
					},
				},
			},
		},
		// case 2 如果文件名不以-分隔，将不会在期望列表中
		{
			files: []string{
				"lr_model_1",
				"lr_model_3_20191217_153345",
				"lr_model_v0_2-20191205_135500",
				"lr_model_v1_2-20191205_140000",
				"lr_model_v1_2-20191210_115019",
				"tf_model_1-20191203_091907",
				"tf_model_1_20191204_140757",
				"tf_model_1-20191211_192507",
			},
			expectDiskModelMap: map[string]DiskModel{
				"lr_model_v0_2": DiskModel{
					ModelName: "lr_model_v0_2",
					Timestamps: []string{
						"20191205_135500",
					},
				},
				"lr_model_v1_2": DiskModel{
					ModelName: "lr_model_v1_2",
					Timestamps: []string{
						"20191205_140000",
						"20191210_115019",
					},
				},
				"tf_model_1": DiskModel{
					ModelName: "tf_model_1",
					Timestamps: []string{
						"20191203_091907",
						"20191211_192507",
					},
				},
			},
		},
	}

	cleanService := NewCleaningService()
	for _, row := range verifyStructs {
		var diskModelMap = map[string]DiskModel{}
		cleanService.getDiskModelsByFiles(diskModelMap, row.files)
		if !reflect.DeepEqual(diskModelMap, row.expectDiskModelMap) {
			t.Errorf("TestFetchDiskData DeepEqual fail,diskModelMap:%+v,expectDiskModelMap:%+v\n",
				diskModelMap, row.expectDiskModelMap)
		}
	}
}

func TestGetFilesToCleanByVersion(t *testing.T) {
	// create test conf, env, service
	conf := &conf.Conf{}
	conf.Db.Driver = "sqlite3"
	conf.Db.Name = "./cleaning_service_test.db"
	env := env.New(conf)

	// insert test data in test db
	db := env.Db
	mock.AutoMigrateAll(env)
	mock.CleanUp(env)

	db.Create(&schema.ModelHistory{ModelName: "tf_model_1", Timestamp: "20191203_091907", Desc: "Validated"}) // ID: 1
	db.Create(&schema.ModelHistory{ModelName: "tf_model_1", Timestamp: "20191204_140757", Desc: "Validated"}) // ID: 2
	db.Create(&schema.ModelHistory{ModelName: "tf_model_1", Timestamp: "20191211_192507", Desc: "Abandoned"}) // ID: 3
	db.Create(&schema.ModelHistory{ModelName: "tf_model_1", Timestamp: "20200111_192507", Desc: "Abandoned"}) // ID: 4
	db.Create(&schema.ModelHistory{ModelName: "tf_model_1", Timestamp: "20200110_192507", Desc: "Abandoned"}) // ID: 5
	db.Create(&schema.ModelHistory{ModelName: "tf_model_1", Timestamp: "20200113_000000", Desc: "Abandoned"}) // ID: 6

	db.Create(&schema.ModelHistory{ModelName: "lr_model_2", Timestamp: "20191203_091907", Desc: "Validated"}) // ID: 1
	db.Create(&schema.ModelHistory{ModelName: "lr_model_2", Timestamp: "20191204_140757", Desc: "Validated"}) // ID: 2
	db.Create(&schema.ModelHistory{ModelName: "lr_model_2", Timestamp: "20191211_192507", Desc: "Validated"}) // ID: 3

	verifyStructs := []struct {
		model_name         string
		disk_model         DiskModel
		versionsToKeep     int
		local_ip           string
		base_path          string
		expectFilesToClean []string
	}{
		// case 1 保留版本为1，只有一个版本，期待清除列表为空
		{
			"lr_model_1",
			DiskModel{
				ModelName: "lr_model_1",
				Timestamps: []string{
					"20191217_144035",
				},
			},
			1,
			"127.1.57.206",
			"/tmp",
			[]string{},
		},
		// case 2 保留版本为2，有两个版本，期待清除列表为空
		{
			"lr_model_2",
			DiskModel{
				ModelName: "lr_model_2",
				Timestamps: []string{
					"20191205_140000",
					"20191210_115019",
				},
			},
			2,
			"127.1.57.206",
			"/tmp",
			[]string{},
		},
		// case 3 保留版本为2，有三个版本，期待清除列表为1
		{
			"tf_model_1",
			DiskModel{
				ModelName: "tf_model_1",
				Timestamps: []string{
					"20191203_091907",
					"20191204_140757",
					"20191211_192507",
				},
			},
			2,
			"127.1.57.206",
			"/tmp",
			[]string{
				"/tmp/tf_model_1-20191203_091907",
			},
		},
		// case 4 保留版本为3，有>3个版本，期待清除多余且老版本
		{
			"tf_model_1",
			DiskModel{
				ModelName: "tf_model_1",
				Timestamps: []string{
					"20191203_091907",
					"20191204_140757",
					"20191211_192507",
					"20200111_192507",
					"20200110_192507",
					"20200113_000000",
				},
			},
			3,
			"127.1.57.206",
			"/tmp",
			[]string{
				"/tmp/tf_model_1-20191211_192507",
				"/tmp/tf_model_1-20191204_140757",
				"/tmp/tf_model_1-20191203_091907",
			},
		},
		// case 5 边界情况，保留版本为0，有多个版本，会保留一个版本
		{
			"tf_model_1",
			DiskModel{
				ModelName: "tf_model_1",
				Timestamps: []string{
					"20191203_091907",
					"20191204_140757",
					"20191211_192507",
					"20200111_192507",
					"20200110_192507",
					"20200113_000000",
				},
			},
			0,
			"127.1.57.206",
			"/tmp",
			[]string{
				"/tmp/tf_model_1-20191211_192507",
				"/tmp/tf_model_1-20200111_192507",
				"/tmp/tf_model_1-20200110_192507",
				"/tmp/tf_model_1-20191204_140757",
				"/tmp/tf_model_1-20191203_091907",
			},
		},
		// case 6 中转机
		{
			"tf_model_1",
			DiskModel{
				ModelName: "tf_model_1",
				Timestamps: []string{
					"20191203_091907",
					"20191204_140757",
					"20191211_192507",
					"20200111_192507",
					"20200110_192507",
					"20200113_000000",
				},
			},
			3,
			"127.2.12.125",
			"/tmp",
			[]string{
				"/tmp/tf_model_1-20191211_192507",
				"/tmp/tf_model_1-20191203_091907",
			},
		},
		// case 7 非中转机
		{
			"tf_model_1",
			DiskModel{
				ModelName: "tf_model_1",
				Timestamps: []string{
					"20191203_091907",
					"20191204_140757",
					"20191211_192507",
					"20200111_192507",
					"20200110_192507",
					"20200113_000000",
				},
			},
			3,
			"10.93.192.201",
			"/tmp",
			[]string{
				"/tmp/tf_model_1-20191211_192507",
				"/tmp/tf_model_1-20191204_140757",
				"/tmp/tf_model_1-20191203_091907",
			},
		},
		// case 7 中转机
		{
			"lr_model_2",
			DiskModel{
				ModelName: "lr_model_2",
				Timestamps: []string{
					"20191203_091907",
					"20191204_140757",
					"20191211_192507",
					"20200111_192507",
					"20200110_192507",
					"20200113_000000",
				},
			},
			3,
			"127.2.12.125",
			"/tmp",
			[]string{
				"/tmp/lr_model_2-20191204_140757",
				"/tmp/lr_model_2-20191203_091907",
			},
		},
	}

	cleanService := NewCleaningService()
	cleanService.TransmitIP = "127.2.12.125"
	for index, row := range verifyStructs {
		env.LocalIp = row.local_ip
		filesToClean, err := cleanService.getFilesToCleanByVersion(env, row.model_name, row.disk_model, row.versionsToKeep, row.base_path)
		if err != nil {
			t.Errorf("TestCleanByVersion fail, case: %d, err: %v\n", index+1, err)
		}
		sort.Strings(filesToClean)
		sort.Strings(row.expectFilesToClean)
		if !reflect.DeepEqual(filesToClean, row.expectFilesToClean) {
			t.Errorf("TestCleanByVersion DeepEqual fail,case: %d,filesToClean:%+v,expectFilesToClean:%+v\n",
				index+1, filesToClean, row.expectFilesToClean)
		}
	}
	// clean up db
	mock.CleanUp(env)
}


func TestGetFilesToCleanByTime(t *testing.T) {

	lastDayVersion := time.Unix(time.Now().Unix() - 86400, 0).Format("20060102_150405")
	last2DayVersion := time.Unix(time.Now().Unix() - 86400*2, 0).Format("20060102_150405")
	nowVersion := time.Now().Format("20060102_150405")
	tomorrowVersion := time.Unix(time.Now().Unix() + 86400, 0).Format("20060102_150405")
	verifyStructs := []struct {
		model_name         string
		disk_model         DiskModel
		hoursToKeep       int
		base_path          string
		expectFilesToClean []string
	}{
		// case 1 
		{
			"lr_model_1",
			DiskModel{
				ModelName: "lr_model_1",
				Timestamps: []string{
					"20191217_144035",
				},
			},
			0,
			"/tmp",
			[]string{
				"/tmp/lr_model_1-20191217_144035",
			},
		},
		// case 2 
		{
			"lr_model_2",
			DiskModel{
				ModelName: "lr_model_2",
				Timestamps: []string{
					"20191205_140000",
					"20191210_115019",
				},
			},
			24,
			"/tmp",
			[]string{
				"/tmp/lr_model_2-20191205_140000",
				"/tmp/lr_model_2-20191210_115019",
			},
		},
		// case 3
		{
			"lr_model_2",
			DiskModel{
				ModelName: "lr_model_2",
				Timestamps: []string{
					"20191205_140000",
					lastDayVersion,
				},
			},
			23,
			"/tmp",
			[]string{
				"/tmp/lr_model_2-20191205_140000",
				"/tmp/lr_model_2-"+lastDayVersion,
			},
		},
		// case 4
		{
			"lr_model_2",
			DiskModel{
				ModelName: "lr_model_2",
				Timestamps: []string{
					"20191205_140000",
					lastDayVersion,
					last2DayVersion,
					nowVersion,
					tomorrowVersion,

				},
			},
			25,
			"/tmp",
			[]string{
				"/tmp/lr_model_2-20191205_140000",
				"/tmp/lr_model_2-"+last2DayVersion,
			},
		},
		// case 5
		{
			"lr_model_2",
			DiskModel{
				ModelName: "lr_model_2",
				Timestamps: []string{
					"20191205_140000",
					lastDayVersion,
					last2DayVersion,
					nowVersion,
					tomorrowVersion,

				},
			},
			49,
			"/tmp",
			[]string{
				"/tmp/lr_model_2-20191205_140000",
			},
		},
		// case 6
		{
			"lr_model_2",
			DiskModel{
				ModelName: "lr_model_2",
				Timestamps: []string{
					"20191205_140000",
					lastDayVersion,
					last2DayVersion,
					nowVersion,
					tomorrowVersion,

				},
			},
			0,
			"/tmp",
			[]string{
				"/tmp/lr_model_2-20191205_140000",
				"/tmp/lr_model_2-"+lastDayVersion,
				"/tmp/lr_model_2-"+last2DayVersion,
				"/tmp/lr_model_2-"+nowVersion,
			},
		},
	}

	cleanService := NewCleaningService()
	for index, row := range verifyStructs {
		filesToClean, err := cleanService.getFilesToCleanByTime(row.model_name, row.disk_model, row.hoursToKeep, row.base_path)
		if err != nil {
			t.Errorf("TestCleanByVersion fail, case: %d, err: %v\n", index+1, err)
		}
		sort.Strings(filesToClean)
		sort.Strings(row.expectFilesToClean)
		if !reflect.DeepEqual(filesToClean, row.expectFilesToClean) {
			t.Errorf("TestCleanByVersion DeepEqual fail,case: %d,filesToClean:%+v,expectFilesToClean:%+v\n",
				index+1, filesToClean, row.expectFilesToClean)
		}
	}

}
