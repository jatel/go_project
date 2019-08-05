package tables

type TableOmniCancelInfo struct {
	Id               uint64 `json:"id"               gorm:"column:id;primary_key;AUTO_INCREMENT"` //自增主键
	Hash             string `json:"hash"		gorm:"column:hash"`
	Txid             string `json:"txid"		gorm:"column:txid"`
	Propertyid       int64  `json:"propertyid"		gorm:"column:propertyid"`
	Amountunreserved string `json:"amountunreserved"		gorm:"column:amountunreserved"`
}

func (t *TableOmniCancelInfo) TableName() string {
	return "t_omni_cancel_info"
}
