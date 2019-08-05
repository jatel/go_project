package tables

type TableOmniTradeInfo struct {
	Id             uint64 `json:"id"               gorm:"column:id;primary_key;AUTO_INCREMENT"` //自增主键
	Txid           string `json:"txid"		gorm:"column:txid"`
	Hash           string `json:"hash"		gorm:"column:hash"`
	Block          int    `json:"block"		gorm:"column:block"`
	Address        string `json:"address"		gorm:"column:address"`
	Amountsold     string `json:"amountsold"		gorm:"column:amountsold"`
	Amountreceived string `json:"amountreceived"		gorm:"column:amountreceived"`
	Tradingfee     string `json:"tradingfee"		gorm:"column:tradingfee"`
}

func (t *TableOmniTradeInfo) TableName() string {
	return "t_omni_trade_info"
}
