package tables

type TableTransactionInputOutputAddressInfo struct {
	Id        int64  `json:"id"                	gorm:"column:id;primary_key;AUTO_INCREMENT"` //自增主键
	Blockhash string `json:"blockhash"         	gorm:"column:blockhash;type:char(64)"`       //当前交易所在区块的哈希
	Time      int64  `json:"time"      			gorm:"column:time"`                                //区块打包时间
	Txid      string `json:"txid"         		gorm:"column:txid;type:char(64)"`                //所在交易哈希
	Index     int64  `json:"index"          	gorm:"column:index"`                            //索引号
	Address   string `json:"address"         	gorm:"column:address;type:varchar(64)"`        //地址
	Isfrom    int8   `json:"isfrom"        		gorm:"column:isfrom"`                           //是否为input，0表示output，1表示为input
}

func (t *TableTransactionInputOutputAddressInfo) TableName() string {
	return "t_transaction_input_output_address_info"
}
