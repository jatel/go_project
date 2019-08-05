package omni

import (
	"encoding/json"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
)

func HandleOmniBlock(oneHeight int32) error {
	// get block omni transaction list
	result, err := jsonrpc.OmniCall(1, "omni_listblocktransactions", []interface{}{oneHeight})
	if nil != err {
		log.Log.Error(err, " HandleOmniBlock jsonrpc OmniCall omni_listblocktransactions fail, height: ", oneHeight)
		StoreFailOmniBlockHeight(oneHeight)
		return err
	}
	trxHashs := make([]string, 0)
	if err := json.Unmarshal(result, &trxHashs); nil != err {
		log.Log.Error(err, " HandleOmniBlock Unmarshal result to trxHashs struct fail")
		StoreFailOmniBlockHeight(oneHeight)
		return err
	}

	if 0 == len(trxHashs) {
		return nil
	}

	for _, oneHash := range trxHashs {
		if err := HandleOmniTransaction(oneHash, false); nil != err {
			StoreFailOmniTransactionHah(oneHash)
		}
	}

	// delete unconfirmed omni transaction
	DeleteRedisUnfmdOmniTransaction(trxHashs)

	return nil
}
