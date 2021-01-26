package logics

import (
	"fmt"
	"server/common"
	"server/conf"
	"server/libs/logger"
	"server/server/dao"
	"server/util"
	"sort"
	"strings"
	"time"
)

// 两个返回值，一个用于渲染前端界面，另一个用于html/template生成统计报表
func ModelListUpdateTimeWithinWeek(modelList []string, conf *conf.Conf) ([]util.ModelUpdateTimingInfo, map[string]interface{}) {
	var model_list_update_timing_info []util.ModelUpdateTimingInfo
	// 生成和html文件对应的数据
	report_map := make(map[string]interface{})
	report_map["Date"] = time.Now().Format("2006-01-02 15:04:05")
	//report_map["TableHeader"] = []string{"模型名称", "模型业务线", "7天平均更新频率", "最近一次更新时间"}
	report_map["TableHeader"] = []string{"模型业务线", "模型名称", "7天平均更新频率", "最近一次更新时间"}
	var model_datas []interface{}

	for index := 0; index < len(modelList); index++ {
		cur_model_update_info := ModelUpdateTimeWithinWeek(modelList[index], conf)
		model_list_update_timing_info = append(model_list_update_timing_info, cur_model_update_info)
		var latest_update_time string
		if len(cur_model_update_info.LastestTimestampArray) > 0 {
			time_str := cur_model_update_info.LastestTimestampArray[0]
			latest_update_time = fmt.Sprintf("%v-%v-%v %v:%v:%v", time_str[0:4], time_str[4:6], time_str[6:8],
				time_str[9:11], time_str[11:13], time_str[13:15])
		}
		update_time_weekly := fmt.Sprintf("%v", cur_model_update_info.ModelUpdateTimeWeekly)
		each_model_timing := []string{cur_model_update_info.ModelChannel, cur_model_update_info.ModelName,
			util.SecondsToHM(update_time_weekly), latest_update_time}
		model_datas = append(model_datas, each_model_timing)
	}
	report_map["ModelData"] = model_datas
	// sort data(by model_name)
	sort.Sort(common.ModelUpdateTimingInfoArrayWrapper{model_list_update_timing_info, func(p, q *util.ModelUpdateTimingInfo) bool {
		return p.ModelChannel < q.ModelChannel
	}})
	return model_list_update_timing_info, report_map
}

// 以一周为单位，计算一周内model更新耗时[单位为seconds]
func ModelUpdateTimeWithinWeek(model_name string, conf *conf.Conf) util.ModelUpdateTimingInfo {
	// mysql's timestamp format [20191205_125040]
	var model_update_time_info util.ModelUpdateTimingInfo
	model_update_time_info.ModelName = model_name
	time_now := time.Now().AddDate(0, 0, conf.ModelTimingRange).Format("2006-01-02 15:04:05")
	time_now_arr := strings.Split(time_now, " ")
	if len(time_now_arr) != 2 {
		logger.Errorf("time[ymd] not correct format")
	}
	ymd_str := time_now_arr[0]
	hms_str := time_now_arr[1]
	prefix_str := ymd_str[0:4] + ymd_str[5:7] + ymd_str[8:10]

	hms_arr := strings.Split(hms_str, ":")
	if len(hms_arr) != 3 {
		logger.Errorf("time[hms] not correct foramt")
	}
	week_timestamp := fmt.Sprintf("%v_%v%v%v", prefix_str, hms_arr[0], hms_arr[1], hms_arr[2])
	// 	从mysql table拿取timestamp > week_timestamp 并且model_name
	model_weekly_info, err := dao.GetModelWeeklyInfo(model_name, week_timestamp)
	if err != nil {
		logger.Errorf("GetModelWeeklyInfo fail, err: %v", err)
		return model_update_time_info
	}
	model_weekly_num := len(model_weekly_info)
	var model_update_interval int64
	var model_update_timestamp []string
	// update, get channel models[desc]
	model_update_time_info.ModelChannel = dao.GetModelType(model_name)
	if model_weekly_num > 0 {
		for index := 0; index < model_weekly_num; index++ {
			curTimestamp := model_weekly_info[index].Timestamp
			model_update_timestamp = append(model_update_timestamp, curTimestamp)
			nextTimestamp := ""
			if index == 0 {
				nextTimestamp = time.Now().Format("20060102_150405")
			} else {
				nextTimestamp = model_weekly_info[index-1].Timestamp
			}
			each_interval := util.GetTimestampInterval(nextTimestamp, curTimestamp)
			model_update_interval += each_interval
		}
		diff_secods := model_update_interval / int64(model_weekly_num) // 平均update 更新耗时，单位为秒
		model_update_time_info.ModelUpdateTimeWeekly = diff_secods
		model_update_time_info.LastestTimestampArray = model_update_timestamp
	}
	return model_update_time_info
}
