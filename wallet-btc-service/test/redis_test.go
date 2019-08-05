package test 
import (
	"fmt"
	"testing"
	"encoding/json"
	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/notify"
)

const (
	REDISUNFMDTRXKEY = "UnconfirmedTransaction"
)

func TestRedis(t *testing.T) {
	if err := database.InitRedis(config.Cfg.RedisOpt.RedisAddress, config.Cfg.RedisOpt.RedisDbNum); nil != err {
		panic(err)
	}

	one := notify.RedisTransaction{"f5e26c8b82401c585235c572ba8265f16f7d9304ed8e31c198eab571754f5331", 1561447903,
		[]notify.RedisInput{{"02460245a84266a112c4209507f29e5483e0a70f411c769d89aa912015d0e7bd", 0, "15eFPqBDrLJp53nPzj8LdHKBGaxqHQ9ZWN", 24000000}},
		[]notify.RedisOutput{{0, 23000000, "OP_DUP OP_HASH160 abb5bb81ddbb3a35865354186384095976836959 OP_EQUALVERIFY OP_CHECKSIG", []string{"1GevFo7SaLKc5GL2dntFerTW2dj5nZUG7d"}, "pubkey", false}}}

	info, err := json.Marshal(one)
	if nil != err {
		fmt.Println(err)
	}
	user := make(map[string]interface{})
	user[one.Txid] = info
	ret, err := database.RedisDb.HMSet(REDISUNFMDTRXKEY, user).Result()
	if nil != err {
		fmt.Println(err)
		return
	}

	fmt.Println(ret)

	hash, err := database.RedisDb.HGetAll(REDISUNFMDTRXKEY).Result()
	if nil != err {
		fmt.Println(err)
		return
	}
	for k, v := range hash {
		fmt.Printf("key: %v, value: %v ", k, v)
		my := notify.RedisTransaction{}
		json.Unmarshal([]byte(v), &my)
	}
}