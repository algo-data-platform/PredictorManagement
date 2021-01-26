package util

// 去重函数
func GetUniqueArray(contexts []string) []string {
	unique_array := make([]string, 0, len(contexts))
	tmp := map[string]struct{}{}
	for _, item := range contexts {
		if _, ok := tmp[item]; !ok {
			tmp[item] = struct{}{}
			unique_array = append(unique_array, item)
		}
	}
	return unique_array
}

// 判断元素是否存在slice中
func IsInSliceString(item string, slice []string) bool {
	if len(slice) == 0 {
		return false
	}
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}

// 判断元素是否存在slice中
func IsInSliceUint(item uint, slice []uint) bool {
	if len(slice) == 0 {
		return false
	}
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}

// 判断两个slice 是否相同
func IsEqualSliceString(sliceA, sliceB []string) bool {
	if len(sliceA) != len(sliceB) {
		return false
	}
	var mapA = make(map[string]bool, len(sliceA))
	for _, rowA := range sliceA {
		mapA[rowA] = true
	}
	for _, rowB := range sliceB {
		if _, ok := mapA[rowB]; !ok {
			return false
		}
	}
	return true
}

// 判断一个slice是否为另一个的子集
func IsSubSliceString(sliceA, sliceB []string) bool {
	if len(sliceA) > len(sliceB) {
		return false
	}
	var mapB = make(map[string]bool, len(sliceB))
	for _, rowB := range sliceB {
		mapB[rowB] = true
	}
	for _, rowA := range sliceA {
		if _, ok := mapB[rowA]; !ok {
			return false
		}
	}
	return true
}

// 删除slice中的第一个等于item的元素，会改变元素顺序，但是速度快
func DelSliceFirstItem(s *[]string, item string) {
	if len(*s) == 0 {
		return
	}
	var found bool
	var idx int
	var v string
	for idx, v = range *s {
		if v == item {
			found = true
			break
		}
	}
	if !found {
		return
	}
	if len(*s) == 1 {
		*s = []string{}
		return
	}
	(*s)[idx] = (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
}

// 从slice中排除excludeSlice
func ExcludeSliceString(sourceSlice, excludeSlice []string) []string {
	var excludeMap = make(map[string]struct{}, len(excludeSlice))
	for _, row := range excludeSlice {
		excludeMap[row] = struct{}{}
	}
	var idx int = 0
	for _, row := range sourceSlice {
		if _, exists := excludeMap[row]; !exists {
			sourceSlice[idx] = row
			idx++
		}
	}
	return sourceSlice[:idx]
}

// 获取两个slice的差集，sliceA不存在sliceB的元素列表
func DiffSliceString(sliceA, sliceB []string) []string {
	var diffSlice = []string{}
	var mapB = make(map[string]bool, len(sliceB))
	for _, rowB := range sliceB {
		mapB[rowB] = true
	}
	for _, rowA := range sliceA {
		if _, ok := mapB[rowA]; !ok {
			diffSlice = append(diffSlice, rowA)
		}
	}
	return diffSlice
}
