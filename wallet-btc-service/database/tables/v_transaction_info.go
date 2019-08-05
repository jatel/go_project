package tables

type ViewTransactionInfo struct {
	Iscoinbase      int8   `json:"iscoinbase"        	gorm:"column:iscoinbase"`                      //是否为coinbase，0表示普通交易，1表示为coinbase
	Time            int64  `json:"time"      			gorm:"column:time"`                                  //区块打包时间
	Blockhash       string `json:"blockhash"         	gorm:"column:blockhash;type:char(64)"`         //当前交易所在区块的哈希
	Blockheight     int32  `json:"blockheight"           gorm:"column:blockheight"`                  //区块高度
	Transactionhash string `json:"transactionhash"      gorm:"column:transactionhash;type:char(64)"` //交易哈希
	Fromhash        string `json:"fromhash"             gorm:"column:fromhash;type:char(64)"`        //交易id
	Fromindex       int64  `json:"fromindex"            gorm:"column:fromindex"`                     //vout索引号
	Coinbase        string `json:"coinbase"         	gorm:"column:coinbase;type:text"`               //coinbase的scriptSig信息
	Fromaddress     string `json:"fromaddress"          gorm:"column:fromaddress;type:varchar(64)"`  //转出账户地址
	Fromvalue       int64  `json:"fromvalue"            gorm:"column:fromvalue"`                     //转账金额
	Tovalue         int64  `json:"tovalue"              gorm:"column:tovalue"`                       //转账金额
	Toindex         int64  `json:"toindex"          	gorm:"column:toindex"`                          //索引号
	Toasm           string `json:"toasm"         		gorm:"column:toasm;type:text"`                    //公钥哈希脚本的asm表示
	Totype          string `json:"totype"         		gorm:"column:totype;type:var char(64)"`          //公钥哈希脚本的类型
	Toaddress       string `json:"toaddress"            gorm:"column:toaddress;type:varchar(64)"`    //转入账户地址
	State           int8   `json:"state"        		gorm:"column:state"`                               //当前花费状态，0表示未花费，1表示已花费，2表示分叉链被回退
}

func (t *ViewTransactionInfo) TableName() string {
	return "v_transaction_info"
}
