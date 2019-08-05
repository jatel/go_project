package httpserver

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database/tables"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
	"github.com/BlockABC/wallet-btc-service/notify"
)

// transaction input
type input struct {
	From_address string `json:"from_address"`
	From_txid    string `json:"from_txid"`
	Vin_index    int64  `json:"vin_index"`
	Value        string `json:"value"`
}

// transaction output
type output struct {
	To_address string `json:"to_address"`
	Vout_index int64  `json:"vout_index"`
	Value      string `json:"value"`
	Type       string `json:"type"`
	Asm        string `json:"asm"`
}

// transaction info
type transactionInfo struct {
	Txid          string   `json:"txid"`
	Block         int32    `json:"block"`
	Blockhash     string   `json:"blockhash"`
	Iscoinbase    bool     `json:"iscoinbase"`
	Fee           string   `json:"fee"`
	Inputs        []input  `json:"inputs"`
	Outputs       []output `json:"outputs"`
	Confirmations int32    `json:"confirmations"`
	Receivetime   string   `json:"receivetime"`
	Blocktime     string   `json:"blocktime"`
}

// convert redis transaction to transaction info
func convertRedisTransactionToTransactionInfo(oneRedisTransaction *notify.RedisTransaction) transactionInfo {
	inputs := []input{}
	outputs := []output{}
	var inputValue int64 = 0
	var outputValue int64 = 0

	for _, oneInput := range oneRedisTransaction.Vin {
		inputs = append(inputs, input{oneInput.Address, oneInput.Txid, oneInput.Vout, fmt.Sprintf("%d", oneInput.Value)})
		inputValue += oneInput.Value
	}

	for _, oneOutput := range oneRedisTransaction.Vout {
		if nil != oneOutput.Addresses && 0 != len(oneOutput.Addresses) {
			outputs = append(outputs, output{oneOutput.Addresses[0], oneOutput.N, fmt.Sprintf("%d", oneOutput.Value), oneOutput.Type, oneOutput.Asm})
			outputValue += oneOutput.Value
		}
	}

	return transactionInfo{
		Txid:          oneRedisTransaction.Txid,
		Fee:           fmt.Sprintf("%d", inputValue-outputValue),
		Inputs:        inputs,
		Outputs:       outputs,
		Confirmations: -1,
		Receivetime:   time.Unix(oneRedisTransaction.ReceiveTime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700"),
	}
}

// convert view transaction to transaction info
func convertViewTransactionInfo(oneTransaction []tables.ViewTransactionInfo) transactionInfo {
	inputs := []input{}
	outputs := []output{}
	var inputValue int64 = 0
	var outputValue int64 = 0
	for _, oneRecord := range oneTransaction {
		oneInput := input{
			From_address: oneRecord.Fromaddress,
			From_txid:    oneRecord.Fromhash,
			Vin_index:    oneRecord.Fromindex,
			Value:        fmt.Sprintf("%d", oneRecord.Fromvalue),
		}

		// if exist not append
		bFindInput := false
		for _, oneExistInput := range inputs {
			if oneExistInput.From_txid == oneInput.From_txid && oneExistInput.Vin_index == oneInput.Vin_index {
				bFindInput = true
				break
			}
		}

		if !bFindInput {
			inputs = append(inputs, oneInput)
			inputValue += oneRecord.Fromvalue
		}

		oneOutput := output{
			To_address: oneRecord.Toaddress,
			Vout_index: oneRecord.Toindex,
			Value:      fmt.Sprintf("%d", oneRecord.Tovalue),
			Type:       oneRecord.Totype,
			Asm:        oneRecord.Toasm,
		}

		// if exist not append
		bFindOutput := false
		for _, oneExistOutput := range outputs {
			if oneExistOutput.Vout_index == oneOutput.Vout_index {
				bFindOutput = true
				break
			}
		}

		if !bFindOutput {
			outputs = append(outputs, oneOutput)
			outputValue += oneRecord.Tovalue
		}
	}

	bIsCoinbase := false
	if 1 == oneTransaction[0].Iscoinbase {
		bIsCoinbase = true
	}

	_, height := GetBlockHeight()

	return transactionInfo{
		Txid:          oneTransaction[0].Transactionhash,
		Block:         oneTransaction[0].Blockheight,
		Blockhash:     oneTransaction[0].Blockhash,
		Iscoinbase:    bIsCoinbase,
		Fee:           fmt.Sprintf("%d", inputValue-outputValue),
		Inputs:        inputs,
		Outputs:       outputs,
		Confirmations: height - oneTransaction[0].Blockheight,
		Blocktime:     time.Unix(oneTransaction[0].Time, 0).UTC().Format("2006-01-02T15:04:05.999999-0700"),
	}
}

func convertToTransactionInfo(pageTransaction []tables.TableTransactionInfo, pageInput []tables.TableInputInfo, pageOutput []tables.TableOutputInfo) []transactionInfo {
	result := make([]transactionInfo, 0)
	_, height := GetBlockHeight()
	for _, oneTrx := range pageTransaction {
		inputs := []input{}
		outputs := []output{}
		var inputValue int64 = 0
		var outputValue int64 = 0
		for _, oneInputInfo := range pageInput {
			if oneInputInfo.Hash == oneTrx.Txid {
				oneInput := input{
					From_address: oneInputInfo.From,
					From_txid:    oneInputInfo.Txid,
					Vin_index:    oneInputInfo.Vout,
					Value:        fmt.Sprintf("%d", oneInputInfo.Value),
				}
				inputs = append(inputs, oneInput)
				inputValue += oneInputInfo.Value
			}
		}
		for _, oneOutputInfo := range pageOutput {
			if oneOutputInfo.Hash == oneTrx.Txid {
				oneOutput := output{
					To_address: oneOutputInfo.To,
					Vout_index: oneOutputInfo.N,
					Value:      fmt.Sprintf("%d", oneOutputInfo.Value),
					Type:       oneOutputInfo.Type,
					Asm:        oneOutputInfo.Asm,
				}
				outputs = append(outputs, oneOutput)
				outputValue += oneOutputInfo.Value
			}
		}

		bIsCoinbase := false
		if 1 == oneTrx.Iscoinbase {
			bIsCoinbase = true
		}
		oneTxInfo := transactionInfo{
			Txid:          oneTrx.Txid,
			Block:         oneTrx.Blockheight,
			Blockhash:     oneTrx.Blockhash,
			Iscoinbase:    bIsCoinbase,
			Fee:           fmt.Sprintf("%d", inputValue-outputValue),
			Inputs:        inputs,
			Outputs:       outputs,
			Confirmations: height - oneTrx.Blockheight,
			Blocktime:     time.Unix(oneTrx.Time, 0).UTC().Format("2006-01-02T15:04:05.999999-0700"),
		}
		result = append(result, oneTxInfo)
	}

	return result
}

func GetBlockHeight() (errInfo error, blockheight int32) {
	result, err := jsonrpc.Call(1, "getblockcount", []interface{}{})
	if nil != err {
		log.Log.Error(err, " GetBlockHeight jsonrpc call getblockcount fail")
		errInfo = err
		return
	}

	var height int32
	if err := json.Unmarshal(result, &height); nil != err {
		log.Log.Error(err, " GetBlockHeight Unmarshal result to block height fail")
		errInfo = err
		return
	}

	return nil, height
}
