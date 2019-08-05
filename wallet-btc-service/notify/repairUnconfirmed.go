package notify

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
)

func RepairUnconfirmedTransaction() error {
	// get mempool transaction
	mempool, err := jsonrpc.Call(1, "getrawmempool", nil)
	if nil != err {
		log.Log.Error(err, " RepairUnconfirmedTransaction jsonrpc call getrawmempool fail")
		return err
	}
	allTrxHash := []string{}
	if err := json.Unmarshal(mempool, &allTrxHash); nil != err {
		log.Log.Error(err, " RepairUnconfirmedTransaction Unmarshal result to allTrxHash struct fail")
		return err
	}
	log.Log.Info("repair unconfirmed transaction get memory transaction success, unconfirmed transaction hashs:", allTrxHash)
	if 0 == len(allTrxHash) {
		return nil
	}

	// get all redis transaction
	err, allRedis := GetAllRedisUnconfirmedTransaction()
	if nil != err {
		log.Log.Error(err, " RepairUnconfirmedTransaction Get All Redis Unconfirmed Transaction fail")
		return err
	}

	// delete blcok transaction
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
		txSelectSql := fmt.Sprintf("select txid from t_transaction_info where txid in (%s);", strRedisHash)
		if err := database.Db.Raw(txSelectSql).Scan(&existTransactionHash).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", txSelectSql)
			return err
		}

		deleteRedisHash := make([]string, 0)
		for _, oneBlockTxid := range existTransactionHash {
			deleteRedisHash = append(deleteRedisHash, oneBlockTxid.Txid)
		}

		log.Log.Info("repair unconfirmed transaction get real need delete unconfirmed transaction success, real need delete unconfirmed transaction hashs:", deleteRedisHash)

		if err := DeleteRedisTransactionByHashs(deleteRedisHash); nil != err {
			log.Log.Error(err, " repair unconfirmed transaction delete block transaction fail")
			return err
		}
	}

	// get real lost unconfirmed transaction
	realHash := make([]string, 0)
	for _, oneHash := range allTrxHash {
		bFind := false
		for _, oneRedis := range allRedis {
			if oneRedis.Txid == oneHash {
				bFind = true
				break
			}
		}

		if !bFind {
			realHash = append(realHash, oneHash)
		}
	}

	if 0 != len(realHash) {
		log.Log.Info("repair unconfirmed transaction get real lost unconfirmed transaction success, real lost unconfirmed transaction hashs:", realHash)

		// insert transaction
		for _, oneHash := range realHash {
			result, err := jsonrpc.Call(1, "getrawtransaction", []interface{}{oneHash, true})
			if nil != err {
				log.Log.Error(err, " RepairUnconfirmedTransaction jsonrpc call getrawtransaction fail, hash: ", oneHash)
				continue
			}
			newTransaction := Transaction{}
			if err := json.Unmarshal(result, &newTransaction); nil != err {
				log.Log.Error(err, " RepairUnconfirmedTransaction Unmarshal result to transaction struct fail")
				continue
			}

			// handle transaction
			newTransaction.ReceiveTime = time.Now().Unix()
			if err := SaveUnconfirmedTransactionToRedis(&newTransaction); nil != err {
				log.Log.Error(err, " RepairUnconfirmedTransaction save transaction fail, transaction hash:", newTransaction.Txid)
				continue
			}
			log.Log.Info("repair lost unconfirmed transaction success, transaction hashs:", oneHash)
		}
	}

	return nil
}
