package schema

import (
	"content_service/api"
	"content_service/conf"
	"content_service/env"
	"reflect"
	"testing"
)

func TestToPredictorModelRecord(t *testing.T) {
	tables := []struct {
		db_model_history    ModelHistory
		expect_model_record api.PredictorModelRecord
	}{
		// case 1 构造ModelRecord,判断是否与期待相同
		{
			// model_history
			ModelHistory{
				ModelName: "model_a",
				Timestamp: "ts_1",
				IsLocked:  0,
				Md5:       "456",
			},
			api.PredictorModelRecord{
				Name:       "model_a",
				Timestamp:  "ts_1",
				FullName:   "model_a-ts_1",
				ConfigName: "/tmp/model_a-ts_1/model_a.json",
				IsLocked:   0,
				Md5:        "456",
			},
		},
		// case 2 改变部分字段，判断与期待是否相同
		{
			// model_history
			ModelHistory{
				ModelName: "model_b",
				Timestamp: "ts_2",
				IsLocked:  0,
				Md5:       "123",
			},
			api.PredictorModelRecord{
				Name:       "model_b",
				Timestamp:  "ts_2",
				FullName:   "model_b-ts_2",
				ConfigName: "/tmp/model_b-ts_2/model_b.json",
				IsLocked:   0,
				Md5:        "123",
			},
		},
	}

	confPtr := &conf.Conf{}
	envPtr := &env.Env{
		Conf: confPtr,
	}
	envPtr.Conf.P2PModelService.DestPath = "/tmp"
	for _, table := range tables {
		model_history := table.db_model_history
		model_record := model_history.ToPredictorModelRecord(envPtr)
		// 比较model_record 和 期望是否相等
		if !reflect.DeepEqual(model_record, table.expect_model_record) {
			t.Errorf("TestToPredictorModelRecord DeepEqual fail, expected model_record:%+v, expect_model_record: %+v",
				model_record, table.expect_model_record)
		}
	}
}
