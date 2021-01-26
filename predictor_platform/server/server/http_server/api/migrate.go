package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"server/common/response"
	"server/libs/logger"
	"server/server/logics"
	"strconv"
	"strings"
)

type Migrate struct{}

// 获取统计数据
func (m *Migrate) GetServiceStats(context *gin.Context) {
	serviceStatis, err := logics.GetServiceStats()
	if err != nil {
		logger.Errorf("GetServiceStats error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("GetServiceStats error: %v", err), context)
		return
	}
	response.DoneWithData(serviceStatis, context)
	return
}

// 获取from_services
func (m *Migrate) GetFromServices(context *gin.Context) {
	fromServices, err := logics.GetFromServices()
	if err != nil {
		logger.Errorf("GetFromServices error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("GetFromServices error: %v", err), context)
		return
	}
	response.DoneWithData(fromServices, context)
	return
}

// 获取to_services
func (m *Migrate) GetToServices(context *gin.Context) {
	toServices, err := logics.GetToServices()
	if err != nil {
		logger.Errorf("GetToServices error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("GetToServices error: %v", err), context)
		return
	}
	response.DoneWithData(toServices, context)
	return
}

// preview
func (m *Migrate) Preview(context *gin.Context) {
	fromService := context.Query("from_service")
	if fromService == "" {
		logger.Errorf("from_service is empty")
		response.ResultWithoutData(201, "from_service is empty", context)
		return
	}
	fromSidStrs := strings.Split(fromService, ",")
	var fromSids = make([]uint, 0, len(fromSidStrs))
	for _, fromSidStr := range fromSidStrs {
		fromSidInt, err := strconv.Atoi(fromSidStr)
		if err != nil {
			response.ResultWithoutData(201, fmt.Sprintf("from_service type error: %v", err), context)
			return
		}
		fromSids = append(fromSids, uint(fromSidInt))
	}
	toService := context.Query("to_service")
	if toService == "" {
		logger.Errorf("to_service is empty")
		response.ResultWithoutData(201, "to_service is empty", context)
		return
	}
	toSidStrs := strings.Split(toService, ",")
	var toSids = make([]uint, 0, len(toSidStrs))
	for _, toSidStr := range toSidStrs {
		toSidInt, err := strconv.Atoi(toSidStr)
		if err != nil {
			response.ResultWithoutData(201, fmt.Sprintf("to_service type error: %v", err), context)
			return
		}
		toSids = append(toSids, uint(toSidInt))
	}
	num, err := strconv.Atoi(context.Query("num"))
	if err != nil {
		logger.Errorf("num type error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("num type error: %v", err), context)
		return
	}
	if num <= 0 {
		response.ResultWithoutData(201, "迁移台数不能小于0", context)
		return
	}
	previewHosts, err := logics.PreviewMigrateHosts(fromSids, toSids, num)
	if err != nil {
		logger.Errorf("PreviewMigrateHosts error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("PreviewMigrateHosts error: %v", err), context)
		return
	}
	response.DoneWithData(previewHosts, context)
	return
}

// 开始迁移
func (m *Migrate) DoMigrate(context *gin.Context) {
	fromService := context.Query("from_service")
	if fromService == "" {
		logger.Errorf("from_service is empty")
		response.ResultWithoutData(201, "from_service is empty", context)
		return
	}
	fromSidStrs := strings.Split(fromService, ",")
	var fromSids = make([]uint, 0, len(fromSidStrs))
	for _, fromSidStr := range fromSidStrs {
		fromSidInt, err := strconv.Atoi(fromSidStr)
		if err != nil {
			response.ResultWithoutData(201, fmt.Sprintf("from_service type error: %v", err), context)
			return
		}
		fromSids = append(fromSids, uint(fromSidInt))
	}
	toService := context.Query("to_service")
	if toService == "" {
		logger.Errorf("to_service is empty")
		response.ResultWithoutData(201, "to_service is empty", context)
		return
	}
	toSidStrs := strings.Split(toService, ",")
	var toSids = make([]uint, 0, len(toSidStrs))
	for _, toSidStr := range toSidStrs {
		toSidInt, err := strconv.Atoi(toSidStr)
		if err != nil {
			response.ResultWithoutData(201, fmt.Sprintf("to_service type error: %v", err), context)
			return
		}
		toSids = append(toSids, uint(toSidInt))
	}

	toMigrateHidsStr := context.Query("to_migrate_hids")
	if toMigrateHidsStr == "" {
		logger.Errorf("to_migrate_hids is empty")
		response.ResultWithoutData(201, "to_migrate_hids is empty", context)
		return
	}
	toMigrateHidStrs := strings.Split(toMigrateHidsStr, ",")
	var toMigrateHids = make([]uint, 0, len(toMigrateHidStrs))
	for _, toMigrateHidStr := range toMigrateHidStrs {
		toMigrateHid, err := strconv.Atoi(toMigrateHidStr)
		if err != nil {
			response.ResultWithoutData(201, fmt.Sprintf("to_migrate_hids type error: %v", err), context)
			return
		}
		toMigrateHids = append(toMigrateHids, uint(toMigrateHid))
	}
	err := logics.DoMigrateHosts(fromSids, toSids, toMigrateHids)
	if err != nil {
		logger.Errorf("DoMigrateHosts error: %v", err)
		response.ResultWithoutData(201, fmt.Sprintf("迁移失败，err: %v", err), context)
		return
	}
	response.DoneWithMessage(fmt.Sprintf("%d台机器迁移成功", len(toMigrateHids)), context)
	return
}
