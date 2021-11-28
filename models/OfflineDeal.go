package models

import (
	"daobackend/common/constants"
	"daobackend/database"
)

type OfflineDealResult struct {
	Data   []*OfflineDeal `json:"data"`
	Status string         `json:"status"`
	Code   string         `json:"code"`
}

type DealIdList struct {
	DealIdList string `json:"deal_id_list"`
}

type OfflineDeal struct {
	ID                  int64  `json:"id"`
	PayloadCid          string `json:"payload_cid"`
	DealCid             string `json:"deal_cid"`
	DealId              int64  `json:"deal_id"`
	PieceCid            string `json:"piece_cid"`
	MinerFid            string `json:"miner_fid"`
	Duration            int    `json:"duration"`
	Cost                string `json:"cost"`
	CreateAt            string `json:"create_at"`
	Verified            bool   `json:"verified"`
	ClientWalletAddress string `json:"client_wallet_address"`
	SignStatus          string `json:"sign_status"`
	DaoAddress          string `json:"dao_address"`
	TxHash              string `json:"tx_hash"`
}

// FindOfflineDeals (&OfflineDeal{Id: "0xadeaCC802D0f2DFd31bE4Fa7434F15782Fd720ac"},"id desc","10","0")
func FindOfflineDeals(whereCondition interface{}, orderCondition, limit, offset string) ([]*OfflineDeal, error) {
	db := database.GetDB()
	if offset == "" {
		offset = "0"
	}
	if limit == "" {
		limit = constants.DEFAULT_SELECT_LIMIT
	}
	var models []*OfflineDeal
	err := db.Where(whereCondition).Offset(offset).Limit(limit).Order(orderCondition).Find(&models).Error
	return models, err
}
