package util

// 获取两个int中最小的一个
func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// 获取两个int中最大的一个
func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}
