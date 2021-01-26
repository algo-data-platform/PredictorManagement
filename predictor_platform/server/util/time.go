package util

import (
	"errors"
	"fmt"
	"server/libs/logger"
	"strconv"
	"time"
)

func SecondsToHM(seconds string) string {
	n, err := strconv.ParseInt(seconds, 10, 64)
	if err != nil {
		logger.Errorf("%d of type %T, error:", n, n, err)
	}
	var hm_format_str string
	if n > 0 {
		hour := n / 3600
		minute := n % 3600 / 60
		hm_format_str = fmt.Sprintf("%d小时%d分钟", hour, minute)
	} else {
		hm_format_str = "7日无更新"
	}
	return hm_format_str
}

// format ["2006-01-02 15:04:05"]
func StringIntoValidTimeFormat(time_str string) (time.Time, error) {
	if len(time_str) < 15 {
		return time.Now(), errors.New("input time invalid!")
	}
	time_format := fmt.Sprintf("%v-%v-%v %v:%v:%v", time_str[0:4], time_str[4:6], time_str[6:8],
		time_str[9:11], time_str[11:13], time_str[13:15])
	return time.Parse("2006-01-02 15:04:05", time_format)
}

// 计算两个timestamp时间差，单位为秒[20200308_184016 20200306_123826]
func GetTimestampInterval(timestamp1 string, timestamp2 string) int64 {
	var update_interval int64
	//suppose timestamp1 is larger than timestamp2
	if timestamp1 < timestamp2 {
		return GetTimestampInterval(timestamp2, timestamp1)
	}
	time_format_1, err := StringIntoValidTimeFormat(timestamp1)
	if err != nil {
		logger.Errorf("format string to time style error: %s", timestamp1)
		return update_interval
	}
	time_format_2, err_ := StringIntoValidTimeFormat(timestamp2)
	if err_ != nil {
		logger.Errorf("format string to time style error: %s", timestamp2)
		return update_interval
	}
	return time_format_1.Unix() - time_format_2.Unix()
}

// 美化时间
func PrettyTime(second int64) string {
	var res string
	switch {
	case second > 3600*24:
		res = fmt.Sprintf("%.1f day", float64(second)/float64(3600*24))
	case second > 3600:
		res = fmt.Sprintf("%.1f hour", float64(second)/float64(3600))
	case second > 60:
		res = fmt.Sprintf("%.1f min", float64(second)/float64(60))
	default:
		res = fmt.Sprintf("%d sec", second)
	}
	return res
}
