package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"server/env"
	"server/server/logics"
	"strings"
)

func ModelInfoModelHistory(context *gin.Context) {
	model_name := context.Query("modelname")
	show_count := context.Query("number")
	model_history_info := logics.GetModelInfo(model_name, show_count, env.Env.Conf)
	context.JSON(http.StatusOK, model_history_info)
}

func ModelInfoUpdateIntervalWeek(context *gin.Context) {
	model_list := context.Query("model_list")
	modelList := strings.Split(model_list, ",")
	update_interval_weekly, _ := logics.ModelListUpdateTimeWithinWeek(modelList, env.Env.Conf)
	context.JSON(http.StatusOK, update_interval_weekly)
}

func ModelInfoModelsMailRecipients(context *gin.Context) {
	model_list := context.Query("model_list")
	models_mail_recipients := logics.GetModelsMailRecipients(model_list)
	context.JSON(http.StatusOK, models_mail_recipients)
}

func ModelInfoSetModelMailRecipients(context *gin.Context) {
	model_name := context.Query("model_name")
	mail_recipients := context.Query("mail_recipients")
	logics.SetModelMailRecipients(model_name, mail_recipients, env.Env.Conf)
	context.JSON(http.StatusOK, "set model mail recipients ok")
}
