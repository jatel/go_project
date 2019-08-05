package tables

type TableOmniSubsendInfo struct {
	Id         uint64 `json:"id"               gorm:"column:id;primary_key;AUTO_INCREMENT"` //自增主键
	Txid       string `json:"txid"		gorm:"column:txid"`
	Propertyid int64  `json:"propertyid"		gorm:"column:propertyid"`
	Divisible  int8   `json:"divisible"		gorm:"column:divisible"`
	Amount     string `json:"amount"		gorm:"column:amount"`
}

func (t *TableOmniSubsendInfo) TableName() string {
	return "t_omni_subsend_info"
}
