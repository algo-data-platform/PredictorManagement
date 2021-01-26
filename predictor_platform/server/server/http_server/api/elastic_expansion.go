package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"server/libs/logger"
	"server/metrics"
	"server/server/dao"
)

// add info group
func InsertElasticExpansion(context *gin.Context) {
	host_ip := context.Query("host_ip")
	// 校验ip
	address := net.ParseIP(host_ip)
	if address == nil {
		response_str := fmt.Sprintf("ip format error, host_ip: %s", host_ip)
		logger.Warnf(response_str)
		context.String(http.StatusBadRequest, response_str)
		return
	}
	// 判断是否已经存在
	exists, err := dao.ExistsHost(host_ip)
	if err != nil {
		metrics.GetMeters()[metrics.TAG_ELASTIC_EXPANSION_INSERT_ERROR].Mark(1)
		response_str := fmt.Sprintf("get host_ip fail, host_ip: %s, err: %v", host_ip, err)
		logger.Errorf(response_str)
		context.String(http.StatusBadRequest, response_str)
	} else if exists && err == nil {
		logger.Warnf("host_ip is already exists, host_ip: %s", host_ip)
		context.JSON(http.StatusOK, "host_ip is already exists")
	} else {
		err = dao.InsertElasticHost(host_ip)
		if err != nil {
			metrics.GetMeters()[metrics.TAG_ELASTIC_EXPANSION_INSERT_ERROR].Mark(1)
			response_str := fmt.Sprintf("insert host_ip fail, host_ip: %s", host_ip)
			logger.Errorf(response_str)
			context.String(http.StatusBadRequest, response_str)
		} else {
			metrics.GetMeters()[metrics.TAG_ELASTIC_EXPANSION_INSERT].Mark(1)
			logger.Infof("insert host_ip success, host_ip: %s", host_ip)
			context.JSON(http.StatusOK, "insert host_ip success")
		}
	}
}

func DeleteElasticExpansion(context *gin.Context) {
	host_ip := context.Query("host_ip")
	// 校验ip
	address := net.ParseIP(host_ip)
	if address == nil {
		response_str := fmt.Sprintf("ip format error, host_ip: %s", host_ip)
		logger.Warnf(response_str)
		context.String(http.StatusBadRequest, response_str)
		return
	}
	// 获取动态扩容的机器
	host, err := dao.GetElasticHost(host_ip)
	if err != nil {
		metrics.GetMeters()[metrics.TAG_ELASTIC_EXPANSION_DELETE_ERROR].Mark(1)
		response_str := fmt.Sprintf("get host_ip fail, host_ip: %s, err: %v", host_ip, err)
		logger.Errorf(response_str)
		context.String(http.StatusBadRequest, response_str)
	} else if host == nil || (host != nil && host.ID == 0) {
		logger.Warnf("host_ip is already deleted or non-expanded host, host_ip: %s", host_ip)
		context.JSON(http.StatusOK, "host_ip is already deleted or non-expanded host")
	} else {
		// 存在缩容机器
		// 事务删除host_service和hosts表中的host_ip记录
		err = dao.TranscatDeleteHost(host.ID, host_ip)
		if err != nil {
			metrics.GetMeters()[metrics.TAG_ELASTIC_EXPANSION_DELETE_ERROR].Mark(1)
			response_str := fmt.Sprintf("delete host fail, host_ip: %s, hid: %d, err: %v", host_ip, host.ID, err)
			logger.Errorf(response_str)
			context.String(http.StatusBadRequest, response_str)
			return
		} else {
			metrics.GetMeters()[metrics.TAG_ELASTIC_EXPANSION_DELETE].Mark(1)
			logger.Infof("delete host_ip success, host_ip: %s", host_ip)
			context.JSON(http.StatusOK, "delete host_ip success")
		}
	}
}
