package util

import (
	"fmt"
	"math"
	"regexp"
	"server/libs/logger"
	"strconv"
	"strings"
)

const KBToGB = 1024 * 1024

// 正则匹配，获取参数[context: url文本内容 regex: 正则表达式]
func GetResourceDataByRegex(context string, regex string) int64 {
	regex_str := regexp.MustCompile(regex)
	total_regex_str_arr := regex_str.FindAllString(context, -1)
	if len(total_regex_str_arr) == 0 {
		return 0
	}
	total_regex_str := total_regex_str_arr[0]
	total_exponene_str := strings.Split(total_regex_str, " ")[1]
	return ComputeExponent(total_exponene_str) / KBToGB
}

// 指数计算
func ComputeExponent(mem_exponent string) int64 {
	mem_and_exponent := strings.Split(mem_exponent, "e+")
	var total_mem int64
	mem_num, err := strconv.ParseFloat(mem_and_exponent[0], 64)
	if err != nil {
		logger.Errorf("strconv.ParseFloat err: %v", err)
	}
	if len(mem_and_exponent) != 2 {
		logger.Errorf("mem and exponent error")
		return total_mem
	}
	exponnet_num, error := strconv.ParseInt(mem_and_exponent[1], 10, 32)
	if error != nil {
		logger.Errorf("strconv.ParseInt err: %v", error)
	}
	total_mem = int64(mem_num * (math.Pow10(int(exponnet_num))) / 1024)
	return total_mem
}

// 指数计算
func ComputeExponentInt64(mem_exponent string) (int64, error) {
	var total_mem int64
	if !strings.Contains(mem_exponent, "e+") {
		mem_num, err := strconv.ParseFloat(mem_exponent, 64)
		if err != nil {
			return total_mem, fmt.Errorf("strconv.ParseFloat err: %v", err)
		}
		return int64(mem_num), nil
	}
	mem_and_exponent := strings.Split(mem_exponent, "e+")
	mem_num, err := strconv.ParseFloat(mem_and_exponent[0], 64)
	if err != nil {
		return total_mem, fmt.Errorf("strconv.ParseFloat err: %v", err)
	}
	if len(mem_and_exponent) != 2 {
		err := fmt.Errorf("mem and exponent error")
		return total_mem, err
	}
	exponnet_num, error := strconv.ParseInt(mem_and_exponent[1], 10, 32)
	if error != nil {
		return total_mem, fmt.Errorf("strconv.ParseInt err: %v", error)
	}
	total_mem = int64(mem_num * (math.Pow10(int(exponnet_num))))
	return total_mem, nil
}

func GetCpuBySeconds(senconds1 int64, t1 int64, senconds2 int64, t2 int64, coreNum int) (float64, error) {
	var res float64
	if t1-t2 <= 0 || coreNum <= 0 {
		return res, fmt.Errorf("divisor is lte zero")
	}
	res = float64(1) - float64(senconds1-senconds2)/float64(t1-t2)/float64(coreNum)
	return res, nil
}
