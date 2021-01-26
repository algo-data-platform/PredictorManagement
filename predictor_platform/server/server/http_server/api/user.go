package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"server/common/response"
	"server/libs/logger"
	"server/server/logics"
)

func UserLogin(context *gin.Context) {
	username := context.Query("username")
	password := context.Query("password")
	logger.Infof("the username is: %s and password is: %s", username, password)
	logger.Infof("user have login in: %s", username)
	isValid := logics.IsValidUser(username, password)
	if isValid {
		response.DoneWithData(map[string]string{"username": username}, context)
		return
	} else {
		response.ResultWithoutData(201, "username and password is not valid, login fail", context)
		return
	}
}

func UserLogout(context *gin.Context) {
	token := context.Query("userToken")
	response_str := fmt.Sprintf("%s logout succ!", token)
	context.String(http.StatusOK, string(response_str))
}
