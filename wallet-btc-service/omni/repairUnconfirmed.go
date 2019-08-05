package omni

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
)

func RepairUnconfirmedOmniTransaction() error {
	// get all unconfirmed omni transaction
	allUnconfirmedTransaction := make([]OmniTransaction, 0)
	result, err := jsonrpc.OmniCall(1, "omni_listpendingtransactions", []interface{}{})
	if nil != err {
		log.Log.Error(err, " RepairUnconfirmedOmniTransaction jsonrpc OmniCall omni_listpendingtransactions fail")
		return err
	}

	if err := json.Unmarshal(result, &allUnconfirmedTransaction); nil != err {
		log.Log.Error(err, " RepairUnconfirmedOmniTransaction Unmarshal result to allUnconfirmedHash struct fail")
		return err
	}

	// get real lost unconfirmed omni transaction hash
	err, allRedis := GetAllRedisUnconfirmedOmniTransaction()
	if nil != err {
		log.Log.Error(err, " RepairUnconfirmedTransaction Get All Redis omni Unconfirmed Transaction fail")
		return err
	}

	// delete block omni transaction
	var strRedisHash string
	for index, oneRedis := range allRedis {
		if 0 == index {
			strRedisHash += "'" + oneRedis.Txid + "'"
		} else {
			strRedisHash += ",'" + oneRedis.Txid + "'"
		}
	}

	if 0 != len(strRedisHash) {
		type blockTxid struct {
			Txid string
		}

		existTransactionHash := make([]blockTxid, 0)
		txSelectSql := fmt.Sprintf("select txid from t_omni_transaction_info where txid in (%s);", strRedisHash)
		if err := database.Db.Raw(txSelectSql).Scan(&existTransactionHash).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", txSelectSql)
			return err
		}

		deleteRedisHash := make([]string, 0)
		for _, oneBlockTxid := range existTransactionHash {
			deleteRedisHash = append(deleteRedisHash, oneBlockTxid.Txid)
		}

		log.Log.Info("repair unconfirmed omni transaction get real need delete unconfirmed transaction success, real need delete unconfirmed omni transaction hashs:", deleteRedisHash)

		if err := DeleteRedisUnfmdOmniTransaction(deleteRedisHash); nil != err {
			log.Log.Error(err, " repair unconfirmed omni transaction delete block transaction fail")
			return err
		}
	}

	// get real lost omni unconfirmed transaction
	realLostOmniTransaction := make([]OmniTransaction, 0)
	for _, oneUnconfirmedTransaction := range allUnconfirmedTransaction {
		bFind := false
		for _, oneRedis := range allRedis {
			if oneRedis.Txid == oneUnconfirmedTransaction.Txid {
				bFind = true
				break
			}
		}

		if !bFind {
			realLostOmniTransaction = append(realLostOmniTransaction, oneUnconfirmedTransaction)
		}
	}

	for _, oneLost := range realLostOmniTransaction {
		oneLost.ReceiveTime = time.Now().Unix()
		if err := SaveUnconfirmedOmniTransactionToRedis(&oneLost); nil != err {
			log.Log.Error(err, " RepairUnconfirmedOmniTransaction save omni transaction fail, omni transaction hash:", oneLost.Txid)
			continue
		}

		log.Log.Info("repair lost unconfirmed omni transaction success, unconfirmed omni transaction hashs:", oneLost.Txid)
	}

	return nil
}
