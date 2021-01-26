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

func TestGetMailModelsMap(t *testing.T) {
	webHook := NewWebHook(&util.WebHookRequest{}, "")
	tables := []struct {
		modelMailsMap map[string][]string
		mailModelsMap map[string][]string
	}{
		// case 1
		{
			map[string][]string{
				"model_v1": []string{
					"lilei@github.com",
					"hanmeimei@github.com",
				},
				"model_v2": []string{
					"lilei@github.com",
					"hanmeimei@github.com",
				},
			},
			map[string][]string{
				"lilei@github.com": []string{
					"model_v1",
					"model_v2",
				},
				"hanmeimei@github.com": []string{
					"model_v1",
					"model_v2",
				},
			},
		},
		// case 2
		{
			map[string][]string{
				"model_v1": []string{
					"kate@github.com",
					"hanmeimei@github.com",
				},
				"model_v2": []string{
					"lilei@github.com",
					"hanmeimei@github.com",
				},
			},
			map[string][]string{
				"kate@github.com": []string{
					"model_v1",
				},
				"lilei@github.com": []string{
					"model_v2",
				},
				"hanmeimei@github.com": []string{
					"model_v1",
					"model_v2",
				},
			},
		},
		// case 3
		{
			map[string][]string{
				"model_v1": []string{
					"kate@github.com",
					"hanmeimei@github.com",
				},
				"model_v2": []string{},
				"model_v3": []string{
					"poli@github.com",
					"hanmeimei@github.com",
				},
			},
			map[string][]string{
				"kate@github.com": []string{
					"model_v1",
				},
				"poli@github.com": []string{
					"model_v3",
				},
				"hanmeimei@github.com": []string{
					"model_v1",
					"model_v3",
				},
				"all": []string{
					"model_v2",
				},
			},
		},
		// case 4
		{
			map[string][]string{},
			map[string][]string{},
		},
		// case 5
		{
			map[string][]string{
				"model_v1": []string{},
				"model_v2": []string{},
				"model_v3": []string{},
			},
			map[string][]string{
				"all": []string{
					"model_v1",
					"model_v2",
					"model_v3",
				},
			},
		},
	}

	for _, table := range tables {
		mailModelsMap := webHook.getMailModelsMap(table.modelMailsMap)

		// 比较请求payload 和 期望是否相等
		if !mock.IsEqualSliceMap(mailModelsMap, table.mailModelsMap) {
			t.Errorf("TestGetMailModelsMap DeepEqual fail, \n[Expect]:\n%+v, \n[Got]:\n%+v", table.mailModelsMap, mailModelsMap)
		}
	}
}

func TestGetModelMailsMap(t *testing.T) {
	webHook := NewWebHook(&util.WebHookRequest{}, "")
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./webhook_logic_test.db"
	db := env.InitMysql(conf)
	dao.SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)
	db.Create(&schema.Model{
		Name:      "model_v1",
		Extension: `{"MailRecipients":["lili@github.com","haha@github.com"]}`,
	})
	db.Create(&schema.Model{
		Name:      "model_v2",
		Extension: "",
	})
	db.Create(&schema.Model{
		Name:      "model_v3",
		Extension: `{"MailRecipients":["hanmeimei@github.com","lili@github.com"]}`,
	})
	db.Create(&schema.Model{
		Name:      "model_v4",
		Extension: `{"MailRecipients":[]}`,
	})

	// case 1
	{
		alertModelList := []string{
			"model_v1",
			"model_v2",
			"model_v3",
			"model_v4",
		}
		expectedModelMailsMap := map[string][]string{
			"model_v1": []string{
				"lili@github.com",
				"haha@github.com",
			},
			"model_v2": []string{},
			"model_v3": []string{
				"hanmeimei@github.com",
				"lili@github.com",
			},
			"model_v4": []string{},
		}
		modelMailsMap, err := webHook.getModelMailsMap(alertModelList)

		if err != nil || !mock.IsEqualSliceMap(expectedModelMailsMap, modelMailsMap) {
			t.Errorf("TestGetModelMailsMap() failed,err:%v \n[Got]:\n%v,\n[Expect]:\n%v",
				err, modelMailsMap, expectedModelMailsMap)
		}
	}
	// case 2
	{
		alertModelList := []string{
			"model_v1",
			"model_v4",
			"tf_xfea_estimator_v5_cpc",
		}
		expectedModelMailsMap := map[string][]string{
			"model_v1": []string{
				"lili@github.com",
				"haha@github.com",
			},
			"model_v4": []string{},
		}
		modelMailsMap, err := webHook.getModelMailsMap(alertModelList)

		if err != nil || !mock.IsEqualSliceMap(expectedModelMailsMap, modelMailsMap) {
			t.Errorf("TestGetModelMailsMap() failed,err:%v \n[Got]:\n%v,\n[Expect]:\n:%v",
				err, modelMailsMap, expectedModelMailsMap)
		}
	}
	// case 3
	{
		alertModelList := []string{}
		expectedModelMailsMap := map[string][]string{}
		modelMailsMap, err := webHook.getModelMailsMap(alertModelList)

		if err == nil || !reflect.DeepEqual(expectedModelMailsMap, modelMailsMap) {
			t.Errorf("TestGetModelMailsMap() failed,err:%v \n[Got]:\n%v,\n[Expect]:\n:%v",
				err, modelMailsMap, expectedModelMailsMap)
		}
	}

	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}

func TestParseModelMatches(t *testing.T) {
	tables := []struct {
		reqEvalMatches []util.EvalMatch
		evalMatchMap   map[string]util.EvalMatch
		modelList      []string
	}{
		// case 1
		{
			[]util.EvalMatch{
				util.EvalMatch{
					Value:  "16",
					Metric: "model_avg_ctr",
					Tags: map[string]interface{}{
						"category": "tf_model_v1",
					},
				},
				util.EvalMatch{
					Value:  "12",
					Metric: "model_avg_ctr",
					Tags: map[string]interface{}{
						"category": "tf_model_v2",
					},
				},
			},
			map[string]util.EvalMatch{
				"tf_model_v1": util.EvalMatch{
					Value:  "16",
					Metric: "model_avg_ctr",
					Tags: map[string]interface{}{
						"category": "tf_model_v1",
					},
				},
				"tf_model_v2": util.EvalMatch{
					Value:  "12",
					Metric: "model_avg_ctr",
					Tags: map[string]interface{}{
						"category": "tf_model_v2",
					},
				},
			},
			[]string{
				"tf_model_v1",
				"tf_model_v2",
			},
		},
		// case 2
		{
			[]util.EvalMatch{
				util.EvalMatch{
					Value:  "16",
					Metric: "model_avg_ctr",
					Tags: map[string]interface{}{
						"model_name": "tf_model_v1",
					},
				},
				util.EvalMatch{
					Value:  "12",
					Metric: "model_avg_ctr",
					Tags: map[string]interface{}{
						"model_name": "tf_model_v2",
					},
				},
			},
			map[string]util.EvalMatch{
				"tf_model_v1": util.EvalMatch{
					Value:  "16",
					Metric: "model_avg_ctr",
					Tags: map[string]interface{}{
						"model_name": "tf_model_v1",
					},
				},
				"tf_model_v2": util.EvalMatch{
					Value:  "12",
					Metric: "model_avg_ctr",
					Tags: map[string]interface{}{
						"model_name": "tf_model_v2",
					},
				},
			},
			[]string{
				"tf_model_v1",
				"tf_model_v2",
			},
		},
		// case 3
		{
			[]util.EvalMatch{
				util.EvalMatch{
					Value:  "16",
					Metric: "model_avg_ctr",
					Tags: map[string]interface{}{
						"model": "tf_model_v1",
					},
				},
				util.EvalMatch{
					Value:  "12",
					Metric: "model_avg_ctr",
					Tags: map[string]interface{}{
						"model": "tf_model_v2",
					},
				},
			},
			map[string]util.EvalMatch{},
			[]string{},
		},
	}

	for _, table := range tables {
		webHook := NewWebHook(&util.WebHookRequest{}, "")
		webHook.ReqData.EvalMatches = table.reqEvalMatches
		evailMatches, modelList := webHook.parseModelMatches()

		// 比较请求payload 和 期望是否相等
		if !reflect.DeepEqual(evailMatches, table.evalMatchMap) || !reflect.DeepEqual(modelList, table.modelList) {
			t.Errorf("TestParseModelMatches DeepEqual fail, \n[Expect]:\n%+v\n%+v, \n[Got]:\n%+v\n%+v",
				table.evalMatchMap, table.modelList, evailMatches, modelList)
		}
	}
}
