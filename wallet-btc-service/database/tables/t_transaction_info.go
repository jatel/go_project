package tables

type TableTransactionInfo struct {
	Id          int64  `json:"id"                	gorm:"column:id;primary_key;AUTO_INCREMENT"` //自增主键
	Blockhash   string `json:"blockhash"         	gorm:"column:blockhash;type:char(64)"`       //当前交易所在区块的哈希
	Blockheight int32  `json:"blockheight"           gorm:"column:blockheight"`                //区块高度
	Time        int64  `json:"time"      			gorm:"column:time"`                                //区块打包时间
	Isfork      int8   `json:"isfork"        		gorm:"column:isfork"`                           //是否为分叉链，0表示为主链，1表示为分叉链
	Txid        string `json:"txid"              	gorm:"column:txid;type:char(64)"`            //交易哈希
	Iscoinbase  int8   `json:"iscoinbase"        	gorm:"column:iscoinbase"`                    //是否为coinbase，0表示普通交易，1表示为coinbase
	Hash        string `json:"hash"               	gorm:"column:hash;type:char(64)"`           //Witness哈希
	Size        int32  `json:"size"               	gorm:"column:size"`                         //交易大小
	Vsize       int64  `json:"vsize"              	gorm:"column:vsize"`                        //加权大小
	Version     int32  `json:"version"          		gorm:"column:version"`                       //版本号
	Locktime    int64  `json:"locktime"         		gorm:"column:locktime"`                      //锁定时间
	Weight      int64  `json:"weight"             	gorm:"column:weight"`                       //权重
}

func (t *TableTransactionInfo) TableName() string {
	return "t_transaction_info"
}
