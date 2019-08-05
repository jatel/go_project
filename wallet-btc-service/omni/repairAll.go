package omni

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/filestore"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
)

var OmniHeightBegin int32 = 0

const TIMERRANGE = 200
const NEARNUMBER = 30

func RepairAll() (errInfo error) {
	// begin and end
	timerBeginHeight := OmniHeightBegin
	timerEndHeight := timerBeginHeight + TIMERRANGE

	// get max block height
	err, maxBlockHeight := GetOmniBlockHeight()
	if nil != err {
		log.Log.Error(err, " RepairAll GetOmniBlockHeight fail")
		errInfo = err
		return
	}

	// reset begin and end
	if timerEndHeight > maxBlockHeight+1 {
		timerEndHeight = maxBlockHeight + 1
		if timerEndHeight-timerBeginHeight < NEARNUMBER {
			timerBeginHeight = timerEndHeight - NEARNUMBER
		}
	}

	// reset OmniHeightBegin
	defer func() {
		if nil == errInfo {
			OmniHeightBegin = timerEndHeight
			filestore.RepairStoreInstance.SaveOmniBegin(OmniHeightBegin)
		}
	}()

	// get lost omni transaction
	err, trxHashs := GetLostOmniTransaction(timerBeginHeight, timerEndHeight)
	if nil != err {
		log.Log.Error(err, " RepairAll get lost omni transaction fail")
		errInfo = err
		return
	}

	if nil == trxHashs || 0 == len(trxHashs) {
		return nil
	}
	log.Log.Info("repair lost omni transaction begin height:", timerBeginHeight, ", end height:", timerEndHeight, ", real lost omni transaction hashs:", trxHashs)

	// save lost omni transaction
	if err := SaveLostOmniTransaction(trxHashs); nil != err {
		log.Log.Error(err, " RepairAll save lost omni transaction fail")
		errInfo = err
		return
	}

	return nil
}

type TableTempOmniTransactionHash struct {
	Txid string `json:"txid"     		gorm:"column:txid;type:char(64);unique"`
}

func (t *TableTempOmniTransactionHash) TableName() string {
	return "t_temp_omni_transaction_hash"
}

var TempOmniTransactionHashMutex sync.Mutex

func GetLostOmniTransaction(begin, end int32) (error, []string) {
	//lock temp table
	TempOmniTransactionHashMutex.Lock()
	defer TempOmniTransactionHashMutex.Unlock()

	// create temp table
	if database.Db.HasTable(&TableTempOmniTransactionHash{}) {
		if err := database.Db.Exec("DROP TABLE t_temp_omni_transaction_hash;").Error; err != nil {
			log.Log.Error(err, " drop table t_temp_omni_transaction_hash fail")
			return err, nil
		}
	}

	if err := database.Db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(&TableTempOmniTransactionHash{}).Error; err != nil {
		log.Log.Error(err, " create table t_temp_omni_transaction_hash fail")
		return err, nil
	}
	// drop table
	defer database.Db.Exec("DROP TABLE t_temp_omni_transaction_hash;")

	// insert into temp table
	bHaveLost := false
	for ; begin < end; begin++ {
		result, err := jsonrpc.OmniCall(1, "omni_listblocktransactions", []interface{}{begin})
		if nil != err {
			log.Log.Error(err, " GetLostOmniTransaction jsonrpc OmniCall omni_listblocktransactions fail, height: ", begin)
			StoreFailOmniBlockHeight(begin)
			continue
		}
		trxHashs := make([]string, 0)
		if err := json.Unmarshal(result, &trxHashs); nil != err {
			log.Log.Error(err, " GetLostOmniTransaction Unmarshal result to trxHashs struct fail")
			StoreFailOmniBlockHeight(begin)
			continue
		}

		if 0 != len(trxHashs) {
			bHaveLost = true
			insertSql := generateLostOmniTrxSql(trxHashs)
			if err := database.Db.Exec(insertSql).Error; nil != err {
				log.Log.Error(err, " exec sql fail: ", insertSql)
				StoreFailOmniBlockHeight(begin)
				continue
			}
		}
	}

	if !bHaveLost {
		return nil, []string{}
	}

	// delete repetition data
	deleteSql := "delete t1 from t_temp_omni_transaction_hash t1, t_omni_transaction_info t2 where t1.txid=t2.txid;"
	if err := database.Db.Exec(deleteSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", deleteSql)
		return err, nil
	}

	// get real omni transaction hash
	var realHash []TableTempOmniTransactionHash
	if err := database.Db.Find(&realHash).Error; nil != err {
		log.Log.Error(err, " select * from t_temp_omni_transaction_hash fail")
		return err, nil
	}

	// result
	result := []string{}
	for _, oneTableHash := range realHash {
		result = append(result, oneTableHash.Txid)
	}
	return nil, result
}

func generateLostOmniTrxSql(trxHashs []string) string {
	insertSql := "insert into t_temp_omni_transaction_hash(txid) values "
	for index, oneHash := range trxHashs {
		if 0 == index {
			insertSql += fmt.Sprintf(`('%s')`, oneHash)
		} else {
			insertSql += fmt.Sprintf(`,('%s')`, oneHash)
		}
	}
	insertSql += ";"
	return insertSql
}

func SaveLostOmniTransaction(trxHashs []string) error {
	// save one transaction
	wg := sync.WaitGroup{}
	trxCh := make(chan int, config.TrxGoroutineRuntime)
	taskFunc := func(index int, trx string, group *sync.WaitGroup) {
		defer func() {
			trxCh <- index
			group.Done()

			// handle panic
			if err := recover(); err != nil {
				log.Log.Error(err, " panic occur when repair lost omni transactions, omni transaction hash: ", trx)
			}
		}()

		HandleOmniTransaction(trx, false)
		log.Log.Info("repair lost omni transaction success, omni transaction hash:", trx)
	}

	for index, oneTrx := range trxHashs {
		if index >= config.TrxGoroutineRuntime {
			<-trxCh
		}

		wg.Add(1)
		go taskFunc(index, oneTrx, &wg)
	}
	wg.Wait()

	return nil
}
