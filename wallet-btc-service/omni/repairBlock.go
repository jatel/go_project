package omni

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/common/utility"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/database/tables"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
)

var blockRepair *OmniBlockRepair = NewOmniBlockRepair()

type OmniBlockRepair struct {
	Filename string
	Rw       sync.RWMutex
}

func NewOmniBlockRepair() *OmniBlockRepair {
	baseDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	baseDir = strings.Replace(baseDir, "\\", "/", -1)
	return &OmniBlockRepair{Filename: baseDir + "/repairOmniBlock"}
}

func StoreFailOmniBlockHeight(height int32) error {
	blockRepair.Rw.Lock()
	defer blockRepair.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(blockRepair.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if nil != err {
		log.Log.Error(err, " store fail omni block height open file fail, block height:", height)
		return err
	}
	defer fs.Close()

	// write block height
	buf := bufio.NewWriter(fs)
	if _, err := fmt.Fprintln(buf, height); nil != err {
		log.Log.Error(err, " store fail omni block height write file fail, block height:", height)
		return err
	}
	result := buf.Flush()
	if nil == result {
		log.Log.Info("store fail omni block height success, block height:", height)
	} else {
		log.Log.Error("store fail omni block height fail, block height:", height)
	}
	return result
}

func BatchStoreFailOmniHeight(allHeight []int32) error {
	blockRepair.Rw.Lock()
	defer blockRepair.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(blockRepair.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if nil != err {
		log.Log.Error(err, " batch store fail omni block height open file fail, block height:", allHeight)
		return err
	}
	defer fs.Close()

	// write block height
	buf := bufio.NewWriter(fs)
	for _, oneHeight := range allHeight {
		if _, err := fmt.Fprintln(buf, oneHeight); nil != err {
			log.Log.Error(err, " batch store fail omni block height write one omni block height to file fail, one block height:", oneHeight)
		}
	}

	result := buf.Flush()
	if nil == result {
		log.Log.Info("batch store fail omni block height success, omni block heights:", allHeight)
	} else {
		log.Log.Error("batch store fail omni block height fail, omni block heights:", allHeight)
	}
	return result
}

func RepairFailOmniBlockHeight() error {
	if !utility.IsFileExist(blockRepair.Filename) {
		return nil
	}

	blockRepair.Rw.Lock()
	defer blockRepair.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(blockRepair.Filename, os.O_RDWR, 0666)
	if nil != err {
		log.Log.Error(err, " repair fail omni block height open file fail")
		return err
	}
	defer fs.Close()

	// read file
	allFileHeight := []int32{}
	br := bufio.NewReader(fs)
	for {
		strHeight, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if nil != err {
			log.Log.Error(err, " read repair omni block file fail")
			return err
		}

		height, err := strconv.ParseInt(string(strHeight), 10, 32)
		if nil != err {
			log.Log.Error(err, " convert one height fail, height:", strHeight)
			continue
		}

		bExist := false
		for _, v := range allFileHeight {
			if v == int32(height) {
				bExist = true
				break
			}
		}
		if !bExist {
			allFileHeight = append(allFileHeight, int32(height))
		}
	}
	fs.Truncate(0)

	if 0 == len(allFileHeight) {
		return nil
	}

	// get block transaction
	allFailBlockHeight := []int32{}
	allTransactionHash := []string{}
	for _, oneHeight := range allFileHeight {
		result, err := jsonrpc.OmniCall(1, "omni_listblocktransactions", []interface{}{oneHeight})
		if nil != err {
			log.Log.Error(err, " repairFailOmniBlockHeight jsonrpc OmniCall omni_listblocktransactions fail, height: ", oneHeight)
			allFailBlockHeight = append(allFailBlockHeight, oneHeight)
			continue
		}

		trxHashs := make([]string, 0)
		if err := json.Unmarshal(result, &trxHashs); nil != err {
			log.Log.Error(err, " repairFailOmniBlockHeight Unmarshal result to trxHashs struct fail")
			allFailBlockHeight = append(allFailBlockHeight, oneHeight)
			continue
		}
		allTransactionHash = append(allTransactionHash, trxHashs...)
	}

	// store fail block height
	fs.Seek(0, 0)
	buf := bufio.NewWriter(fs)
	for _, oneFailHeight := range allFailBlockHeight {
		if _, err := fmt.Fprintln(buf, oneFailHeight); nil != err {
			log.Log.Error(err, " repairFailOmniBlockHeight store one fail omni block height write to file fail, one fail omni block height:", oneFailHeight)
		}
	}
	buf.Flush()

	// get real block transaction hash
	err, realHash := getOmniBlockTransactionHashs(allTransactionHash)
	if nil != err {
		log.Log.Error(err, " repairFailOmniBlockHeight get real fail block transaction hash fail")
		return err
	}

	return repairFailOmniBlockTransaction(realHash)
}

type TableTempOmniBlockTrxHash struct {
	Txid string `json:"txid"     		gorm:"column:hash;type:char(64);unique"`
}

func (t *TableTempOmniBlockTrxHash) TableName() string {
	return "t_temp_omni_block_trx_hash"
}

var TableTempOmniBlockTrxHashMutex sync.Mutex

func getOmniBlockTransactionHashs(allTransactionHash []string) (error, []string) {
	if nil == allTransactionHash || 0 == len(allTransactionHash) {
		return nil, nil
	}

	// ues in when a few hash
	if len(allTransactionHash) <= database.MAX_WITh_IN {
		return simpleGetOmniTransactionHashs(allTransactionHash)
	}

	//lock temp table
	TableTempOmniBlockTrxHashMutex.Lock()
	defer TableTempOmniBlockTrxHashMutex.Unlock()

	// create temp table
	if database.Db.HasTable(&TableTempOmniBlockTrxHash{}) {
		if err := database.Db.Exec("DROP TABLE t_temp_omni_block_trx_hash;").Error; err != nil {
			log.Log.Error(err, " drop table t_temp_omni_block_trx_hash fail")
			return err, nil
		}
	}

	if err := database.Db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(&TableTempOmniBlockTrxHash{}).Error; err != nil {
		log.Log.Error(err, " create table t_temp_omni_block_trx_hash fail")
		return err, nil
	}
	// drop table
	defer database.Db.Exec("DROP TABLE t_temp_omni_block_trx_hash;")

	// insert data
	page := len(allTransactionHash) / database.MAX_WITH_INSERT
	for i := 0; i < page; i++ {
		begin := i * database.MAX_WITH_INSERT
		end := (i + 1) * database.MAX_WITH_INSERT
		pageHash := allTransactionHash[begin:end]
		insertSql := "insert into t_temp_omni_block_trx_hash(txid) values "
		for index, oneTransactionHash := range pageHash {
			if 0 == index {
				insertSql += fmt.Sprintf(`('%s')`, oneTransactionHash)
			} else {
				insertSql += fmt.Sprintf(`,('%s')`, oneTransactionHash)
			}
		}
		insertSql += ";"
		if err := database.Db.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			return err, nil
		}
	}

	if 0 != len(allTransactionHash)%database.MAX_WITH_INSERT {
		begin := page * database.MAX_WITH_INSERT
		pageHash := allTransactionHash[begin:]
		insertSql := "insert into t_temp_omni_block_trx_hash(txid) values "
		for index, oneTransactionHash := range pageHash {
			if 0 == index {
				insertSql += fmt.Sprintf(`('%s')`, oneTransactionHash)
			} else {
				insertSql += fmt.Sprintf(`,('%s')`, oneTransactionHash)
			}
		}
		insertSql += ";"
		if err := database.Db.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			return err, nil
		}
	}

	// delete repetition data
	deleteSql := "delete t1 from t_temp_omni_block_trx_hash t1, t_omni_transaction_info t2 where t1.hash=t2.txid;"
	if err := database.Db.Exec(deleteSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", deleteSql)
		return err, nil
	}

	// get real block hash
	var realTableHash []TableTempOmniBlockTrxHash
	if err := database.Db.Find(&realTableHash).Error; nil != err {
		log.Log.Error(err, " select * from t_temp_omni_block_trx_hash fail")
		return err, nil
	}

	resultHashs := []string{}
	for _, oneTableHash := range realTableHash {
		resultHashs = append(resultHashs, oneTableHash.Txid)
	}

	return nil, resultHashs
}

func repairFailOmniBlockTransaction(realHash []string) error {
	if nil == realHash || 0 == len(realHash) {
		return nil
	}

	// store transaction
	wg := sync.WaitGroup{}
	trxCh := make(chan int, config.TrxGoroutineRuntime)
	taskFunc := func(oneIndex int, trx string, group *sync.WaitGroup) {
		defer func() {
			trxCh <- oneIndex
			group.Done()

			// handle panic
			if err := recover(); err != nil {
				log.Log.Error(err, " panic occur when repair fail omni block, omni transaction hash:", trx)
				StoreFailOmniTransactionHah(trx)
			}
		}()

		if err := HandleOmniTransaction(trx, false); nil != err {
			log.Log.Error("repair fail omni block one omni transaction fail, omni transaction hash:", trx)
			StoreFailOmniTransactionHah(trx)
		} else {
			log.Log.Info("repair fail omni block one omni transaction success, omni transaction hash:", trx)
		}
	}
	for index, oneTransactionHash := range realHash {
		if index >= config.TrxGoroutineRuntime {
			<-trxCh
		}
		wg.Add(1)
		go taskFunc(index, oneTransactionHash, &wg)
	}
	wg.Wait()

	return nil
}

func simpleGetOmniTransactionHashs(allTransactionHash []string) (error, []string) {
	var result []tables.TableOmniTransactionInfo
	if err := database.Db.Model(&tables.TableOmniTransactionInfo{}).Where("`txid` IN (?)", allTransactionHash).Find(&result).Error; nil != err {
		log.Log.Error(err, " select * from t_omni_transaction_info fail")
		return err, nil
	}

	realHash := []string{}
	for _, oneHash := range allTransactionHash {
		bFind := false
		for _, oneTableHash := range result {
			if oneHash == oneTableHash.Txid {
				bFind = true
				break
			}
		}

		if !bFind {
			realHash = append(realHash, oneHash)
		}
	}

	return nil, realHash
}
