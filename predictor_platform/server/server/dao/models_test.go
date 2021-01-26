package dao

import (
	"reflect"
	"server/conf"
	"server/env"
	"server/mock"
	"server/schema"
	"testing"
)

func TestGetModelExtensionMap(t *testing.T) {
	conf := &conf.Conf{}
	conf.MysqlDb.Driver = "sqlite3"
	conf.MysqlDb.Database = "./models_test.db"
	db := env.InitMysql(conf)
	SetMysqlDB(db)
	mock.AutoMigrateAll(db, conf.MysqlDb.Driver)
	mock.CleanUp(db, conf.MysqlDb.Driver)
	db.Create(&schema.Model{
		Name:      "model_v1",
		Extension: `{"MailRecipients":["lili@gmail.com","haha@gmail.com"]}`,
	})
	db.Create(&schema.Model{
		Name:      "model_v2",
		Extension: "",
	})
	db.Create(&schema.Model{
		Name:      "model_v3",
		Extension: `{"MailRecipients":["hanmeimei@gmail.com","lili@gmail.com"]}`,
	})

	// case 1
	{
		modelList := []string{
			"model_v1",
			"model_v2",
			"model_v3",
		}
		expectedModelExtensionsMap := map[string]string{
			"model_v1": `{"MailRecipients":["lili@gmail.com","haha@gmail.com"]}`,
			"model_v2": ``,
			"model_v3": `{"MailRecipients":["hanmeimei@gmail.com","lili@gmail.com"]}`,
		}
		modelExtensionsMap, err := GetModelExtensionMap(modelList)

		if err != nil || !reflect.DeepEqual(expectedModelExtensionsMap, modelExtensionsMap) {
			t.Errorf("TestGetModelExtensionsMap() failed,err:%v \n[Got]:\n%v,\n[Expect]:\n%v",
				err, modelExtensionsMap, expectedModelExtensionsMap)
		}
	}
	// case 2
	{
		modelList := []string{
			"model_v1",
			"model_v4",
			"model_v5",
		}
		expectedModelExtensionsMap := map[string]string{
			"model_v1": `{"MailRecipients":["lili@gmail.com","haha@gmail.com"]}`,
		}
		modelExtensionsMap, err := GetModelExtensionMap(modelList)

		if err != nil || !reflect.DeepEqual(expectedModelExtensionsMap, modelExtensionsMap) {
			t.Errorf("TestGetModelExtensionsMap() failed,err:%v \n[Got]:\n%v,\n[Expect]:\n%v",
				err, modelExtensionsMap, expectedModelExtensionsMap)
		}
	}

	// case 3
	{
		modelList := []string{}
		expectedModelExtensionsMap := map[string]string{}
		modelExtensionsMap, err := GetModelExtensionMap(modelList)

		if err == nil || !reflect.DeepEqual(expectedModelExtensionsMap, modelExtensionsMap) {
			t.Errorf("TestGetModelExtensionsMap() failed,err:%v \n[Got]:\n%v,\n[Expect]:\n%v",
				err, modelExtensionsMap, expectedModelExtensionsMap)
		}
	}

	// clean up db
	mock.CleanUp(db, conf.MysqlDb.Driver)
}
