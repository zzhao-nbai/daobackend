package signature

import (
	"daobackend/common"
	"daobackend/common/constants"
	"daobackend/common/errorinfo"
	"daobackend/config"
	"daobackend/database"
	"daobackend/logs"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	"os"
	"strings"
)

func SignatureManager(router *gin.RouterGroup) {
	router.GET(constants.URL_DAO_SIGNATURE, GetTaskListShouldBeSig)
}

func GetTaskListShouldBeSig(c *gin.Context) {
	taskStatusWanted := config.GetConfig().TaskStatusDaoWanted
	recipient := common2.HexToAddress(config.GetConfig().Recipient)
	pk1 := os.Getenv("daoOwnerPK1")
	pk2 := os.Getenv("daoOwnerPK2")
	paid := big.NewInt(1000000000000000)
	terms := big.NewInt(2000000000000000)
	taskList,err := GetTaskListShouldBeSigService()
	if err != nil {
		logs.GetLogger().Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, common.CreateErrorResponse(errorinfo.GET_EVENT_FROM_DB_ERROR_CODE, errorinfo.GET_EVENT_FROM_DB_ERROR_MSG))
		return
	}

	for _,v := range taskList.Data.Deals {
		if strings.ToLower(v.PayloadCid) != "bafykbzacea4cz7kz77wx6zcqajd3fykb3zwmewiggy4x33wacuwads5favmf2"{
			continue
		}
		if   strings.Contains(strings.ToLower(taskStatusWanted),strings.ToLower(v.Status)) {
			hasBeenSiged,err := CheckIfDealsHasBeenSiged(v)
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}
			if !hasBeenSiged {
				//todo dao signature 1
				daoWalletAddress1 := common2.HexToAddress("0x05856015d07F3E24936B7D20cB3CcfCa3D34B41d") //pay for gas
				err = doDaoSigOnContract(v.PayloadCid,v.UUID,v.DealCid,paid,recipient,terms,true,pk1,daoWalletAddress1)
				if err != nil {
					logs.GetLogger().Error(err)
					continue
				}
				//todo dao signature 2
				daoWalletAddress2 := common2.HexToAddress("0x6f2B76024196e82D81c8bC5eDe7cff0B0276c9C1") //pay for gas
				err = doDaoSigOnContract(v.PayloadCid,v.UUID,v.DealCid,paid,recipient,terms,true,pk2,daoWalletAddress2)
				if err != nil {
					logs.GetLogger().Error(err)
					continue
				}
				err = database.SaveOne(v)
				if err != nil {
					logs.GetLogger().Error(err)
					continue
				}
			}
		}
	}
	c.JSON(http.StatusOK, common.CreateSuccessResponse(taskList))
}
