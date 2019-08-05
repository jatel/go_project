package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/BlockABC/wallet-btc-service/jsonrpc"
	"github.com/BlockABC/wallet-btc-service/notify"
)

func TestGetBlock(t *testing.T) {
	var hash = "00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048"
	result, err := jsonrpc.Call(1, "getblock", []interface{}{hash, 2})
	if nil != err {
		fmt.Println(err)
	}
	fmt.Println(string(result))
	newBlock := notify.Block{}
	if err := json.Unmarshal(result, &newBlock); nil != err {
		fmt.Println(err)
	}
	fmt.Println(newBlock)
}

func TestGetRawMempool(t *testing.T) {
	mempool, err := jsonrpc.Call(1, "getrawmempool", nil)
	if nil != err {
		fmt.Println(err)
	}
	var allTrxHash []string
	fmt.Println(string(mempool))
	if err := json.Unmarshal(mempool, &allTrxHash); nil != err {
		fmt.Println(err)
	}
	fmt.Println(allTrxHash)
}

func TestGetRawTransaction(t *testing.T) {
	//oneTrx := "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098"
	//oneTrx := "d299f0c965de2a117aa157ed6b135924bada0cf0cb0dbc1ab0e5b74190244af3"
	oneTrx := "9d8e5b85fb21ccae567a07c81445399ded250dcc4428f52a46cf5ce10ab327c8"
	result, err := jsonrpc.Call(1, "getrawtransaction", []interface{}{oneTrx, true})
	if nil != err {
		fmt.Println(err)
	}
	fmt.Println(string(result))
	newTransaction := notify.Transaction{}
	if err := json.Unmarshal(result, &newTransaction); nil != err {
		fmt.Println(err)
	}
	fmt.Println(newTransaction)
}
