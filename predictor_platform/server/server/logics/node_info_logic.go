package logics

import (
	"encoding/json"
	"fmt"
	"regexp"
	"server/env"
	"server/libs/logger"
	"server/schema"
	"server/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

var NodeResMap map[string]util.NodeResourceInfo

func GetNodeMetaInfo(meta_data_url string, timeout int) util.NodeMetaInfo {
	logger.Infof("get meta_data_url: %s", meta_data_url)
	var node_meta_info util.NodeMetaInfo
	resp_meta_data_content, _ := util.GetContentFromUrl(meta_data_url, time.Millisecond*time.Duration(timeout), 3)
	if err := json.Unmarshal(resp_meta_data_content, &node_meta_info); err != nil {
		logger.Errorf("json meta data unmarshal error: %v", err)
	}
	return node_meta_info
}

func fetchNodeInfoData(timeout int, row schema.Host, predictor_http_port int, ch chan<- util.NodeInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	var cur_node_info util.NodeInfo
	cur_node_info.Host = row.Ip
	cur_node_info.DataCenter = row.DataCenter
	// metric url为http://host:9100/metric
	cur_node_metric := fmt.Sprintf("http://%s:9100/metrics", row.Ip)
	cur_node_info.ResourceInfo = GetResourceInfo(cur_node_metric, timeout, row.Ip)
	// meta url 为http://host:port/get_service_model_info
	cur_node_meta := fmt.Sprintf("http://%s:%d/get_service_model_info", row.Ip, predictor_http_port)
	cur_node_info.StatusInfo = GetNodeMetaInfo(cur_node_meta, timeout).Msg.Services
	ch <- cur_node_info
}

func GetNodeInfoList(predictor_http_port int, timeout int) []util.NodeInfo {
	var wg sync.WaitGroup
	var infos []util.NodeInfo
	ch := make(chan util.NodeInfo, 1024)
	// get all ip list
	var rows []schema.Host
	if env.Env.MysqlDB.Find(&rows).RecordNotFound() {
		logger.Errorf("Mysql record not found")
		return infos
	}
	wg.Add(len(rows))
	for _, cur_row := range rows {
		go fetchNodeInfoData(timeout, cur_row, predictor_http_port, ch, &wg)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	for node_info := range ch {
		infos = append(infos, node_info)
	}
	return infos
}

func getLoadAverage(url_context string, query_re_pat string) float32 {
	var result float32
	reg := regexp.MustCompile(query_re_pat)
	load_average_str_arr := reg.FindAllString(url_context, -1)
	if len(load_average_str_arr) != 0 {
		load_average_str := load_average_str_arr[0]
		load_average, err := strconv.ParseFloat(strings.Split(load_average_str, " ")[1], 16)
		if err != nil {
			logger.Fatalf("strconv.ParseFloat err: %v", err)
		}
		result = float32(load_average)
	}
	return result
}

// 返回节点的load_average m1/5/15
func getLoadAverageAll(url_context string) util.LoadAverageAll {
	var load_average_all util.LoadAverageAll
	var load_average_re_pat string
	load_average_re_pat = `node_load1\s\d.+\d`
	load_average_all.Average1 = getLoadAverage(url_context, load_average_re_pat)
	load_average_re_pat = `node_load5\s\d.+\d`
	load_average_all.Average5 = getLoadAverage(url_context, load_average_re_pat)
	load_average_re_pat = `node_load15\s\d.+\d`
	load_average_all.Average15 = getLoadAverage(url_context, load_average_re_pat)
	return load_average_all
}

// 从9100端口获取cpu/mem/disk等状态信息
func GetResourceInfo(url_address string, http_timeout int, host string) util.NodeResourceInfo {
	var statusinfo util.NodeResourceInfo
	resp_content, _ := util.GetContentFromUrl(url_address, time.Millisecond*time.Duration(http_timeout), 3)
	resp_content_str := string(resp_content)
	if len(resp_content_str) != 0 {
		statusinfo.NodeAvail = util.Node_On
	} else {
		statusinfo.NodeAvail = util.Node_Off
	}
	// 获取cpu core
	core_re_pat := `cpu=.*?mode=`
	reg := regexp.MustCompile(core_re_pat)
	core_str := reg.FindAll(resp_content, -1)
	var cpu_core_str []string
	for i := 0; i < len(core_str); i++ {
		cpu_core_str = append(cpu_core_str, string(core_str[i]))
	}
	cpu_core_num := util.GetUniqueArray(cpu_core_str)
	statusinfo.CoreNum = len(cpu_core_num)

	// 获取total mem
	total_mem_re_pat := `MemTotal_bytes.*e\+.*`
	total_mem := util.GetResourceDataByRegex(resp_content_str, total_mem_re_pat)
	statusinfo.TotalMem = total_mem
	// 采用free mem
	free_mem_re_pat := `MemAvailable_bytes.*e\+.*`
	free_mem := util.GetResourceDataByRegex(resp_content_str, free_mem_re_pat)
	statusinfo.AvailMem = free_mem
	// 获取node_load1(亦即m1)
	load_average := getLoadAverageAll(resp_content_str)
	statusinfo.LoadAverage = load_average
	// 获取总disk(亦即/data0磁盘大小)
	total_filesytem_re_pat := `node_filesystem_size_bytes.*/data0.*`
	total_filesystem := util.GetResourceDataByRegex(resp_content_str, total_filesytem_re_pat)
	statusinfo.TotalDisk = total_filesystem
	// 获取free disk
	free_filesystem_re_pat := `node_filesystem_avail_bytes.*/data0.*`
	free_filesystem := util.GetResourceDataByRegex(resp_content_str, free_filesystem_re_pat)
	statusinfo.AvailDisk = free_filesystem
	// 获取cpu_idle_seconds_total

	var cpuIdleSecondsTotal int64
	core_idle_pat := `node_cpu_seconds_total\{cpu=.*?mode=\"idle\"}\s(.*)`
	reg = regexp.MustCompile(core_idle_pat)
	core_str = reg.FindAll([]byte(resp_content), -1)
	for i := 0; i < len(core_str); i++ {
		total_exponene_str := strings.Split(string(core_str[i]), " ")[1]
		idleSencondsRow, err := util.ComputeExponentInt64(total_exponene_str)
		if err != nil {
			logger.Errorf("ComputeExponentInt64 fail,url_address=%s, err=%v", url_address, err)
			continue
		}
		cpuIdleSecondsTotal = cpuIdleSecondsTotal + idleSencondsRow
	}
	statusinfo.LastCpuIdleSecondsTotal = cpuIdleSecondsTotal
	statusinfo.LastUptime = time.Now().Unix()
	// cpu
	if _, exists := NodeResMap[host]; exists {
		cpu, err := util.GetCpuBySeconds(
			cpuIdleSecondsTotal,
			time.Now().Unix(),
			NodeResMap[host].LastCpuIdleSecondsTotal,
			NodeResMap[host].LastUptime,
			len(cpu_core_num),
		)
		if err != nil {
			logger.Errorf("GetCpuBySeconds fail, err=%v", err)
		} else {
			// 防止遇到节点node_cpu_seconds_total无法取到的情况
			if cpuIdleSecondsTotal == 0 {
				cpu = float64(-1)
			}
			statusinfo.Cpu = cpu
			logger.Debugf("url_address=%v, cpu=%v", url_address, cpu)
		}
	}

	return statusinfo
}
