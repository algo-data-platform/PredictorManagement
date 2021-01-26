package common

import (
	"content_service/conf"
	"content_service/libs/logger"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"net"
	"os"
	"os/exec"
	"strings"
)

const Delimiter string = "-"
const CTRL_A string = "\001"
const ConfigSuffix string = ".json"
const LOCAL_IP = "LOCAL_IP"
const PredictorRouterMode = "RouterMode"
const PredictorServerMode = "ServerMode"
const PredictorRouterServiceName = "predictor_router_service"

func GetLocalIp(conf *conf.Conf) (string, error) {
	var ip = conf.LocalIp
	if ip != LOCAL_IP && ip != "" { // very weak sanity check
		return ip, nil
	}
	// in case no valid ip was set, try get local ip
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid ip address!")
}

// find what's missing in dest array according to src array
func FindComplement(src []string, dest []string) []string {
	complement := []string{}
	for _, e := range src {
		if !Contains(dest, e) {
			complement = append(complement, e)
		}
	}
	return complement
}

// check if a string exists in a slice
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// prettify a struct with json style, for debug logging purpose
// when error occours returns a hard-coded string "[[PrettyError]]"
func Pretty(v interface{}) string {
	if res, err := json.MarshalIndent(v, "", "  "); err != nil {
		return "[[PrettyError]]"
	} else {
		return string(res)
	}
}

// find files on os
func FindFiles(dir string, pattern string, daysToKeep string) ([]string, error) {
	cmd := exec.Command("find", dir, "-name", pattern, "-type", "d", "-mtime", "+"+daysToKeep)
	if stdout, err := cmd.CombinedOutput(); err != nil {
		return []string{}, fmt.Errorf("cmd failed: err=%v, output=%s, cmd=%v", err, stdout, cmd)
	} else {
		files := strings.Split(string(stdout), "\n")
		// remove last one if it is an empty string
		// happens a lot when parsed from stdout
		if len(files) != 0 && files[len(files)-1] == "" {
			files = files[:len(files)-1]
		}
		return files, nil
	}
}

// delete files
func DeleteFiles(files []string) error {
	for _, file := range files {
		// safe guarding against bad stuff
		if file == "/" || file == "../" || file == "../.." {
			return fmt.Errorf("trying to remove root dir! stop!!!")
		}
		if file == "" {
			continue
		}

		cmd := exec.Command("rm", "-rf", file)
		if stdout, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cmd failed: err=%v, output=%s, cmd=%v", err, stdout, cmd)
		}

		logger.Infof("deleted file: %v", file)
	}

	return nil
}

// a simple hash function
func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// TagError 为error的duck类型，扩充了tag string
type TagError struct {
	ErrMsg string // for error string
	ErrTag string // for extend tag string
}

func (tagError *TagError) Error() string {
	return tagError.ErrMsg
}

// 过滤带版本模型名字后面的版本号
func TrimModelVersion(model_name_with_version string) string {
	return strings.Split(model_name_with_version, Delimiter)[0]
}

// 获取两个int32中最大的一个
func MaxInt32(x, y int32) int32 {
	if x > y {
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

// 获取两个int中最小的一个
func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
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

// 过滤字符串数组中的空串和重复串
func RemoveRepeatAndEmpty(list []string) []string {
	mapdata := make(map[string]int)
	if len(list) <= 0 {
		return nil
	}
	for _, str := range list {
		mapdata[strings.TrimSpace(str)] = 1
	}
	var ret_list []string
	for k, _ := range mapdata {
		if k == "" {
			continue
		}
		ret_list = append(ret_list, k)
	}
	return ret_list
}

// 平均切割一个slice为多个slice
// 如长度为23的一个数组切割为长度为8,8,7，而不是7,7,9
// example: ["1","2","3","4","5","6","7","8"] 分成3份为 [["1","2","3"],["4","5","6"],["7","8"]]
func DivideSlices(s []string, num int) [][]string {
	childList := make([][]string, 0, num)
	if len(s) == 0 || num <= 0 {
		return childList
	}
	minSize := int(math.Floor(float64(len(s)) / float64(num)))
	remainder := len(s) % num
	var end int
	for i := 0; i < num; i++ {
		add := 0
		if remainder > 0 {
			add = 1
			remainder = remainder - 1
		}
		start := end
		end = start + minSize + add
		childs := s[start:end]
		childList = append(childList, childs)
	}
	return childList
}

// 获取元素在slice的第一个索引位置
// @return bool 是否找到，int 所在位置
func GetIndexFromSlice(item string, slice []string) (bool, int) {
	if len(slice) == 0 {
		return false, 0
	}
	for index, value := range slice {
		if value == item {
			return true, index
		}
	}
	return false, 0
}

// 获取元素在二维slice的索引位置
// @return bool 是否找到，int 所在位置
func GetIndexFromChildList(item string, childList [][]string) (bool, int) {
	if len(childList) == 0 {
		return false, 0
	}
	for index, childs := range childList {
		if IsInSliceString(item, childs) {
			return true, index
		}
	}
	return false, 0
}

// 目录是否存在
func IsDir(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil {
		return false
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

// 解析域名对应的ip地址，支持ip地址传入
func ResolveIP(domain_name string) (string, error) {
	addr, err := net.ResolveIPAddr("ip", domain_name)
    if err != nil {
        return "", fmt.Errorf("Resolvtion error : %v", err.Error())
	}
	return addr.String(), nil
}
// 判断两个float vector是否余弦相似，精度1e-6
// 余弦相似度计算公式为  cos_sim = vec_a与vec_b的点积 / (vec_a模 * vec_b模)
func IsFloatVectorCosineSim(vec_a []float64, vec_b []float64) (bool,float64) {
  if len(vec_a) != len(vec_b) {
    return false, 0
  }
  vec_size := len(vec_a)
  var module_a,module_b,module_a_squa,module_b_squa,scalar_product,similarity float64
  for i:=0; i<vec_size; i++ {
    module_a_squa += math.Pow(vec_a[i], 2)
    module_b_squa += math.Pow(vec_b[i], 2)
    scalar_product += vec_a[i] * vec_b[i]
  }
  module_a = math.Sqrt(module_a_squa)
  module_b = math.Sqrt(module_b_squa)
  similarity = scalar_product / (module_a * module_b)
  if math.Abs((1-similarity)*1e+8) > 100 {
    return false, similarity
  }
  return true, similarity
}
