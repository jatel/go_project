package tables

type TableOmniRecipientInfo struct {
	Id      uint64 `json:"id"               gorm:"column:id;primary_key;AUTO_INCREMENT"` //自增主键
	Txid    string `json:"txid"		gorm:"column:txid"`
	Address string `json:"address" 		gorm:"column:address"`
	Amount  string `json:"amount" 		gorm:"column:amount"`
}

func (t *TableOmniRecipientInfo) TableName() string {
	return "t_omni_recipient_info"
}
