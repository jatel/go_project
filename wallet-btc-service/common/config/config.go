package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/spf13/viper"
)

var (
	Cfg *Config

	// It doesn't have to be accurate and non-thread-safe
	TrxGoroutineRuntime int
)

type BtcOpt struct {
	RpcUser             string
	RpcPassword         string
	RpcPort             int
	RpcAddress          string
	ListenAddress       string
	ApiServerAddress    string
	NotifyServerAddress string

	// bitcoin
	RepairBlock          string
	RepairTransaction    string
	RepairUnconfirmedTrx string
	RepairAll            string

	// omni
	RepairOmniBlock          string
	RepairOmniTransaction    string
	RepairUnconfirmedOmniTrx string
	RepairOmniAll            string

	ModifyGoroutine   string
	TrxGoroutineNum   int
	BlockGoroutineNum int
	MaxGoroutine      int
}

type OmniOpt struct {
	RpcUser     string
	RpcPassword string
	RpcPort     int
	RpcAddress  string
}

type DbOpt struct {
	Address        string `json:"address"`
	User           string `json:"user"`
	Password       string `json:"password"`
	DbName         string `json:"db_name"`
	MaxOpenConn    int    `json:"max_open_conn"`
	MaxIdleConn    int    `json:"max_idle_conn"`
	MaxWaitTimeout int    `json:"max_wait_timeout"`
}

type Number struct {
	BlockNum int     `json:"blocknum"`
	Normal   float64 `json:"normal"`
	Priority float64 `json:"priority"`
	Quick    float64 `json:"quick"`
}

type RedisOpt struct {
	RedisAddress string `json:"redis_address"`
	RedisDbNum   int    `json:"redis_db_number"`
}

type Config struct {
	BtcOpt   BtcOpt   `json:"btc_opt"`
	OmniOpt  OmniOpt  `json:"omni_opt"`
	DbOpt    DbOpt    `json:"db_opt"`
	RedisOpt RedisOpt `json:"redis_opt"`
	Number   Number   `json:"number"`
}

func init() {
	var err error
	Cfg, err = Initialize()
	if err != nil {
		panic(err)
	}

	TrxGoroutineRuntime = Cfg.BtcOpt.TrxGoroutineNum
}

func Initialize() (*Config, error) {
	var BaseDir string
	if flag.Lookup("test.v") == nil {
		BaseDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		BaseDir = strings.Replace(BaseDir, "\\", "/", -1)
	} else {
		BaseDir = "/tmp"
	}
	Cfg = &Config{}
	if err := Cfg.ReadConfigFile(BaseDir + "/wallet-btc-service.toml"); err != nil {
		return nil, err
	}

	if err := log.Initialize(BaseDir + "/log/"); err != nil {
		return nil, err
	}
	Cfg.readFileConfig()
	log.Log.Debug("Initialize config module success:", Cfg)
	return Cfg, nil
}

//读取配置文件，如果文件不存在，则创建一个文件
func (c *Config) ReadConfigFile(file string) error {
	viper.SetConfigFile(file)

	log.InitDefaultConfig() //set log config
	c.initDefaultConfig()

	fmt.Println("config file:", viper.ConfigFileUsed())
	if _, err := os.Stat(viper.ConfigFileUsed()); os.IsNotExist(err) {
		if _, err := os.Create(viper.ConfigFileUsed()); err != nil {
			fmt.Println("create file err:", err)
			return err
		}
		fmt.Println("write config into file")
		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

//此处设置默认配置，如果没有在此处设置，当读不到值时，会使用零值
func (c *Config) initDefaultConfig() {
	// bitcoin node info
	viper.SetDefault("btc.rpcuser", "btc")
	viper.SetDefault("btc.rpcpassword", "blockchain")
	viper.SetDefault("btc.rpcport", 8332)
	viper.SetDefault("btc.rpcaddress", "39.108.13.219")

	// notify and http server
	viper.SetDefault("btc.listenaddress", ":8080")
	viper.SetDefault("btc.apiserveraddress", ":8888")
	viper.SetDefault("btc.notifyserveraddress", "https://push.eospark.com")

	// timer task
	// 0 0 每小时运行一次
	// 0 0 0 * 每天午夜运行一次
	viper.SetDefault("btc.repairblock", "0 0 * * * *")
	viper.SetDefault("btc.repairtransaction", "0 10 * * * *")
	viper.SetDefault("btc.repairunconfirmedtrx", "0 20 * * * *")
	viper.SetDefault("btc.repairall", "0 40 * * * *")

	// omni
	viper.SetDefault("btc.repairomniblock", "0 0 * * * *")
	viper.SetDefault("btc.repairomnitransaction", "0 10 * * * *")
	viper.SetDefault("btc.repairunconfirmedomnitrx", "0 20 * * * *")
	viper.SetDefault("btc.repairomniall", "0 50 * * * *")
	viper.SetDefault("btc.modifygorutinue", "0 0-59/10 * * * *")

	// goroutine
	viper.SetDefault("btc.trxgoroutinenum", 100)
	viper.SetDefault("btc.blockgoroutinenum", 20)
	viper.SetDefault("btc.maxgoroutine", 2000)

	// omni node info
	viper.SetDefault("omni.rpcuser", "omni")
	viper.SetDefault("omni.rpcpassword", "blockchain")
	viper.SetDefault("omni.rpcport", 8335)
	viper.SetDefault("omni.rpcaddress", "47.106.178.52")

	// database
	viper.SetDefault("db.address", "39.108.13.219:3306")
	viper.SetDefault("db.user", "btc")
	viper.SetDefault("db.password", "#Bitcoin_2019")
	viper.SetDefault("db.db_name", "btc_database")
	viper.SetDefault("db.max_open_conn", 2000)
	viper.SetDefault("db.max_idle_conn", 1000)
	viper.SetDefault("db.max_wait_timeout", 600)

	// redis
	viper.SetDefault("redis.address", "wallet-sz-inner.redis.rds.aliyuncs.com:6379")
	viper.SetDefault("redis.db_number", 0)

	// value
	viper.SetDefault("number.blocknum", 1000)
	viper.SetDefault("number.normal", 1.5)
	viper.SetDefault("number.priority", 2)
	viper.SetDefault("number.quick", 5)

}

//如果文件中配置被更改，此处读取会覆盖默认配置参数
func (c *Config) readFileConfig() {
	c.BtcOpt.RpcUser = viper.GetString("btc.rpcuser")
	c.BtcOpt.RpcPassword = viper.GetString("btc.rpcpassword")
	c.BtcOpt.RpcPort = viper.GetInt("btc.rpcport")
	c.BtcOpt.RpcAddress = viper.GetString("btc.rpcaddress")
	c.BtcOpt.ListenAddress = viper.GetString("btc.listenaddress")
	c.BtcOpt.ApiServerAddress = viper.GetString("btc.apiserveraddress")
	c.BtcOpt.NotifyServerAddress = viper.GetString("btc.notifyserveraddress")
	c.BtcOpt.RepairBlock = viper.GetString("btc.repairblock")
	c.BtcOpt.RepairTransaction = viper.GetString("btc.repairtransaction")
	c.BtcOpt.RepairUnconfirmedTrx = viper.GetString("btc.repairunconfirmedtrx")
	c.BtcOpt.RepairAll = viper.GetString("btc.repairall")
	c.BtcOpt.RepairOmniBlock = viper.GetString("btc.repairomniblock")
	c.BtcOpt.RepairOmniTransaction = viper.GetString("btc.repairomnitransaction")
	c.BtcOpt.RepairUnconfirmedOmniTrx = viper.GetString("btc.repairunconfirmedomnitrx")
	c.BtcOpt.RepairOmniAll = viper.GetString("btc.repairomniall")
	c.BtcOpt.ModifyGoroutine = viper.GetString("btc.modifygorutinue")
	c.BtcOpt.TrxGoroutineNum = viper.GetInt("btc.trxgoroutinenum")
	c.BtcOpt.BlockGoroutineNum = viper.GetInt("btc.blockgoroutinenum")
	c.BtcOpt.MaxGoroutine = viper.GetInt("btc.maxgoroutine")

	// omni
	c.OmniOpt.RpcUser = viper.GetString("omni.rpcuser")
	c.OmniOpt.RpcPassword = viper.GetString("omni.rpcpassword")
	c.OmniOpt.RpcPort = viper.GetInt("omni.rpcport")
	c.OmniOpt.RpcAddress = viper.GetString("omni.rpcaddress")

	// database
	c.DbOpt.Address = viper.GetString("db.address")
	c.DbOpt.User = viper.GetString("db.user")
	c.DbOpt.Password = viper.GetString("db.password")
	c.DbOpt.DbName = viper.GetString("db.db_name")
	c.DbOpt.MaxOpenConn = viper.GetInt("db.max_open_conn")
	c.DbOpt.MaxIdleConn = viper.GetInt("db.max_idle_conn")
	c.DbOpt.MaxWaitTimeout = viper.GetInt("db.max_wait_timeout")

	// redis
	c.RedisOpt.RedisAddress = viper.GetString("redis.address")
	c.RedisOpt.RedisDbNum = viper.GetInt("redis.db_number")

	// value
	c.Number.BlockNum = viper.GetInt("number.blocknum")
	c.Number.Normal = viper.GetFloat64("number.normal")
	c.Number.Priority = viper.GetFloat64("number.priority")
	c.Number.Quick = viper.GetFloat64("number.quick")
}
