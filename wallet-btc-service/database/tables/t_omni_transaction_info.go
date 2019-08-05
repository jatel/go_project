package tables

type TableOmniTransactionInfo struct {
	Id                           uint64 `json:"id"               gorm:"column:id;primary_key;AUTO_INCREMENT"` //自增主键
	Txid                         string `json:"txid"		gorm:"column:txid"`
	Blockhash                    string `json:"blockhash" 		gorm:"column:blockhash"`
	Blocktime                    int64  `json:"blocktime"		gorm:"column:blocktime"`
	Block                        int32  `json:"block"		gorm:"column:block"`
	Positioninblock              int32  `json:"positioninblock"		gorm:"column:positioninblock"`
	Fee                          string `json:"fee"		gorm:"column:fee"`
	Sendingaddress               string `json:"sendingaddress"		gorm:"column:sendingaddress"`
	Referenceaddress             string `json:"referenceaddress"		gorm:"column:referenceaddress"`
	Type_int                     uint64 `json:"type_int"		gorm:"column:type_int"`
	Type                         string `json:"type"		gorm:"column:type"`
	Ismine                       int8   `json:"ismine"		gorm:"column:ismine"`
	Version                      uint64 `json:"version"		gorm:"column:version"`
	Valid                        int8   `json:"valid"		gorm:"column:valid"`
	Invalidreason                string `json:"invalidreason"		gorm:"column:invalidreason"`
	Purchasedpropertyid          int64  `json:"purchasedpropertyid"		gorm:"column:purchasedpropertyid"`
	Purchasedpropertyname        string `json:"purchasedpropertyname"		gorm:"column:purchasedpropertyname"`
	Purchasedpropertydivisible   int8   `json:"purchasedpropertydivisible"		gorm:"column:purchasedpropertydivisible"`
	Purchasedtokens              string `json:"purchasedtokens"		gorm:"column:purchasedtokens"`
	Issuertokens                 string `json:"issuertokens"		gorm:"column:issuertokens"`
	Propertyid                   uint64 `json:"propertyid"		gorm:"column:propertyid"`
	Divisible                    int8   `json:"divisible"		gorm:"column:divisible"`
	Amount                       string `json:"amount"		gorm:"column:amount"`
	Totalstofee                  string `json:"totalstofee"		gorm:"column:totalstofee"`
	Ecosystem                    string `json:"ecosystem"		gorm:"column:ecosystem"`
	Bitcoindesired               string `json:"bitcoindesired"		gorm:"column:bitcoindesired"`
	Timelimit                    uint8  `json:"timelimit"		gorm:"column:timelimit"`
	Feerequired                  string `json:"feerequired"		gorm:"column:feerequired"`
	Action                       string `json:"action"		gorm:"column:action"`
	Propertyidforsale            uint64 `json:"propertyidforsale"		gorm:"column:propertyidforsale"`
	Propertyidforsaleisdivisible int8   `json:"propertyidforsaleisdivisible"		gorm:"column:propertyidforsaleisdivisible"`
	Amountforsale                string `json:"amountforsale"		gorm:"column:amountforsale"`
	Propertyiddesired            uint64 `json:"propertyiddesired"		gorm:"column:propertyiddesired"`
	Propertyiddesiredisdivisible int8   `json:"propertyiddesiredisdivisible"		gorm:"column:propertyiddesiredisdivisible"`
	Amountdesired                string `json:"amountdesired"		gorm:"column:amountdesired"`
	Unitprice                    string `json:"unitprice"		gorm:"column:unitprice"`
	Amountremaining              string `json:"amountremaining"		gorm:"column:amountremaining"`
	Amounttofill                 string `json:"amounttofill"		gorm:"column:amounttofill"`
	Status                       string `json:"status"		gorm:"column:status"`
	Canceltxid                   string `json:"canceltxid"		gorm:"column:canceltxid"`
	Propertytype                 string `json:"propertytype"		gorm:"column:propertytype"`
	Category                     string `json:"category"		gorm:"column:category"`
	Subcategory                  string `json:"subcategory"		gorm:"column:subcategory"`
	Propertyname                 string `json:"propertyname"		gorm:"column:propertyname"`
	Data                         string `json:"data"		gorm:"column:data"`
	Url                          string `json:"url"		gorm:"column:url"`
	Tokensperunit                string `json:"tokensperunit"		gorm:"column:tokensperunit"`
	Deadline                     int64  `json:"deadline"		gorm:"column:deadline"`
	Earlybonus                   uint8  `json:"earlybonus"		gorm:"column:earlybonus"`
	Percenttoissuer              uint8  `json:"percenttoissuer"		gorm:"column:percenttoissuer"`
	Featureid                    uint64 `json:"featureid"		gorm:"column:featureid"`
	Activationblock              uint64 `json:"activationblock"		gorm:"column:activationblock"`
	Minimumversion               uint64 `json:"minimumversion"		gorm:"column:minimumversion"`
}

func (t *TableOmniTransactionInfo) TableName() string {
	return "t_omni_transaction_info"
}
