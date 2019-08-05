package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
	"github.com/BlockABC/wallet-btc-service/omni"
)

func TestOmniTransaction(t *testing.T) {
	// database initialize
	dbOpt := config.Cfg.DbOpt
	err := database.Initialize(dbOpt.Address, dbOpt.User, dbOpt.Password, dbOpt.DbName, dbOpt.MaxOpenConn, dbOpt.MaxIdleConn, dbOpt.MaxWaitTimeout)
	if err != nil {
		panic(err)
	}

	//oneHeight := 324140
	oneHeight := 443058
	result, err := jsonrpc.OmniCall(1, "omni_listblocktransactions", []interface{}{oneHeight})
	if nil != err {
		fmt.Println(err, " OmniTask jsonrpc OmniCall omni_listblocktransactions fail, height: ", oneHeight)

	}
	trxHashs := make([]string, 0)
	if err := json.Unmarshal(result, &trxHashs); nil != err {
		fmt.Println(err, " OmniTask Unmarshal result to trxHashs struct fail")

	}

	if 0 != len(trxHashs) {
		for _, oneHash := range trxHashs {
			result, err := jsonrpc.OmniCall(1, "omni_gettransaction", []interface{}{oneHash})
			if nil != err {
				fmt.Println(err, " OmniTask jsonrpc OmniCall omni_gettransaction fail, transaction hash: ", oneHash)

			}
			fmt.Println(string(result))
			newTransaction := omni.OmniTransaction{}
			if err := json.Unmarshal(result, &newTransaction); nil != err {
				fmt.Println(err, " OmniTask Unmarshal result to OmniTransaction struct fail")
			}
			if err := omni.SaveOmniTransaction(&newTransaction); nil != err {
				fmt.Println(err, " OmniTask save omni transaction fail, transaction hash:", oneHash)
			}
		}
	}
}
