package omni

import (
	"encoding/json"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
)

func GetOmniBlockHeight() (errInfo error, blockheight int32) {
	result, err := jsonrpc.OmniCall(1, "getblockcount", []interface{}{})
	if nil != err {
		log.Log.Error(err, " GetOmniBlockHeight jsonrpc call getblockcount fail")
		errInfo = err
		return
	}

	var height int32
	if err := json.Unmarshal(result, &height); nil != err {
		log.Log.Error(err, " GetOmniBlockHeight Unmarshal result to block height fail")
		errInfo = err
		return
	}

	return nil, height
}
