package models

type EventPolygon struct {
	ID              int64  `json:"id"`
	TxHash          string `json:"tx_hash"`
	EventName       string `json:"event_name"`
	PayloadCid      string `json:"payload_cid"`
	TokenAddress    string `json:"token_address"`
	MinPayment      string `json:"min_payment"`
	ContractAddress string `json:"contract_address"`
	LockedFee       string `json:"locked_fee"`
	Deadline        string `json:"deadline"`
	BlockNo         uint64 `json:"block_no"`
	MinerAddress    string `json:"miner_address"`
	AddressFrom     string `json:"address_from"`
	AddressTo       string `json:"address_to"`
	CoinType        string `json:"coin_type"`
	LockPaymentTime string `json:"lock_payment_time"`
	CreateAt        string `json:"create_at"`
}
