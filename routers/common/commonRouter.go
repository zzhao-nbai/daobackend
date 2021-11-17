package common

import (
	"daobackend/common"
	"daobackend/common/constants"
	"github.com/gin-gonic/gin"
	"net/http"
)

func HostManager(router *gin.RouterGroup) {
	router.GET(constants.URL_HOST_GET_HOST_INFO, GetSwanMinerVersion)
}

func GetSwanMinerVersion(c *gin.Context) {
	info := getHostInfo()
	c.JSON(http.StatusOK, common.CreateSuccessResponse(info))
}
