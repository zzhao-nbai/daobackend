package scheduler

import (
	"context"
	"daobackend/common/httpClient"
	"daobackend/common/utils"
	"daobackend/config"
	"daobackend/database"
	"daobackend/goBind"
	"daobackend/initclient/polygonclient"
	"daobackend/logs"
	"daobackend/models"
	"daobackend/routers/signature"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/robfig/cron"
	"github.com/shopspring/decimal"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"
)

func DaoSignature() {
	c := cron.New()
	err := c.AddFunc(config.GetConfig().DaoSignatureRule, func() {
		logs.GetLogger().Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ dao signature scheduler is running at " + time.Now().Format("2006-01-02 15:04:05"))
		err := DaoSignatureService()
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}
	})
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	c.Start()
}

func DaoSignatureService()error{
	taskStatusWanted := config.GetConfig().TaskStatusDaoWanted
	recipient := common2.HexToAddress(config.GetConfig().Recipient)
	pk1 := os.Getenv("daoOwnerPK1")
	pk2 := os.Getenv("daoOwnerPK2")
	taskList,err := GetTaskListShouldBeSigService()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	for _,v := range taskList.Data.Deals {
		/*if strings.ToLower(v.PayloadCid) != "bafykbzacea4cz7kz77wx6zcqajd3fykb3zwmewiggy4x33wacuwads5favmf2"{
			continue
		}*/

		price, err :=signature.GetFileCoinLastestPriceService()
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		if strings.Contains(strings.ToLower(taskStatusWanted),strings.ToLower(v.Status)) {
			hasBeenSiged,err := CheckIfDealsHasBeenSiged(v)
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}
			if !hasBeenSiged {
				quantity := new(big.Int)
				quantity.SetBytes([]byte(v.Cost))
				cost,err:= decimal.NewFromString(v.Cost)
				if err != nil {
					logs.GetLogger().Error(err)
					continue
				}
				fee := decimal.NewFromFloat(price.Filecoin.Usd)
				finalCost := cost.Mul(fee)
				//todo dao signature 1
				daoWalletAddress1 := common2.HexToAddress("0x05856015d07F3E24936B7D20cB3CcfCa3D34B41d") //pay for gas
				err = doDaoSigOnContract(v.PayloadCid,v.UUID,v.DealCid,finalCost.BigInt(),recipient,true,pk1,daoWalletAddress1)
				if err != nil {
					logs.GetLogger().Error(err)
					continue
				}
				//todo dao signature 2
				daoWalletAddress2 := common2.HexToAddress("0x6f2B76024196e82D81c8bC5eDe7cff0B0276c9C1") //pay for gas
				err = doDaoSigOnContract(v.PayloadCid,v.UUID,v.DealCid,finalCost.BigInt(),recipient,true,pk2,daoWalletAddress2)
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
	if err != nil {
		return err
	}
	return nil
}

func GetTaskListShouldBeSigService()(*models.OfflineDealResult,error){
	url := config.GetConfig().GetTaskFromSwanUrl
	response, err := httpClient.SendRequestAndGetBytes(http.MethodGet, url, nil, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	var results *models.OfflineDealResult
	err = json.Unmarshal(response, &results)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return results, nil
}


func CheckIfDealsHasBeenSiged(deal *models.OfflineDeal)(bool,error){
	dealList, err := models.FindOfflineDeals(&models.OfflineDeal{UUID:deal.UUID,DealCid: deal.DealCid,PayloadCid: deal.PayloadCid,ID: deal.ID},"id desc","10","0")
	if err != nil {
		logs.GetLogger().Error(err)
		return true,err
	}
	if len(dealList) > 0 {
		return true,nil
	}else{
		return false,nil
	}
}

func doDaoSigOnContract(cid string, orderId string, dealId string, paid *big.Int, recipient common2.Address, status bool,privateKeyOfDao string,daoWalletAccount common2.Address) error {
	daoAddress := common2.HexToAddress(config.GetConfig().SwanDaoOralceAddress)
	client := polygonclient.WebConn.ConnWeb
	nonce, err := client.PendingNonceAt(context.Background(), daoWalletAccount)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if strings.HasPrefix(strings.ToLower(privateKeyOfDao), "0x") {
		privateKeyOfDao = privateKeyOfDao[2:]
	}

	privateKey, _ := crypto.HexToECDSA(privateKeyOfDao)
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	callOpts, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainId)

	//callOpts := new(bind.TransactOpts)
	callOpts.Nonce = big.NewInt(int64(nonce))
	callOpts.GasPrice = gasPrice
	callOpts.GasLimit = config.GetConfig().GasLimit
	callOpts.Context = context.Background()

	daoOracleContractInstance, err := goBind.NewFilswanOracle(daoAddress, client)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	tx, err := daoOracleContractInstance.SignTransaction(callOpts, cid , orderId , dealId , paid , recipient , status )
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	logs.GetLogger().Info("dao sig tx hash: ",tx.Hash())
	txRecept, err := utils.CheckTx(client, tx)
	if err != nil {
		logs.GetLogger().Error(err)
	} else {
		if txRecept.Status == uint64(1) {
			logs.GetLogger().Println("dao signature success! txHash=" + tx.Hash().Hex())
		} else {
			logs.GetLogger().Println("dao signature failed! txHash=" + tx.Hash().Hex())
		}
	}

	return nil
}


