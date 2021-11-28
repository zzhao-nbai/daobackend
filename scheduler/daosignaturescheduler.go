package scheduler

import (
	"context"
	"daobackend/common/constants"
	"daobackend/common/httpClient"
	"daobackend/common/utils"
	"daobackend/config"
	"daobackend/database"
	"daobackend/goBind"
	"daobackend/initclient/polygonclient"
	"daobackend/logs"
	"daobackend/models"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/robfig/cron"
	"math/big"
	"net/http"
	"os"
	"strconv"
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

func DaoSignatureService() error {
	recipient := common2.HexToAddress(config.GetConfig().Recipient)
	pk1 := os.Getenv("daoOwnerPK1")
	pk2 := os.Getenv("daoOwnerPK2")
	pk3 := os.Getenv("daoOwnerPK3")
	taskList, err := GetTaskListShouldBeSigService()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	idListStringValue := ""
	for _, v := range taskList.Data {
		/*if strings.ToLower(v.PayloadCid) != "bafykbzacea4cz7kz77wx6zcqajd3fykb3zwmewiggy4x33wacuwads5favmf2"{
			continue
		}*/
		idListStringValue = idListStringValue + "," + strconv.FormatInt(v.DealId, 10)
		signStatus := constants.DAO_SIGN_STATUS_SUCCESS
		//todo dao signature 1
		daoWalletAddress1 := common2.HexToAddress("0x05856015d07F3E24936B7D20cB3CcfCa3D34B41d") //pay for gas
		hasBeenSiged1, err := CheckIfDealsHasBeenSigned(v, daoWalletAddress1.Hex())
		if err != nil {
			logs.GetLogger().Error(err)
			hasBeenSiged1 = true
		}
		if !hasBeenSiged1 {
			txHash1, err := doDaoSigOnContract(v.PayloadCid, v.DealCid, recipient, pk1, daoWalletAddress1)
			if err != nil {
				signStatus = constants.DAO_SIGN_STATUS_FAIL
				logs.GetLogger().Error(err)
			}
			err = SaveDealInfoToDB(v, signStatus, daoWalletAddress1.Hex(), txHash1)
			if err != nil {
				logs.GetLogger().Error(err)
			}
		}

		//todo dao signature 2
		daoWalletAddress2 := common2.HexToAddress("0x6f2B76024196e82D81c8bC5eDe7cff0B0276c9C1") //pay for gas
		hasBeenSiged2, err := CheckIfDealsHasBeenSigned(v, daoWalletAddress2.Hex())
		if err != nil {
			logs.GetLogger().Error(err)
			hasBeenSiged2 = true
		}
		if !hasBeenSiged2 {
			txHash2, err := doDaoSigOnContract(v.PayloadCid, v.DealCid, recipient, pk2, daoWalletAddress2)
			if err != nil {
				signStatus = constants.DAO_SIGN_STATUS_FAIL
				logs.GetLogger().Error(err)
			}
			err = SaveDealInfoToDB(v, signStatus, daoWalletAddress2.Hex(), txHash2)
			if err != nil {
				logs.GetLogger().Error(err)
			}
		}

		//todo dao signature 3
		daoWalletAddress3 := common2.HexToAddress("0x800210CfB747992790245eA878D32F188d01a03A") //pay for gas
		hasBeenSiged3, err := CheckIfDealsHasBeenSigned(v, daoWalletAddress3.Hex())
		if err != nil {
			logs.GetLogger().Error(err)
			hasBeenSiged3 = true
		}
		if !hasBeenSiged3 {
			txHash3, err := doDaoSigOnContract(v.PayloadCid, v.DealCid, recipient, pk3, daoWalletAddress3)
			if err != nil {
				signStatus = constants.DAO_SIGN_STATUS_FAIL
				logs.GetLogger().Error(err)
			}
			err = SaveDealInfoToDB(v, signStatus, daoWalletAddress3.Hex(), txHash3)
			if err != nil {
				logs.GetLogger().Error(err)
			}
		}
	}
	if strings.Trim(idListStringValue, " ") != "" {
		err = UpdateSgnedDealInfoToPaymentGateway(idListStringValue)
		if err != nil {
			logs.GetLogger().Error(err)
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func SaveDealInfoToDB(deal *models.OfflineDeal, signStatus, daoAddress, txHash string) error {
	offlineDeal := new(models.OfflineDeal)
	offlineDeal.DealId = deal.DealId
	offlineDeal.PayloadCid = deal.PayloadCid
	offlineDeal.DealCid = deal.DealCid
	offlineDeal.PieceCid = deal.PieceCid
	offlineDeal.MinerFid = deal.MinerFid
	offlineDeal.Duration = deal.Duration
	offlineDeal.Cost = deal.Cost
	offlineDeal.CreateAt = deal.CreateAt
	offlineDeal.Verified = deal.Verified
	offlineDeal.ClientWalletAddress = deal.ClientWalletAddress
	offlineDeal.SignStatus = signStatus
	offlineDeal.DaoAddress = daoAddress
	offlineDeal.TxHash = txHash
	err := database.SaveOne(offlineDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	return nil
}

func UpdateSgnedDealInfoToPaymentGateway(dealIdList string) error {
	url := config.GetConfig().UpdateDealsToPaymentGatewayUrl
	idList := new(models.DealIdList)
	idList.DealIdList = dealIdList
	paramBytes, err := json.Marshal(idList)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	_, err = httpClient.SendRequestAndGetBytes(http.MethodPut, url, paramBytes, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func GetTaskListShouldBeSigService() (*models.OfflineDealResult, error) {
	url := config.GetConfig().GetDealsFromPaymentGatewayUrl
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

func CheckIfDealsHasBeenSigned(deal *models.OfflineDeal, daoWalletAddress string) (bool, error) {
	dealList, err := models.FindOfflineDeals(&models.OfflineDeal{DealId: deal.DealId, DaoAddress: daoWalletAddress}, "id desc", "10", "0")
	if err != nil {
		logs.GetLogger().Error(err)
		return true, err
	}
	if len(dealList) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func doDaoSigOnContract(cid string, dealId string, recipient common2.Address, privateKeyOfDao string, daoWalletAccount common2.Address) (string, error) {
	daoAddress := common2.HexToAddress(config.GetConfig().SwanDaoOralceAddress)
	client := polygonclient.WebConn.ConnWeb
	nonce, err := client.PendingNonceAt(context.Background(), daoWalletAccount)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}

	if strings.HasPrefix(strings.ToLower(privateKeyOfDao), "0x") {
		privateKeyOfDao = privateKeyOfDao[2:]
	}

	privateKey, _ := crypto.HexToECDSA(privateKeyOfDao)
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
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
		return "", err
	}

	tx, err := daoOracleContractInstance.SignTransaction(callOpts, cid, dealId, recipient)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	logs.GetLogger().Info("dao sig tx hash: ", tx.Hash())
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

	return tx.Hash().Hex(), err
}
