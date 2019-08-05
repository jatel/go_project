package tables

type TableBlockInfo struct {
	Id                uint64  `json:"id"               gorm:"column:id;primary_key;AUTO_INCREMENT"`            //自增主键
	Height            int32   `json:"height"           gorm:"column:height"`                                   //区块高度
	Isfork            int8    `json:"isfork"        gorm:"column:isfork"`                                      //是否为分叉链，0表示为主链，1表示为分叉链
	Version           int     `json:"version"        	gorm:"column:version;"`                                  //版本号
	Time              int64   `json:"time"        		gorm:"column:time"`                                        //区块打包时间
	Bits              string  `json:"bits"          	gorm:"column:bits;type:varchar(64)"`                      //难度对应值
	Nonce             int64   `json:"nonce"            gorm:"column:nonce"`                                    //随机数
	Difficulty        float64 `json:"difficulty"       gorm:"column:difficulty"`                               //难度值
	Size              int32   `json:"size"             gorm:"column:size"`                                     //区块大小
	Weight            int64   `json:"weight"          	gorm:"column:weight"`                                   //权重
	Mediantime        int64   `json:"mediantime"       gorm:"column:mediantime"`                               //时间
	Chainwork         string  `json:"chainwork"       	gorm:"column:chainwork;type:varchar(64)"`               //哈希次数
	Hash              string  `json:"hash"     		gorm:"column:hash;type:char(64);unique_index"`                //当前区块的哈希
	Merkleroot        string  `json:"merkleroot"       gorm:"column:merkleroot;type:char(64)"`                 //默克尔树根
	Previousblockhash string  `json:"previousblockhash"         gorm:"column:previousblockhash;type:char(64)"` //前一个区块哈希
	Ntx               int64   `json:"ntx"         		gorm:"column:ntx"`                                         //交易个数
}

func (t *TableBlockInfo) TableName() string {
	return "t_block_info"
}
