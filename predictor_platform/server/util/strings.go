package util

import (
	"strconv"
)

// 将[]uint以sep为间隔拼接字符串
// eg. elems = []uint{3,5,23} sep = "_"
// return "3_5_23"
func JoinUint(elems []uint, sep string) string {
	var joinStr string
	for idx, elem := range elems {
		joinStr = joinStr + strconv.Itoa(int(elem))
		if idx != len(elems)-1 {
			joinStr = joinStr + "_"
		}
	}
	return joinStr
}
