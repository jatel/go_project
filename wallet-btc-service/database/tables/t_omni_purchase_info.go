package tables

type TableOmniPurchaseInfo struct {
	Id               uint64 `json:"id"               gorm:"column:id;primary_key;AUTO_INCREMENT"` //自增主键
	Txid             string `json:"txid"		gorm:"column:txid"`
	Vout             int64  `json:"vout" 		gorm:"column:vout"`
	Amountpaid       string `json:"amountpaid" 		gorm:"column:amountpaid"`
	Ismine           int8   `json:"ismine" 		gorm:"column:ismine"`
	Referenceaddress string `json:"referenceaddress" 		gorm:"column:referenceaddress"`
	Propertyid       int64  `json:"propertyid" 		gorm:"column:propertyid"`
	Amountbought     string `json:"amountbought" 		gorm:"column:amountbought"`
	Valid            int8   `json:"valid" 		gorm:"column:valid"`
}

func (t *TableOmniPurchaseInfo) TableName() string {
	return "t_omni_purchase_info"
}
