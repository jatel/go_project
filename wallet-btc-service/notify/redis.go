package notify

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
)

const (
	REDISUNFMDTRXKEY = "UnconfirmedTransaction"
)

type RedisInput struct {
	Txid    string `json:"txid"`
	Vout    int64  `json:"vout"`
	Address string `json:"address"`
	Value   int64  `json:"value"`
}

type RedisOutput struct {
	N         int64    `json:"n"`
	Value     int64    `json:"value"`
	Asm       string   `json:"asm"`
	Addresses []string `json:"addresses"`
	Type      string   `json:"type"`
	IsSpent   bool     `json:"isspent"`
}

type RedisTransaction struct {
	Txid        string        `json:"txid"`
	ReceiveTime int64         `json:"receivetime"`
	Vin         []RedisInput  `json:"vin"`
	Vout        []RedisOutput `json:"vout"`
}

func convertUnconfirmedTransactionToRedisTransaction(oneTransaction *Transaction) (error, *RedisTransaction) {
	if isCoinbase(oneTransaction) {
		return errors.New("transaction is coinbase"), nil
	}

	vin := make([]RedisInput, 0)
	// get output from redis
	cacheInfo, err := database.RedisDb.HGetAll(REDISUNFMDTRXKEY).Result()
	if nil != err {
		log.Log.Error(err, " redis get UnconfirmedTransaction fail")
		return err, nil
	}

	cacheTransaction := []RedisTransaction{}
	for k, v := range cacheInfo {
		one := RedisTransaction{}
		if err := json.Unmarshal([]byte(v), &one); nil != err {
			log.Log.Error(err, " Unmarshal to RedisTransaction fail, redis map UnconfirmedTransaction key:", k)
			continue
		}
		cacheTransaction = append(cacheTransaction, one)
	}

	// set address and value
	notFindInput := make([]Input, 0)
	for _, oneInput := range oneTransaction.Vin {
		for cacheIndex, oneCache := range cacheTransaction {
			for index, oneCacheOutput := range oneCache.Vout {
				if oneCacheOutput.N == oneInput.Vout && oneCache.Txid == oneInput.Txid {
					if nil != oneCacheOutput.Addresses && 0 != len(oneCacheOutput.Addresses) {
						vin = append(vin, RedisInput{oneInput.Txid, oneInput.Vout, oneCacheOutput.Addresses[0], oneCacheOutput.Value})
						cacheTransaction[cacheIndex].Vout[index].IsSpent = true
						SaveOneRedisTransaction(&cacheTransaction[cacheIndex])
						goto nextOne
					}
				}
			}
		}
		notFindInput = append(notFindInput, oneInput)
	nextOne:
	}

	// get output from database
	realNotFindInput := make([]Input, 0)
	if 0 != len(notFindInput) {
		strCondition := ""
		for index, oneInput := range notFindInput {
			if 0 == index {
				strCondition += fmt.Sprintf("'%s'", oneInput.Txid)
			} else {
				strCondition += fmt.Sprintf(",'%s'", oneInput.Txid)
			}
		}

		type info struct {
			Hash  string `json:"hash"`
			N     int64  `json:"n"`
			To    string `json:"to"`
			Value int64  `json:"value"`
		}

		inputInfo := []info{}
		findInputInfo := make([]info, 0)
		selectSql := fmt.Sprintf("select `hash`, `n`, `to`, `value` from t_output_info where (`hash`) in (%s);", strCondition)
		if err := database.Db.Raw(selectSql).Scan(&inputInfo).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", selectSql)
			return err, nil
		}

		for _, oneInput := range notFindInput {
			for _, oneTable := range inputInfo {
				if oneTable.Hash == oneInput.Txid && oneTable.N == oneInput.Vout {
					vin = append(vin, RedisInput{oneInput.Txid, oneInput.Vout, oneTable.To, oneTable.Value})
					findInputInfo = append(findInputInfo, oneTable)
					goto nextNotFind
				}
			}
			realNotFindInput = append(realNotFindInput, oneInput)
		nextNotFind:
		}

		if 0 != len(findInputInfo) {
			for _, oneInfo := range findInputInfo {
				updateSql := fmt.Sprintf("update t_output_info set state=2 where `hash`='%s' and `n`=%d;", oneInfo.Hash, oneInfo.N)
				if err := database.Db.Exec(updateSql).Error; nil != err {
					log.Log.Error(err, " exec sql fail: ", updateSql)
					continue
				}
			}
		}
	}

	// can not find output
	if 0 != len(realNotFindInput) {
		log.Log.Error("can not find unspent output inputs:", realNotFindInput)
		for _, oneInput := range realNotFindInput {
			vin = append(vin, RedisInput{oneInput.Txid, oneInput.Vout, "", 0})
		}
	}

	// set vout
	vout := make([]RedisOutput, 0)
	for _, oneOutput := range oneTransaction.Vout {
		bUsed := false
		for _, oneCache := range cacheTransaction {
			for index, oneCacheInput := range oneCache.Vin {
				if oneCacheInput.Txid == oneTransaction.Txid && oneCacheInput.Vout == oneOutput.N {
					if nil != oneOutput.ScriptPubKey.Addresses && 0 != len(oneOutput.ScriptPubKey.Addresses) {
						oneCache.Vin[index].Address = oneOutput.ScriptPubKey.Addresses[0]
						oneCache.Vin[index].Value = int64(oneOutput.Value * 100000000)
						SaveOneRedisTransaction(&oneCache)
					}
					bUsed = true
					goto find
				}
			}
		}
	find:
		vout = append(vout, RedisOutput{oneOutput.N, int64(oneOutput.Value * 100000000), oneOutput.ScriptPubKey.Asm, oneOutput.ScriptPubKey.Addresses, oneOutput.ScriptPubKey.Type, bUsed})
	}

	result := RedisTransaction{oneTransaction.Txid, oneTransaction.ReceiveTime, vin, vout}
	return nil, &result
}

func SaveUnconfirmedTransactionToRedis(oneTransaction *Transaction) error {
	err, oneRedisTransaction := convertUnconfirmedTransactionToRedisTransaction(oneTransaction)
	if nil != err {
		return err
	}

	saveErr := SaveOneRedisTransaction(oneRedisTransaction)

	if nil == saveErr {
		oneTransactionNotify(oneTransaction)
	}

	return saveErr
}

func SaveOneRedisTransaction(oneRedisTransaction *RedisTransaction) error {
	info, err := json.Marshal(*oneRedisTransaction)
	if nil != err {
		log.Log.Error(err, " marshal oneRedisTransaction fail")
		return err
	}

	if _, err := database.RedisDb.HSet(REDISUNFMDTRXKEY, oneRedisTransaction.Txid, info).Result(); nil != err {
		log.Log.Error(err, " save oneRedisTransaction to redis fail, transaction hash:", oneRedisTransaction.Txid)
		return err
	}

	return nil
}

func DeleteRedisBlockTransaction(newBlock *Block) error {
	hashs := make([]string, 0)
	for _, oneTransaction := range newBlock.Tx {
		hashs = append(hashs, oneTransaction.Txid)
	}

	return DeleteRedisTransactionByHashs(hashs)
}

func DeleteRedisTransactionByHashs(hashs []string) error {
	for _, oneHash := range hashs {
		bExist, err := database.RedisDb.HExists(REDISUNFMDTRXKEY, oneHash).Result()
		if nil != err {
			log.Log.Error(err, " DeleteRedisTransactionByHashs if transaction exist fail, transaction hash:", oneHash)
			continue
		}
		if bExist {
			_, err := database.RedisDb.HDel(REDISUNFMDTRXKEY, oneHash).Result()
			if nil != err {
				log.Log.Error(err, " DeleteRedisTransactionByHashs delete redis transaction fail, transaction hash:", oneHash)
			}
		}
	}

	return nil
}

func GetRedisUsedInfoByAddress(address []string) (error, map[string]bool) {
	allInfo, err := database.RedisDb.HGetAll(REDISUNFMDTRXKEY).Result()
	if nil != err {
		log.Log.Error(err, " get transaction from redis by address fail, address:", address)
		return err, nil
	}

	result := make(map[string]bool)
	for key, value := range allInfo {
		oneRedisTransaction := RedisTransaction{}
		if err := json.Unmarshal([]byte(value), &oneRedisTransaction); nil != err {
			log.Log.Error(err, " Unmarshal to redis transaction fail, key:", key)
			continue
		}

		// input
		for _, oneInput := range oneRedisTransaction.Vin {
			for _, oneAddress := range address {
				if oneAddress == oneInput.Address {
					result[oneAddress] = true
					break
				}
			}
		}

		// output
		for _, oneOutput := range oneRedisTransaction.Vout {
			for _, oneAddress := range address {
				if nil != oneOutput.Addresses && 0 != len(oneOutput.Addresses) && oneAddress == oneOutput.Addresses[0] {
					result[oneAddress] = true
					break
				}
			}
		}
	}

	return nil, result
}

func GetRedisUnconfirmedTransactionByAddress(address []string) (error, []RedisTransaction) {
	allInfo, err := database.RedisDb.HGetAll(REDISUNFMDTRXKEY).Result()
	if nil != err {
		log.Log.Error(err, " get transaction from redis by address fail, address:", address)
		return err, nil
	}

	temp := make([]RedisTransaction, 0)
	for key, value := range allInfo {
		oneRedisTransaction := RedisTransaction{}
		if err := json.Unmarshal([]byte(value), &oneRedisTransaction); nil != err {
			log.Log.Error(err, " Unmarshal to redis transaction fail, key:", key)
			continue
		}

		// input
		for _, oneInput := range oneRedisTransaction.Vin {
			for _, oneAddress := range address {
				if oneAddress == oneInput.Address {
					temp = append(temp, oneRedisTransaction)
					goto nextOne
				}
			}
		}

		// output
		for _, oneOutput := range oneRedisTransaction.Vout {
			for _, oneAddress := range address {
				if nil != oneOutput.Addresses && 0 != len(oneOutput.Addresses) && oneAddress == oneOutput.Addresses[0] {
					temp = append(temp, oneRedisTransaction)
					goto nextOne
				}
			}
		}
	nextOne:
	}

	result := make([]RedisTransaction, 0)
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

func GetRedisUnconfirmedTransactionByTxid(txid string) (error, bool, *RedisTransaction) {
	bExist, err := database.RedisDb.HExists(REDISUNFMDTRXKEY, txid).Result()
	if nil != err {
		log.Log.Error(err, " GetRedisUnconfirmedTransactionByTxid if transaction exist fail, transaction hash:", txid)
		return err, false, nil
	}
	if !bExist {
		return nil, false, nil
	}

	info, err := database.RedisDb.HGet(REDISUNFMDTRXKEY, txid).Result()
	if nil != err {
		log.Log.Error(err, " GetRedisUnconfirmedTransactionByTxid get transaction fail, transaction hash:", txid)
		return err, false, nil
	}

	oneRedisTransaction := RedisTransaction{}
	if err := json.Unmarshal([]byte(info), &oneRedisTransaction); nil != err {
		log.Log.Error(err, " Unmarshal to redis transaction fail, transaction hash:", txid)
		return err, false, nil
	}

	return nil, true, &oneRedisTransaction
}

func GetRedisUnconfirmedTransactionBalanceByAddress(address []string) (error, map[string]int64) {
	allInfo, err := database.RedisDb.HGetAll(REDISUNFMDTRXKEY).Result()
	if nil != err {
		log.Log.Error(err, " get transaction from redis by address fail, address:", address)
		return err, nil
	}

	result := make(map[string]int64)
	for key, value := range allInfo {
		oneRedisTransaction := RedisTransaction{}
		if err := json.Unmarshal([]byte(value), &oneRedisTransaction); nil != err {
			log.Log.Error(err, " Unmarshal to redis transaction fail, key:", key)
			continue
		}

		// output
		for _, oneOutput := range oneRedisTransaction.Vout {
			for _, oneAddress := range address {
				if nil != oneOutput.Addresses && 0 != len(oneOutput.Addresses) && oneAddress == oneOutput.Addresses[0] && !oneOutput.IsSpent {
					balance, ok := result[oneAddress]
					if ok {
						result[oneAddress] = balance + oneOutput.Value
					} else {
						result[oneAddress] = oneOutput.Value
					}
				}
			}
		}

	}
	return nil, result
}

type UnspentOutput struct {
	Txid    string
	Address string
	N       int64
	Value   int64
}

func GetRedisUnconfirmedTransactionUnspentByAddress(address []string) (error, []UnspentOutput) {
	allInfo, err := database.RedisDb.HGetAll(REDISUNFMDTRXKEY).Result()
	if nil != err {
		log.Log.Error(err, " get transaction from redis by address fail, address:", address)
		return err, nil
	}

	result := make([]UnspentOutput, 0)
	for key, value := range allInfo {
		oneRedisTransaction := RedisTransaction{}
		if err := json.Unmarshal([]byte(value), &oneRedisTransaction); nil != err {
			log.Log.Error(err, " Unmarshal to redis transaction fail, key:", key)
			continue
		}

		// output
		for _, oneOutput := range oneRedisTransaction.Vout {
			for _, oneAddress := range address {
				if nil != oneOutput.Addresses && 0 != len(oneOutput.Addresses) && oneAddress == oneOutput.Addresses[0] && !oneOutput.IsSpent {
					result = append(result, UnspentOutput{oneRedisTransaction.Txid, oneAddress, oneOutput.N, oneOutput.Value})
				}
			}
		}
	}

	return nil, result
}

func GetAllRedisUnconfirmedTransaction() (error, []RedisTransaction) {
	allInfo, err := database.RedisDb.HGetAll(REDISUNFMDTRXKEY).Result()
	if nil != err {
		log.Log.Error(err, " get transaction from redis  fail")
		return err, nil
	}

	result := make([]RedisTransaction, 0)
	for key, value := range allInfo {
		oneRedisTransaction := RedisTransaction{}
		if err := json.Unmarshal([]byte(value), &oneRedisTransaction); nil != err {
			log.Log.Error(err, " Unmarshal to redis transaction fail, key:", key)
			continue
		}
		result = append(result, oneRedisTransaction)
	}

	return nil, result
}
