package models

type DaoEventLog struct {
	ID                    int64  `json:"id"`
	TxHash                string `json:"tx_hash"`
	Recipient             string `json:"recipient"`
	PayloadCid            string `json:"payload_cid"`
	OrderId               string `json:"order_id"`
	DealCid               string `json:"deal_cid"`
	Terms                 string `json:"terms"`
	Cost                  string `json:"cost"`
	DaoAddress            string `json:"dao_address"`
	DaoPassTime           string `json:"dao_pass_time"`
	BlockNo               uint64 `json:"block_no"`
	BlockTime             string `json:"block_time"`
	Network               string `json:"network"`
	Status                bool   `json:"status"`
	SignatureUnlockStatus string `json:"signature_unlock_status"`
}

type DaoSignatureResult struct {
	Recipient  string `json:"recipient"`
	PayloadCid string `json:"payload_cid"`
	OrderId    string `json:"order_id"`
	DealCid    string `json:"deal_cid"`
	Threshold  string `json:"threshold"`
}
