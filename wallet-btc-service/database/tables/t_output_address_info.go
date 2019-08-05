package tables

type TableOutputAddressInfo struct {
	Id        int64  `json:"id"                	gorm:"column:id;primary_key;AUTO_INCREMENT"` //自增主键
	Blockhash string `json:"blockhash"         	gorm:"column:blockhash;type:char(64)"`       //当前交易所在区块的哈希
	Hash      string `json:"hash"         		gorm:"column:hash;type:char(64)"`                //所在交易哈希
	N         int64  `json:"n"          		gorm:"column:n"`                                   //索引号
	Address   string `json:"address"         	gorm:"column:address;type:varchar(64)"`        //公钥哈希脚本的十六进制表示
}

func (t *TableOutputAddressInfo) TableName() string {
	return "t_output_address_info"
}
