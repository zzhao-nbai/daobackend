package polygonclient

import (
	"daobackend/config"
	"daobackend/logs"
	"github.com/ethereum/go-ethereum/ethclient"
	"time"
)

type ConnSetup struct {
	ConnWeb *ethclient.Client
}

//setup polygon client connection
var WebConn = new(ConnSetup)

func ClientInit() {
	for {
		rpcUrl := config.GetConfig().RpcUrlPolygon
		client, err := ethclient.Dial(rpcUrl)
		if err != nil {
			logs.GetLogger().Error("Try to reconnect block chain node" + rpcUrl + " ...")
			time.Sleep(time.Second * 5)
		} else {
			WebConn.ConnWeb = client
			break
		}
	}
}
