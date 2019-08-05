package omni

import (
	"encoding/json"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
)

const (
	REDISUNFMDOMNITRXKEY = "UnconfirmedOmniTransaction"
)

func SaveUnconfirmedOmniTransactionToRedis(oneOmniTransaction *OmniTransaction) error {
	info, err := json.Marshal(*oneOmniTransaction)
	if nil != err {
		log.Log.Error(err, " marshal one omni transaction fail")
		return err
	}

	if _, err := database.RedisDb.HSet(REDISUNFMDOMNITRXKEY, oneOmniTransaction.Txid, info).Result(); nil != err {
		log.Log.Error(err, " save one omni transaction to redis fail, transaction hash:", oneOmniTransaction.Txid)
		return err
	}

	oneOmniTransactionNotify(oneOmniTransaction)

	return nil
}

func DeleteRedisUnfmdOmniTransaction(trxHashs []string) error {
	for _, oneHash := range trxHashs {
		bExist, err := database.RedisDb.HExists(REDISUNFMDOMNITRXKEY, oneHash).Result()
		if nil != err {
			log.Log.Error(err, " DeleteRedisUnfmdOmniTransaction if omni transaction exist fail, omni transaction hash:", oneHash)
			continue
		}
		if bExist {
			_, err := database.RedisDb.HDel(REDISUNFMDOMNITRXKEY, oneHash).Result()
			if nil != err {
				log.Log.Error(err, " DeleteRedisUnfmdOmniTransaction delete omni transaction fail, omni transaction hash:", oneHash)
			}
		}
	}
	return nil
}

func GetAllRedisUnconfirmedOmniTransaction() (error, []OmniTransaction) {
	allInfo, err := database.RedisDb.HGetAll(REDISUNFMDOMNITRXKEY).Result()
	if nil != err {
		log.Log.Error(err, " get omni transaction from redis  fail")
		return err, nil
	}

	result := make([]OmniTransaction, 0)
	for key, value := range allInfo {
		oneRedisOmniTransaction := OmniTransaction{}
		if err := json.Unmarshal([]byte(value), &oneRedisOmniTransaction); nil != err {
			log.Log.Error(err, " Unmarshal to omni transaction fail, key:", key)
			continue
		}
		result = append(result, oneRedisOmniTransaction)
	}

	return nil, result
}

func GetRedisUnconfirmedOmniTransactionByTxid(txid string) (error, bool, *OmniTransaction) {
	bExist, err := database.RedisDb.HExists(REDISUNFMDOMNITRXKEY, txid).Result()
	if nil != err {
		log.Log.Error(err, " GetRedisUnconfirmedOmniTransactionByTxid if omni transaction exist fail, omni transaction hash:", txid)
		return err, false, nil
	}
	if !bExist {
		return nil, false, nil
	}

	info, err := database.RedisDb.HGet(REDISUNFMDOMNITRXKEY, txid).Result()
	if nil != err {
		log.Log.Error(err, " GetRedisUnconfirmedOmniTransactionByTxid get omni transaction fail, omni transaction hash:", txid)
		return err, false, nil
	}

	oneRedisOmniTransaction := OmniTransaction{}
	if err := json.Unmarshal([]byte(info), &oneRedisOmniTransaction); nil != err {
		log.Log.Error(err, " Unmarshal to omni transaction fail, omni transaction hash:", txid)
		return err, false, nil
	}

	return nil, true, &oneRedisOmniTransaction
}

func GetRedisUnconfirmedOmniTransactionByAddressAndType(address []string, transactionType []string) (error, []OmniTransaction) {
	allInfo, err := database.RedisDb.HGetAll(REDISUNFMDOMNITRXKEY).Result()
	if nil != err {
		log.Log.Error(err, " get transaction from redis by address fail, address:", address)
		return err, nil
	}

	temp := make([]OmniTransaction, 0)
	for key, value := range allInfo {
		oneRedisOmniTransaction := OmniTransaction{}
		if err := json.Unmarshal([]byte(value), &oneRedisOmniTransaction); nil != err {
			log.Log.Error(err, " Unmarshal to redis omni transaction fail, key:", key)
			continue
		}

		// type judgment
		bType := false
		for _, oneType := range transactionType {
			if oneRedisOmniTransaction.Type == oneType {
				bType = true
				break
			}
		}
		if !bType {
			continue
		}

		// address judgment
		bFind := false
		for _, oneAddress := range address {
			if oneRedisOmniTransaction.Referenceaddress == oneAddress || oneRedisOmniTransaction.Sendingaddress == oneAddress {
				bFind = true
				break
			}
		}

		if !bFind {
			continue
		}

		// one real transaction
		temp = append(temp, oneRedisOmniTransaction)
	}

	result := make([]OmniTransaction, 0)
	if 0 != len(temp) {
		for len(temp) > 0 {
			value := temp[0].ReceiveTime
			index := 0
			for i := 0; i < len(temp); i++ {
				if temp[i].ReceiveTime > value {
					value = temp[i].ReceiveTime
					index = i
				}
			}
			result = append(result, temp[index])
			if 1 == len(temp) {
				break
			}

			if 0 == index {
				temp = temp[index+1:]
			} else if index == len(temp)-1 {
				temp = temp[:index]
			} else {
				temp = append(temp[:index], temp[index+1:]...)
			}
		}
	}

	return nil, result
}
