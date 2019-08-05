package tables

type TableInputInfo struct {
	Id        int64  `json:"id"                	gorm:"column:id;primary_key;AUTO_INCREMENT"` //自增主键
	Blockhash string `json:"blockhash"         	gorm:"column:blockhash;type:char(64)"`       //当前交易所在区块的哈希
	Time      int64  `json:"time"      			gorm:"column:time"`                                //区块打包时间
	Isfork    int8   `json:"isfork"        		gorm:"column:isfork"`                           //是否为分叉链，0表示为主链，1表示为分叉链
	Hash      string `json:"hash"         		gorm:"column:hash;type:char(64)"`                //所在交易哈希
	Txid      string `json:"txid"              	gorm:"column:txid;type:char(64)"`            //交易id
	Vout      int64  `json:"vout"               gorm:"column:vout"`                          //vout索引号
	Sequence  int64  `json:"sequence"          	gorm:"column:sequence"`                      //脚本序列号
	Hex       string `json:"hex"         		gorm:"column:hex;type:text"`                      //交易的签名信息脚本的十六进制表示
	Asm       string `json:"asm"         		gorm:"column:asm;type:text"`                      //交易的签名信息脚本的asm码表示
	Coinbase  string `json:"coinbase"         	gorm:"column:coinbase;type:text"`             //coinbase的scriptSig信息
	From      string `json:"from"             	gorm:"column:from;type:varchar(64)"`          //转出账户地址
	Value     int64  `json:"value"              gorm:"column:value"`                         //转账金额
}

func (t *TableInputInfo) TableName() string {
	return "t_input_info"
}
