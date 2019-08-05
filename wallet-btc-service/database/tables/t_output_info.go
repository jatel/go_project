package tables

type TableOutputInfo struct {
	Id        int64  `json:"id"                	gorm:"column:id;primary_key;AUTO_INCREMENT"` //自增主键
	Blockhash string `json:"blockhash"         	gorm:"column:blockhash;type:char(64)"`       //当前交易所在区块的哈希
	Time      int64  `json:"time"      			gorm:"column:time"`                                //区块打包时间
	Isfork    int8   `json:"isfork"        		gorm:"column:isfork"`                           //是否为分叉链，0表示为主链，1表示为分叉链
	Hash      string `json:"hash"         		gorm:"column:hash;type:char(64)"`                //所在交易哈希
	Value     int64  `json:"value"              gorm:"column:value"`                         //转账金额
	N         int64  `json:"n"          		gorm:"column:n"`                                   //索引号
	Hex       string `json:"hex"         		gorm:"column:hex;type:text"`                      //公钥哈希脚本的十六进制表示
	Asm       string `json:"asm"         		gorm:"column:asm;type:text"`                      //公钥哈希脚本的asm表示
	Type      string `json:"type"         		gorm:"column:type;type:var char(64)"`            //公钥哈希脚本的类型
	ReqSigs   int32  `json:"reqSigs"            gorm:"column:reqSigs"`                       //需要签名的个数
	To        string `json:"to"             	gorm:"column:to;type:varchar(64)"`              //转入账户地址
	State     int8   `json:"state"        		gorm:"column:state"`                             //当前花费状态，0表示未花费，1表示已花费，2表示分叉链被回退
}

func (t *TableOutputInfo) TableName() string {
	return "t_output_info"
}
