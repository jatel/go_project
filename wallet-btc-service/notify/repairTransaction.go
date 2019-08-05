package notify

import (
	"bufio"
	"encoding/json"
	"errors"
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

type TransactionRepair struct {
	Filename string
	Rw       sync.RWMutex
}

func NewTransactionRepair() *TransactionRepair {
	baseDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	baseDir = strings.Replace(baseDir, "\\", "/", -1)
	return &TransactionRepair{Filename: baseDir + "/repairTransaction"}
}

func GenerateRepairTrxInfo(blockheight int32, transactionHash string) string {
	return fmt.Sprintf("%d_%s", blockheight, transactionHash)
}

func GetRepairTrxInfo(info string) (errInfo error, blockheight int32, transactionHash string) {
	result := strings.Split(info, "_")
	if 2 != len(result) {
		errInfo = errors.New("GetRepairTrxInfo fail invalid parameter")
		return
	}

	height, err := strconv.ParseInt(result[0], 10, 32)
	if nil != err {
		errInfo = err
		return
	}

	blockheight = int32(height)
	transactionHash = result[1]
	return
}

func (repair *TransactionRepair) StoreFailHash(blockheight int32, transactionHash string) error {
	repair.Rw.Lock()
	defer repair.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(repair.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if nil != err {
		log.Log.Error(err, " store fail transaction hash open file fail, block height:", blockheight, ", transaction hash:", transactionHash)
		return err
	}
	defer fs.Close()

	// write Transaction hash
	buf := bufio.NewWriter(fs)
	if _, err := fmt.Fprintln(buf, GenerateRepairTrxInfo(blockheight, transactionHash)); nil != err {
		log.Log.Error(err, " store fail transaction hash write file fail, block height:", blockheight, ", transaction hash:", transactionHash)
		return err
	}
	result := buf.Flush()
	if nil == result {
		log.Log.Info("store fail transaction hash success, block height:", blockheight, ", transaction hash:", transactionHash)
	} else {
		log.Log.Error("store fail transaction hash fail, block height:", blockheight, ", transaction hash:", transactionHash)
	}
	return result
}

func (repair *TransactionRepair) BatchStoreFailHash(allHeight []int32, allHash []string) error {
	if len(allHeight) != len(allHash) {
		return errors.New("allHeight allHash and don't have the same length")
	}

	repair.Rw.Lock()
	defer repair.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(repair.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if nil != err {
		log.Log.Error(err, "batch store transaction fail hash open file fail")
		return err
	}
	defer fs.Close()

	// write Transaction hash
	buf := bufio.NewWriter(fs)
	for i := 0; i < len(allHash); i++ {
		if _, err := fmt.Fprintln(buf, GenerateRepairTrxInfo(allHeight[i], allHash[i])); nil != err {
			log.Log.Error(err, "batch store transaction hash one transaction hash write file fail, one block height:", allHeight[i], ", one transaction hash:", allHash[i])
		}
	}
	result := buf.Flush()
	if nil == result {
		log.Log.Info("repair batch store transaction hash success, transaction hashs:", allHash)
	} else {
		log.Log.Error("repair batch store transaction hash fail, transaction hashs:", allHash)
	}
	return result
}

func (repair *TransactionRepair) RepairAllItems() error {
	if !utility.IsFileExist(repair.Filename) {
		return nil
	}

	repair.Rw.Lock()

	// open file
	fs, err := os.Open(repair.Filename)
	if nil != err {
		log.Log.Error(err, " Repair open transaction file fail")
		repair.Rw.Unlock()
		return err
	}

	// read file
	allHeight := []int32{}
	allHash := []string{}
	br := bufio.NewReader(fs)
	for {
		oneLine, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if nil != err {
			log.Log.Error(err, " read repair transaction file fail")
			fs.Close()
			repair.Rw.Unlock()
			return err
		}

		err, height, hash := GetRepairTrxInfo(string(oneLine))
		if nil != err {
			continue
		}

		bExist := false
		for _, v := range allHash {
			if v == hash {
				bExist = true
				break
			}
		}

		if !bExist {
			allHash = append(allHash, hash)
			allHeight = append(allHeight, height)
		}
	}
	fs.Close()

	// clean file
	if newFs, err := os.OpenFile(repair.Filename, os.O_WRONLY|os.O_TRUNC, 0666); nil != err {
		newFs.Close()
	}
	repair.Rw.Unlock()

	if 0 == len(allHash) {
		return nil
	}

	mapHashHeight := make(map[string]int32, 0)
	for i := 0; i < len(allHash); i++ {
		mapHashHeight[allHash[i]] = allHeight[i]
	}

	// get real fail block hash
	err, realHashs := getRealTrxHashs(allHash)
	if nil != err {
		repair.BatchStoreFailHash(allHeight, allHash)
		return err
	}

	if nil == realHashs || len(realHashs) == 0 {
		return nil
	}
	log.Log.Info("repair fail transaction get real transaction hash success, real fail transaction hashs:", realHashs)

	// delete unconfirmed transaction
	DeleteRedisTransactionByHashs(allHash)

	realHashHeight := make(map[string]int32, 0)
	for i := 0; i < len(realHashs); i++ {
		realHashHeight[realHashs[i]] = mapHashHeight[realHashs[i]]
	}
	return repair.repairTransactions(realHashHeight)
}

type TableTempTransactionHash struct {
	Txid string `json:"txid"     		gorm:"column:txid;type:char(64);unique"`
}

func (t *TableTempTransactionHash) TableName() string {
	return "t_temp_transaction_hash"
}

var TempTransactionHashMutex sync.Mutex

func getRealTrxHashs(allHash []string) (error, []string) {
	// ues in when a few hash
	if len(allHash) <= database.MAX_WITh_IN {
		return simpleGetRealTrxHashs(allHash)
	}

	//lock temp table
	TempTransactionHashMutex.Lock()
	defer TempTransactionHashMutex.Unlock()

	// create temp table
	if database.Db.HasTable(&TableTempTransactionHash{}) {
		if err := database.Db.Exec("DROP TABLE t_temp_transaction_hash;").Error; err != nil {
			log.Log.Error(err, " drop table t_temp_transaction_hash fail")
			return err, nil
		}
	}

	if err := database.Db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(&TableTempTransactionHash{}).Error; err != nil {
		log.Log.Error(err, " create table t_temp_transaction_hash fail")
		return err, nil
	}
	// drop table
	defer database.Db.Exec("DROP TABLE t_temp_transaction_hash;")

	// insert data
	page := len(allHash) / database.MAX_WITH_INSERT
	for i := 0; i < page; i++ {
		begin := i * database.MAX_WITH_INSERT
		end := (i + 1) * database.MAX_WITH_INSERT
		pageHash := allHash[begin:end]
		insertSql := "insert into t_temp_transaction_hash(txid) values "
		for index, transactionHash := range pageHash {
			if 0 == index {
				insertSql += fmt.Sprintf(`('%s')`, transactionHash)
			} else {
				insertSql += fmt.Sprintf(`,('%s')`, transactionHash)
			}
		}
		insertSql += ";"
		if err := database.Db.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			return err, nil
		}
	}

	if 0 != len(allHash)%database.MAX_WITH_INSERT {
		begin := page * database.MAX_WITH_INSERT
		pageHash := allHash[begin:]
		insertSql := "insert into t_temp_transaction_hash(txid) values "
		for index, transactionHash := range pageHash {
			if 0 == index {
				insertSql += fmt.Sprintf(`('%s')`, transactionHash)
			} else {
				insertSql += fmt.Sprintf(`,('%s')`, transactionHash)
			}
		}
		insertSql += ";"
		if err := database.Db.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			return err, nil
		}
	}

	// delete repetition data
	deleteSql := "delete t1 from t_temp_transaction_hash t1, t_transaction_info t2 where t1.txid=t2.txid;"
	if err := database.Db.Exec(deleteSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", deleteSql)
		return err, nil
	}

	// get real transaction hash
	var realHash []TableTempTransactionHash
	if err := database.Db.Find(&realHash).Error; nil != err {
		log.Log.Error(err, " select * from t_temp_transaction_hash fail")
		return err, nil
	}

	var result []string
	for _, one := range realHash {
		result = append(result, one.Txid)
	}
	return nil, result
}

func simpleGetRealTrxHashs(allHash []string) (error, []string) {
	var result []tables.TableTransactionInfo
	if err := database.Db.Model(&tables.TableTransactionInfo{}).Where("`Txid` IN (?)", allHash).Find(&result).Error; nil != err {
		log.Log.Error(err, " select * from t_transaction_info fail")
		return err, nil
	}

	realHash := []string{}
	for _, oneHash := range allHash {
		bFind := false
		for _, oneTransaction := range result {
			if oneHash == oneTransaction.Txid {
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

func (repair *TransactionRepair) repairTransactions(realHashHeight map[string]int32) error {
	allTrxCh := make(chan Transaction, config.TrxGoroutineRuntime)
	go repairTransactionTask(realHashHeight, allTrxCh)

	// get success transaction from channel
	allTrx := []Transaction{}
	for oneTrx := range allTrxCh {
		allTrx = append(allTrx, oneTrx)
	}

	// update state and from
	return updateStateAndFrom(allTrx)
}

func repairTransactionTask(realHashHeight map[string]int32, totalTrxCh chan<- Transaction) error {
	defer func() {
		// close channel
		close(totalTrxCh)

		// handle panic
		if err := recover(); err != nil {
			log.Log.Error(err, " panic occur when run repair transaction task")
		}
	}()

	// save one transaction
	wg := sync.WaitGroup{}
	trxCh := make(chan int32, config.TrxGoroutineRuntime)
	taskFunc := func(height int32, trx string, group *sync.WaitGroup) {
		defer func() {
			trxCh <- height
			group.Done()

			// handle panic
			if err := recover(); err != nil {
				log.Log.Error(err, " panic occur when repair transactions, block height:", height, ", transaction hash:", trx)
				StoreFailTransactionHash(height, trx)
			}
		}()

		if err := GetTransactionAndStoreWithChannel(height, trx, totalTrxCh); nil != err {
			log.Log.Error("repair transaction fail, block height:", height, ", transaction hash:", trx)
			StoreFailTransactionHash(height, trx)
		} else {
			log.Log.Info("repair transaction success, block height:", height, ", transaction hash:", trx)
		}
	}

	index := 0
	for oneTrx, oneHeight := range realHashHeight {
		if index >= config.TrxGoroutineRuntime {
			<-trxCh
		}
		wg.Add(1)
		go taskFunc(oneHeight, oneTrx, &wg)
		index++
	}
	wg.Wait()

	return nil
}

func GetTransactionAndStoreWithChannel(height int32, hash string, trxCh chan<- Transaction) error {
	result, err := jsonrpc.Call(1, "getrawtransaction", []interface{}{hash, true})
	if nil != err {
		log.Log.Error(err, " GetTransactionAndStoreWithChannel jsonrpc call getrawtransaction fail, block height:", height, ", transaction hash:", hash)
		return err
	}
	newTransaction := Transaction{}
	if err := json.Unmarshal(result, &newTransaction); nil != err {
		log.Log.Error(err, " GetTransactionAndStoreWithChannel Unmarshal result to transaction struct fail")
		return err
	}

	// handle transaction
	bHasBlock := false
	if "" != newTransaction.BlockHash {
		bHasBlock = true
		newTransaction.Blockheight = height
	}
	funcSave := SaveTransaction
	if !bHasBlock {
		funcSave = SaveUnconfirmedTransactionToRedis
	}
	resultInfo := funcSave(&newTransaction)

	// write success transaction to channel
	if bHasBlock && nil == resultInfo {
		trxCh <- newTransaction
	}

	return resultInfo
}

func GetTransactionAndStore(height int32, hash string) error {
	result, err := jsonrpc.Call(1, "getrawtransaction", []interface{}{hash, true})
	if nil != err {
		log.Log.Error(err, " GetTransactionAndStore jsonrpc call getrawtransaction fail, block height:", height, ", transaction hash:", hash)
		return err
	}
	newTransaction := Transaction{}
	if err := json.Unmarshal(result, &newTransaction); nil != err {
		log.Log.Error(err, " GetTransactionAndStore Unmarshal result to transaction struct fail")
		return err
	}

	// handle transaction
	bHasBlock := false
	if "" != newTransaction.BlockHash {
		bHasBlock = true
		newTransaction.Blockheight = height
	}
	funcSave := SaveTransaction
	if !bHasBlock {
		funcSave = SaveUnconfirmedTransactionToRedis
	}
	return funcSave(&newTransaction)
}
