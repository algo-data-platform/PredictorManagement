package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"server/common/response"
	"server/env"
	"server/libs/logger"
	"server/schema"
	"server/server/dao"
	"server/server/logics"
	"server/util"
	"strconv"
	"strings"
)

// get table list
func MysqlTables(context *gin.Context) {
	var all_table util.AllTable
	all_table.Database = env.Env.Conf.MysqlDb.Database
	all_table.TableNames = env.Env.Conf.MysqlDb.TableNames
	all_table_marshal, err := json.Marshal(all_table)
	if err != nil {
		logger.Errorf("marshal json error: %v", err)
		context.String(http.StatusForbidden, string("marshal json error"))
	} else {
		context.String(http.StatusOK, string(all_table_marshal))
	}
}

// 展示具体table的mysql视图
func MysqlShow(context *gin.Context) {
	table := context.Query("table")
	parse_data := dao.ShowTableData(table)
	if parse_data != nil {
		context.String(http.StatusOK, string(parse_data))
	} else {
		context.String(http.StatusBadRequest, string("show table fail"))
	}
}

// insert mysql数据
func MysqlInsert(context *gin.Context) {
	table := context.Query("table")
	var response_str string
	var insertRes bool
	switch table {
	case "hosts":
		host := &schema.Host{
			Ip:         context.Query("ip"),
			DataCenter: context.Query("data_center"),
			Desc:       context.Query("desc"),
		}
		err := dao.CreateHost(host)
		if err != nil {
			insertRes = false
			break
		}
		insertRes = true
	case "services":
		service := &schema.Service{
			Name: context.Query("name"),
			Desc: context.Query("desc"),
		}
		err := dao.CreateService(service)
		if err != nil {
			insertRes = false
			break
		}
		insertRes = true
	case "models":
		model := &schema.Model{
			Name: context.Query("name"),
			Path: context.Query("path"),
			Desc: context.Query("desc"),
		}
		err := dao.CreateModel(model)
		if err != nil {
			insertRes = false
			break
		}
		insertRes = true
	case "host_services":
		hid, err := strconv.ParseUint(context.Query("hid"), 10, 64)
		if err != nil {
			logger.Errorf("parse hid error: %v", err)
			insertRes = false
			break
		}
		sid, err := strconv.ParseUint(context.Query("sid"), 10, 64)
		if err != nil {
			logger.Errorf("parse sid error: %v", err)
			insertRes = false
			break
		}
		load_weight, err := strconv.ParseUint(context.Query("load_weight"), 10, 64)
		if err != nil {
			logger.Errorf("parse load_weight error: %v", err)
			insertRes = false
			break
		}
		host_service := &schema.HostService{
			Hid:        uint(hid),
			Sid:        uint(sid),
			LoadWeight: uint(load_weight),
			Desc:       context.Query("desc"),
		}
		err = dao.CreateHostService(host_service)
		if err != nil {
			insertRes = false
			break
		}
		insertRes = true
	case "service_models":
		sid, err := strconv.ParseUint(context.Query("sid"), 10, 64)
		if err != nil {
			logger.Errorf("parse sid error: %v", err)
			insertRes = false
			break
		}
		mid, err := strconv.ParseUint(context.Query("mid"), 10, 64)
		if err != nil {
			logger.Errorf("parse mid error: %v", err)
			insertRes = false
			break
		}
		service_model := &schema.ServiceModel{
			Sid:  uint(sid),
			Mid:  uint(mid),
			Desc: context.Query("desc"),
		}
		if err := dao.CreateServiceModel(service_model); err != nil {
			insertRes = false
			break
		}
		insertRes = true
	case "model_histories":
		var is_locked uint64
		var err error
		is_locked_query := context.Query("is_locked")
		if len(is_locked_query) != 0 {
			is_locked, err = strconv.ParseUint(is_locked_query, 10, 64)
			if err != nil {
				logger.Errorf("parse is_locked error: %v", err)
				insertRes = false
				break
			}
		} else {
			is_locked = 0
		}
		model_history := &schema.ModelHistory{
			ModelName: context.Query("model_name"),
			Timestamp: context.Query("timestamp"),
			Md5:       context.Query("md5"),
			IsLocked:  uint(is_locked),
			Desc:      context.Query("desc"),
		}
		if err := dao.CreateModelHistory(model_history); err != nil {
			logger.Errorf("create table fail: %v", err)
			//update,add send mail when fail
			mail_subject := "create model history fail"
			mail_body := fmt.Sprintf("%v", err)
			err := env.Env.Mailer.SendMail("Predictor", env.Env.Conf.FromEmail, env.Env.Conf.MysqlDb.AlarmList, mail_subject, []byte(mail_body))
			if err != nil {
				logger.Errorf("send mail error: %v", err)
			} else {
				logger.Infof("send mail succ!")
			}
			insertRes = false
			break
		}
		var model *schema.Model
		model_name := context.Query("model_name")
		model, err = dao.GetModelByName(model_name)
		if err != nil {
			insertRes = false
			break
		}
		// 如果模型不存在，插入模型
		if model.ID == 0 {
			model = &schema.Model{
				Name: context.Query("model_name"),
				Path: "/data0/vad/algo_service/data/",
				Desc: "",
			}
			if err := dao.CreateModel(model); err != nil {
				logger.Errorf("insert record into  table fail: %s", table)
				mail_subject := "insert models fail"
				mail_body := fmt.Sprintf("%v", err)
				err := env.Env.Mailer.SendMail("Predictor", env.Env.Conf.FromEmail, env.Env.Conf.MysqlDb.AlarmList, mail_subject, []byte(mail_body))
				if err != nil {
					logger.Errorf("send mail error: %v", err)
				} else {
					logger.Infof("send mail succ!")
				}
				insertRes = false
				break
			}
			var err error
			model, err = dao.GetModelByName(model.Name)
			if err != nil {
				insertRes = false
				break
			}
		}
		// insert record into service_models
		var consistence_service *schema.Service
		consistence_service, err = dao.GetServiceByName(env.Env.Conf.ConsistenceService)
		if err != nil {
			insertRes = false
			break
		}
		var service_model *schema.ServiceModel
		exists, err := dao.ExistsServiceModelBySidMid(consistence_service.ID, model.ID)
		if err != nil {
			insertRes = false
			break
		}
		if !exists {
			service_model = &schema.ServiceModel{
				Sid:  uint(consistence_service.ID),
				Mid:  uint(model.ID),
				Desc: fmt.Sprintf("algo_service_consistence -> %v", model.Name),
			}
			if err := dao.CreateServiceModel(service_model); err != nil {
				logger.Errorf("insert record into table fail: %v", err)
				mail_subject := "insert service_models fail"
				mail_body := fmt.Sprintf("%v", err)
				err := env.Env.Mailer.SendMail("Predictor", env.Env.Conf.FromEmail, env.Env.Conf.MysqlDb.AlarmList, mail_subject, []byte(mail_body))
				if err != nil {
					logger.Errorf("send mail error: %v", err)
				} else {
					logger.Infof("send mail succ!")
				}
				insertRes = false
				break
			}
		}
		insertRes = true
	case "configs":
		config := &schema.Config{
			Config:      context.Query("config"),
			Description: context.Query("desc"),
		}
		err := dao.CreateConfig(config)
		if err != nil {
			insertRes = false
			break
		}
		insertRes = true
	case "service_configs":
		cid, err := strconv.ParseUint(context.Query("cid"), 10, 64)
		if err != nil {
			logger.Errorf("parse cid error: %v", err)
			insertRes = false
			break
		}
		sid, err := strconv.ParseUint(context.Query("sid"), 10, 64)
		if err != nil {
			logger.Errorf("parse sid error: %v", err)
			insertRes = false
			break
		}
		service_config := &schema.ServiceConfig{
			Cid:         uint(cid),
			Sid:         uint(sid),
			Description: context.Query("desc"),
		}
		err = dao.CreateServiceConfig(service_config)
		if err != nil {
			insertRes = false
			break
		}
		insertRes = true
	default:
		logger.Warnf("insert table not exist!")
		insertRes = false
		break
	}

	if insertRes {
		response_str = fmt.Sprintf("insert into  table %s succ", table)
		context.String(http.StatusOK, response_str)
	} else {
		response_str = fmt.Sprintf("insert table %s fail", table)
		context.String(http.StatusBadRequest, response_str)
	}
}

// delete mysql 数据
func MysqlDelete(context *gin.Context) {
	table := context.Query("table")
	var response_str string
	delete_status := dao.DeleteTableData(table, context, env.Env.Conf)
	if delete_status {
		response_str = fmt.Sprintf("delete table %s's data succ", table)
		context.String(http.StatusOK, response_str)
	} else {
		response_str = fmt.Sprintf("delete table %s's data fail", table)
		context.String(http.StatusBadRequest, response_str)
	}
}

// update mysql数据
func MysqlUpdate(context *gin.Context) {
	table := context.Query("table")
	var response_str string
	var update_status bool
	switch table {
	case "host_services":
		id, err := strconv.ParseUint(context.Query("id"), 10, 64)
		if err != nil {
			errStr := fmt.Sprintf("parse id error: %v", err)
			logger.Errorf(errStr)
			response_str = fmt.Sprintf("update table %s's data fail, err: %s", table, errStr)
			context.String(http.StatusBadRequest, response_str)
			return
		}
		if id <= 0 {
			response_str = fmt.Sprintf("update table %s's data fail, err: id is lte zero", table)
			context.String(http.StatusBadRequest, response_str)
			return
		}
		idUint := uint(id)
		hid, err := strconv.ParseUint(context.Query("hid"), 10, 64)
		if err != nil {
			errStr := fmt.Sprintf("parse hid error: %v", err)
			logger.Errorf(errStr)
			response_str = fmt.Sprintf("update table %s's data fail, err: %s", table, errStr)
			context.String(http.StatusBadRequest, response_str)
			return
		}
		hidUint := uint(hid)
		sid, err := strconv.ParseUint(context.Query("sid"), 10, 64)
		if err != nil {
			errStr := fmt.Sprintf("parse sid error: %v", err)
			logger.Errorf(errStr)
			response_str = fmt.Sprintf("update table %s's data fail, err: %s", table, errStr)
			context.String(http.StatusBadRequest, response_str)
			return
		}
		sidUint := uint(sid)
		desc := context.Query("desc")
		load_weight, err := strconv.ParseUint(context.Query("load_weight"), 10, 64)
		if err != nil {
			errStr := fmt.Sprintf("parse load_weight error: %v", err)
			logger.Errorf(errStr)
			response_str = fmt.Sprintf("update table %s's data fail, err: %s", table, errStr)
			context.String(http.StatusBadRequest, response_str)
			return
		}
		loadWeightUint := uint(load_weight)
		originHostService, err := dao.GetHostServiceById(uint(id))
		if err != nil {
			errStr := fmt.Sprintf("GetHostServiceById error: %v", err)
			logger.Errorf(errStr)
			response_str = fmt.Sprintf("update table %s's data fail, err: %s", table, errStr)
			context.String(http.StatusBadRequest, response_str)
			return
		}
		if sidUint != originHostService.Sid {
			loadWeightUint, err = logics.GetServiceInitWeightBySid(sidUint, "")
			if err != nil {
				errStr := fmt.Sprintf("GetServiceInitWeightBySid error: %v", err)
				logger.Errorf(errStr)
				response_str = fmt.Sprintf("update table %s's data fail, err: %s", table, errStr)
				context.String(http.StatusBadRequest, response_str)
				return
			}
		}
		// gorm在更新时，如果用model方式字段为默认值，不会更新，采用map的方式可以避免这个问题
		hostServiceMap := map[string]interface{}{
			"hid":         hidUint,
			"sid":         sidUint,
			"load_weight": loadWeightUint,
			"desc":        desc,
		}
		err = dao.UpdateHostServiceData(idUint, hostServiceMap)
		if err != nil {
			update_status = false
		} else {
			update_status = true
		}
	case "service_configs":
		id, err := strconv.ParseUint(context.Query("id"), 10, 64)
		if err != nil {
			errStr := fmt.Sprintf("parse id error: %v", err)
			logger.Errorf(errStr)
			response_str = fmt.Sprintf("update table %s's data fail, err: %s", table, errStr)
			context.String(http.StatusBadRequest, response_str)
			return
		}
		if id <= 0 {
			response_str = fmt.Sprintf("update table %s's data fail, err: id is lte zero", table)
			context.String(http.StatusBadRequest, response_str)
			return
		}
		idUint := uint(id)
		cid, err := strconv.ParseUint(context.Query("cid"), 10, 64)
		if err != nil {
			errStr := fmt.Sprintf("parse hid error: %v", err)
			logger.Errorf(errStr)
			response_str = fmt.Sprintf("update table %s's data fail, err: %s", table, errStr)
			context.String(http.StatusBadRequest, response_str)
			return
		}
		cidUint := uint(cid)
		sid, err := strconv.ParseUint(context.Query("sid"), 10, 64)
		if err != nil {
			errStr := fmt.Sprintf("parse sid error: %v", err)
			logger.Errorf(errStr)
			response_str = fmt.Sprintf("update table %s's data fail, err: %s", table, errStr)
			context.String(http.StatusBadRequest, response_str)
			return
		}
		sidUint := uint(sid)
		desc := context.Query("desc")
		serviceConfigMap := map[string]interface{}{
			"cid":         cidUint,
			"sid":         sidUint,
			"description": desc,
		}
		update_status = dao.UpdateServiceConfigData(idUint, serviceConfigMap)
	default:
		update_status = dao.UpdateTableData(table, context, env.Env.Conf)
	}
	if update_status {
		response_str = fmt.Sprintf("update table %s's data succ", table)
		context.String(http.StatusOK, response_str)
	} else {
		response_str = fmt.Sprintf("update table %s's data fail", table)
		context.String(http.StatusBadRequest, response_str)
	}
}

// 批量添加机器
func MysqlBatchInsertHosts(context *gin.Context) {
	host_ips := context.PostForm("host_ips")
	if len(host_ips) == 0 {
		response.ResultWithoutData(201, "host_ips is empty", context)
		return
	}
	hostIps := strings.Split(host_ips, "\n")
	hostIpList := []string{}
	for _, hostIp := range hostIps {
		hostIp = strings.TrimSpace(hostIp)
		if len(hostIp) == 0 {
			continue
		}
		address := net.ParseIP(hostIp)
		if address == nil {
			response.ResultWithoutData(201, fmt.Sprintf("host_ips is not valid: %s", hostIp), context)
			return
		}
		hostIpList = append(hostIpList, hostIp)
	}
	if len(hostIpList) == 0 {
		response.ResultWithoutData(202, "host_ips is empty", context)
		return
	}

	err := logics.BatchInsertHostIps(hostIpList)
	if err != nil {
		response.ResultWithoutData(202, fmt.Sprintf("insert host_ips error: %v", err), context)
		return
	}
	response.DoneWithMessage(fmt.Sprintf("%d台机器添加成功", len(hostIpList)), context)
	return
}
