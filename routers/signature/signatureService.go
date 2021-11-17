package signature

import (
	"context"
	"daobackend/common/httpClient"
	"daobackend/config"
	"daobackend/goBind"
	"daobackend/initclient/polygonclient"
	"daobackend/logs"
	"daobackend/models"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"net/http"
	"strings"
)

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

func doDaoSigOnContract(cid string, orderId string, dealId string, paid *big.Int, recipient common.Address, terms *big.Int, status bool,privateKeyOfDao string,daoWalletAccount common.Address) error {
	daoAddress := common.HexToAddress(config.GetConfig().SwanDaoOralceAddress)
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
	return nil
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

func GetFileCoinLastestPriceService() (*models.PriceResult, error) {
	url := "https://api.coingecko.com/api/v3/simple/price?ids=filecoin&vs_currencies=usd"
	response, err := httpClient.SendRequestAndGetBytes(http.MethodGet, url, nil, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	var price *models.PriceResult
	err = json.Unmarshal(response, &price)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return price, nil
}
