package main

import (
	"daobackend/common/constants"
	"daobackend/config"
	"daobackend/database"
	"daobackend/initclient/polygonclient"
	"daobackend/logs"
	"daobackend/routers/common"
	"daobackend/routers/signature"
	"daobackend/scheduler"
	"fmt"
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/joho/godotenv"
	"os"
	"time"
)

func main() {
	LoadEnv()

	polygonclient.ClientInit()

	// init database
	db := database.Init()
	scheduler.DaoSignatureService()
	//scheduler.DaoSignature()

	defer func() {
		err := db.Close()
		if err != nil {
			logs.GetLogger().Error(err)
		}
	}()

	r := gin.Default()
	r.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     true,
		ValidateHeaders: false,
	}))

	v1 := r.Group("/api/v1")
	common.HostManager(v1.Group(constants.URL_HOST_GET_COMMON))
	signature.SignatureManager(v1.Group(constants.URL_EVENT_SIGNATURE))

	err := r.Run(":" + config.GetConfig().Port)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		logs.GetLogger().Error(err)
	}
	fmt.Println("name: ", os.Getenv("privateKey"))
}
