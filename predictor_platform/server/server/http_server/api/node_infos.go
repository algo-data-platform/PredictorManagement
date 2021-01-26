package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"server/common"
	"server/libs/logger"
)

func NodeInfos(context *gin.Context) {
	// nodeInfo 放到初始化和monitor 中定时更新，提高响应速度及重复调用
	json_node_info, err := json.Marshal(common.GNodeInfos)
	if err != nil {
		logger.Errorf("node_infos err: %v", err)
		context.String(http.StatusForbidden, "marshal json error")
	} else {
		node_info_str := string(json_node_info)
		context.String(http.StatusOK, node_info_str)
	}
}
