package notify

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/database/tables"
	"github.com/BlockABC/wallet-btc-service/filestore"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
)

const (
	INTERVALLENGTH = 200
	NEARLENGTH     = 30
)

var BlockHeightBegin int32 = 0

func RepairAll() (errInfo error) {
	blockTimerBegin := BlockHeightBegin
	blockTimerEnd := blockTimerBegin + INTERVALLENGTH

	err, maxHeight := GetBlockHeight()
	if nil != err {
		log.Log.Error(err, " GetBlockHeight fail when repair all")
		errInfo = err
		return
	}

	if blockTimerEnd > maxHeight+1 {
		blockTimerEnd = maxHeight + 1
		if blockTimerEnd-blockTimerBegin < NEARLENGTH {
			blockTimerBegin = blockTimerEnd - NEARLENGTH
		}
	}

	// reset BlockHeightBegin
	defer func() {
		if nil == errInfo {
			BlockHeightBegin = blockTimerEnd
			filestore.RepairStoreInstance.SaveBlockBegin(BlockHeightBegin)
		}
	}()

	// repair lost block
	if err := repairLostBlock(blockTimerBegin, blockTimerEnd); nil != err {
		log.Log.Error(err, " repair lost block fail when repair all")
		errInfo = err
		return
	}

	// repair lost transaction
	if err := repairLostTransaction(blockTimerBegin, blockTimerEnd); nil != err {
		log.Log.Error(err, " repair lost transaction fail when repair all")
		errInfo = err
		return
	}

	// repair state and from
	if err := repairStateAndFrom(blockTimerBegin, blockTimerEnd); nil != err {
		log.Log.Error(err, " repair state and from  fail when repair all")
		errInfo = err
		return
	}

	// repair t_transaction_input_output_address_info
	if err := repairTransactionAddress(blockTimerBegin, blockTimerEnd); nil != err {
		log.Log.Error(err, " repair state and from  fail when repair all")
		errInfo = err
		return
	}

	return nil
}

func repairLostBlock(begin, end int32) error {
	blockHeights := []tables.TableBlockInfo{}
	if err := database.Db.Select("height").Where("height >= ? AND height < ?", begin, end).Order("height asc").Find(&blockHeights).Error; nil != err {
		log.Log.Error(err, fmt.Sprintf(" exec sql fail select height from t_block_info where height >= %d and height<%d ", begin, end))
		return err
	}

	for i := 0; i <= len(blockHeights); i++ {
		var allLostHight []int32

		if 0 == len(blockHeights) {
			for lostBegin := begin; lostBegin < end; lostBegin++ {
				allLostHight = append(allLostHight, lostBegin)
			}
		} else {
			if 0 == i {
				if blockHeights[0].Height != begin {
					for lostBegin := begin; lostBegin < blockHeights[0].Height; lostBegin++ {
						allLostHight = append(allLostHight, lostBegin)
					}
				}
			} else if len(blockHeights) == i {
				if blockHeights[i-1].Height != end-1 {
					for lostBegin := blockHeights[i-1].Height + 1; lostBegin < end; lostBegin++ {
						allLostHight = append(allLostHight, lostBegin)
					}
				}
			} else {
				if blockHeights[i].Height != blockHeights[i-1].Height+1 {
					for lostBegin := blockHeights[i-1].Height + 1; lostBegin < blockHeights[i].Height; lostBegin++ {
						allLostHight = append(allLostHight, lostBegin)
					}
				}
			}
		}

		if 0 == len(allLostHight) {
			continue
		}

		log.Log.Info("repairLostBlock one range lost block begin:", allLostHight[0], ", end:", allLostHight[len(allLostHight)-1])

		//get and store block
		wg := sync.WaitGroup{}
		trxCh := make(chan int, config.Cfg.BtcOpt.BlockGoroutineNum)
		for index, oneHeight := range allLostHight {
			if index < config.Cfg.BtcOpt.BlockGoroutineNum {
				wg.Add(1)
				go blockTask(index, oneHeight, trxCh, &wg)
			} else {
				<-trxCh
				wg.Add(1)
				go blockTask(index, oneHeight, trxCh, &wg)
			}
		}
		wg.Wait()

		log.Log.Info("repairLostBlock repair one range lost block success, begin:", allLostHight[0], ", end:", allLostHight[len(allLostHight)-1])
	}

	return nil
}

func blockTask(blockIndex int, blockHeight int32, ch chan int, group *sync.WaitGroup) {
	defer func() {
		ch <- blockIndex
		group.Done()

		// handle panic
		if err := recover(); err != nil {
			log.Log.Error(err, " blockTask panic when repair lost block, block height: ", blockHeight)
			FailBlockHeightHandle(blockHeight)
		}
	}()

	if err := GetBlockWithHeightAndStore(blockHeight); nil != err {
		log.Log.Error(err, " repair block fail, block height:", blockHeight)
		FailBlockHeightHandle(blockHeight)
	} else {
		log.Log.Info("repair block success, block height:", blockHeight)
	}
}

func GetBlockWithHeightAndStore(height int32) error {
	err, hash := GetBlockHashWithHeight(height)
	if nil != err {
		log.Log.Error(err, " GetBlockWithHeightAndStore get block hash with height fail, block height: ", height)
		return err
	}

	if err := GetBlockAndStoreNotUpdateStateAndFrom(hash); nil != err {
		log.Log.Error(err, " GetBlockWithHeightAndStore get block with hash and store fail, block hash: ", hash)
		return err
	}

	return nil
}

func GetBlockHashWithHeight(height int32) (errInfo error, blockHash string) {
	result, err := jsonrpc.Call(1, "getblockhash", []interface{}{height})
	if nil != err {
		log.Log.Error(err, " GetBlockHashWithHeight jsonrpc call getblockhash fail, block height: ", height)
		return err, ""
	}

	var hash string
	if err := json.Unmarshal(result, &hash); nil != err {
		log.Log.Error(err, " GetBlockHashWithHeight Unmarshal result to block hash fail")
		return err, ""
	}

	return nil, hash
}

func FailBlockHeightHandle(blockHeight int32) {
blockHash:
	err, hash := GetBlockHashWithHeight(blockHeight)
	if nil != err {
		goto blockHash
	}
	StoreFailBlockHash(hash)
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

func GetBlockDbMaxHeight() (errInfo error, blockheight int32) {
	type maxHeight struct {
		End int32
	}

	var result maxHeight
	selectSql := "select max(height) as end from t_block_info;"
	if err := database.Db.Raw(selectSql).Scan(&result).Error; nil != err {
		errInfo = err
		return
	}

	return nil, result.End
}

// lost transaction
type TableTempTransactionCount struct {
	Blockhash string `json:"blockhash"         	gorm:"column:blockhash;type:char(64);unique"` //当前交易所在区块的哈希
	Ntx       int64  `json:"ntx"         		gorm:"column:ntx"`                                 //交易个数
}

func (t *TableTempTransactionCount) TableName() string {
	return "t_temp_transaction_count"
}

type TableTempLostTransactionHash struct {
	Txid        string `json:"txid"     		gorm:"column:txid;type:char(64);unique"`
	Blockheight int32  `json:"blockheight"           gorm:"column:blockheight"`
}

func (t *TableTempLostTransactionHash) TableName() string {
	return "t_temp_lost_transaction_hash"
}

var TempTransactionCountMutex sync.Mutex

func repairLostTransaction(begin, end int32) error {
	//lock temp table
	TempTransactionCountMutex.Lock()
	defer TempTransactionCountMutex.Unlock()

	// create temp table
	if database.Db.HasTable(&TableTempTransactionCount{}) {
		if err := database.Db.Exec("DROP TABLE t_temp_transaction_count;").Error; err != nil {
			log.Log.Error(err, " drop table t_temp_transaction_count fail")
			return err
		}
	}

	if err := database.Db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(&TableTempTransactionCount{}).Error; err != nil {
		log.Log.Error(err, " create table t_temp_transaction_count fail")
		return err
	}
	// drop table
	defer database.Db.Exec("DROP TABLE t_temp_transaction_count;")

	// insert transaction count data
	blockHashSelectSql := fmt.Sprintf("select hash from t_block_info where height >= %d AND height < %d", begin, end)
	insertSql := fmt.Sprintf(`insert into t_temp_transaction_count(blockhash, ntx) select blockhash, count(*) from t_transaction_info where blockhash in (%s) group by blockhash;`, blockHashSelectSql)
	if err := database.Db.Exec(insertSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", insertSql)
		return err
	}

	insertSql = fmt.Sprintf("insert into t_temp_transaction_count(blockhash, ntx) select hash, 0 from t_block_info where hash in (%s) and hash not in (select blockhash from t_temp_transaction_count);", blockHashSelectSql)
	if err := database.Db.Exec(insertSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", insertSql)
		return err
	}

	// find lost transaction block
	deleteBlockSql := "delete t1 from t_temp_transaction_count t1, t_block_info t2 where t1.blockhash = t2.hash and t1.ntx = t2.ntx;"
	if err := database.Db.Exec(deleteBlockSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", deleteBlockSql)
		return err
	}

	// create temp table
	if database.Db.HasTable(&TableTempLostTransactionHash{}) {
		if err := database.Db.Exec("DROP TABLE t_temp_lost_transaction_hash;").Error; err != nil {
			log.Log.Error(err, " drop table t_temp_lost_transaction_hash fail")
			return err
		}
	}

	if err := database.Db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(&TableTempLostTransactionHash{}).Error; err != nil {
		log.Log.Error(err, " create table t_temp_lost_transaction_hash fail")
		return err
	}
	// drop table
	defer database.Db.Exec("DROP TABLE t_temp_lost_transaction_hash;")

	// get lost transaction block
	var realBlockHash []TableTempTransactionCount
	if err := database.Db.Find(&realBlockHash).Error; nil != err {
		log.Log.Error(err, " select * from t_temp_transaction_count fail")
		return err
	}

	if 0 == len(realBlockHash) {
		return nil
	}
	log.Log.Info("repairLostTransaction block height begin:", begin, ", block height end:", end, ", block hash and block transaction number:", realBlockHash)

	// insert transaction hash
	for _, oneBlock := range realBlockHash {
		result, err := jsonrpc.Call(1, "getblock", []interface{}{oneBlock.Blockhash, 2})
		if nil != err {
			log.Log.Error(err, " repairLostTransaction jsonrpc call getblock fail, hash: ", oneBlock.Blockhash)
			return err
		}
		newBlock := Block{}
		if err := json.Unmarshal(result, &newBlock); nil != err {
			log.Log.Error(err, " repairLostTransaction Unmarshal result to block struct fail")
			return err
		}

		// insert into t_temp_lost_transaction_hash
		page := len(newBlock.Tx) / database.MAX_WITH_INSERT
		for i := 0; i < page; i++ {
			begin := i * database.MAX_WITH_INSERT
			end := (i + 1) * database.MAX_WITH_INSERT
			pageHash := newBlock.Tx[begin:end]
			insertSql := "insert into t_temp_lost_transaction_hash(txid, blockheight) values "
			for index, oneTrx := range pageHash {
				if 0 == index {
					insertSql += fmt.Sprintf(`('%s', %d)`, oneTrx.Txid, newBlock.Height)
				} else {
					insertSql += fmt.Sprintf(`,('%s', %d)`, oneTrx.Txid, newBlock.Height)
				}
			}
			insertSql += ";"
			if err := database.Db.Exec(insertSql).Error; nil != err {
				log.Log.Error(err, " exec sql fail: ", insertSql)
				return err
			}
		}

		if 0 != len(newBlock.Tx)%database.MAX_WITH_INSERT {
			begin := page * database.MAX_WITH_INSERT
			pageHash := newBlock.Tx[begin:]
			insertSql := "insert into t_temp_lost_transaction_hash(txid, blockheight) values "
			for index, oneTrx := range pageHash {
				if 0 == index {
					insertSql += fmt.Sprintf(`('%s', %d)`, oneTrx.Txid, newBlock.Height)
				} else {
					insertSql += fmt.Sprintf(`,('%s', %d)`, oneTrx.Txid, newBlock.Height)
				}
			}
			insertSql += ";"
			if err := database.Db.Exec(insertSql).Error; nil != err {
				log.Log.Error(err, " exec sql fail: ", insertSql)
				return err
			}
		}
	}

	// find lost transaction
	deleteTrxSql := "delete t1 from t_temp_lost_transaction_hash t1, t_transaction_info t2 where t1.txid = t2.txid;"
	if err := database.Db.Exec(deleteTrxSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", deleteTrxSql)
		return err
	}

	// get lost transaction
	var realTrxHash []TableTempLostTransactionHash
	if err := database.Db.Find(&realTrxHash).Error; nil != err {
		log.Log.Error(err, " select * from t_temp_lost_transaction_hash fail")
		return err
	}

	if 0 == len(realTrxHash) {
		return nil
	}
	log.Log.Info("repairLostTransaction block height begin:", begin, ", block height end:", end, ", real lost transaction hash and belong to block heght:", realTrxHash)

	// repair lost transaction
	wg := sync.WaitGroup{}
	trxCh := make(chan int, config.TrxGoroutineRuntime)
	taskFunc := func(trxIndex int, height int32, trx string, group *sync.WaitGroup) {
		defer func() {
			trxCh <- trxIndex
			group.Done()

			// handle panic
			if err := recover(); err != nil {
				log.Log.Error(err, " panic occur when repair lost transaction, block height:", height, ", transaction hash:", trx)
				StoreFailTransactionHash(height, trx)
			}
		}()

		if err := GetTransactionAndStore(height, trx); nil != err {
			StoreFailTransactionHash(height, trx)
			log.Log.Error("repair lost transaction fail, block height:", height, ", transaction hash:", trx)
		} else {
			log.Log.Info("repair lost transaction success, block height:", height, ", transaction hash:", trx)
		}
	}
	for index, oneTrx := range realTrxHash {
		if index >= config.TrxGoroutineRuntime {
			<-trxCh
		}
		wg.Add(1)
		go taskFunc(index, oneTrx.Blockheight, oneTrx.Txid, &wg)
	}
	wg.Wait()

	return nil
}

func repairStateAndFrom(begin, end int32) error {
	subquerySql := fmt.Sprintf("select hash from t_block_info where height >= %d AND height < %d", begin, end)

	// update t_output_info state
	updateOutputStateSql := fmt.Sprintf("update t_output_info t1, t_input_info t2 set t1.state=1 where t1.hash=t2.txid and t1.n=t2.vout and t1.blockhash in(%s);", subquerySql)
	if err := database.Db.Exec(updateOutputStateSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", updateOutputStateSql)
		return err
	}
	log.Log.Info("repair state update t_output_info state with t_input_info success, block height begin:", begin, ", block height end:", end)

	// update t_input_info from and value with t_output_info
	updateInputFromSql := fmt.Sprintf("update t_input_info t1, t_output_info t2 set t1.from=t2.to, t1.value=t2.value where t1.txid=t2.hash and t1.vout=t2.n and t1.blockhash in(%s);", subquerySql)
	if err := database.Db.Exec(updateInputFromSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", updateInputFromSql)
		return err
	}
	log.Log.Info("repair state update t_input_info from and value with t_output_info success, block height begin:", begin, ", block height end:", end)

	return nil
}

func repairTransactionAddress(begin, end int32) error {
	insertInputSql := fmt.Sprintf("replace into t_transaction_input_output_address_info (blockhash, `time`, txid, `address`, isfrom) select distinct blockhash, `time`, `hash`, `from`, 1 from t_input_info where `from` != '' and blockhash in (select hash from t_block_info where height >= %d and height < %d);",
		begin, end)
	if err := database.Db.Exec(insertInputSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", insertInputSql)
		return err
	}
	log.Log.Info("repairTransactionAddress insert into t_transaction_input_output_address_info input address success, block height begin:", begin, ", end:", end)
	insertOutputSql := fmt.Sprintf("replace into t_transaction_input_output_address_info (blockhash, `time`, txid, `address`, isfrom) select distinct blockhash, `time`, `hash`, `to`, 0 from t_output_info where `to` != '' and blockhash in (select hash from t_block_info where height >= %d and height < %d);",
		begin, end)
	if err := database.Db.Exec(insertOutputSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", insertOutputSql)
		return err
	}
	log.Log.Info("repairTransactionAddress insert into t_transaction_input_output_address_info output address success, block height begin:", begin, ", end:", end)
	return nil
}
