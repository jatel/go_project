package omni

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/common/utility"
	"github.com/BlockABC/wallet-btc-service/database"
)

var transactionRepair *OmniTransactionRepair = NewOmniTransactionRepair()

type OmniTransactionRepair struct {
	Filename string
	Rw       sync.RWMutex
}

func NewOmniTransactionRepair() *OmniTransactionRepair {
	baseDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	baseDir = strings.Replace(baseDir, "\\", "/", -1)
	return &OmniTransactionRepair{Filename: baseDir + "/repairOmniTransaction"}
}

func StoreFailOmniTransactionHah(hash string) error {
	transactionRepair.Rw.Lock()
	defer transactionRepair.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(transactionRepair.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if nil != err {
		log.Log.Error(err, " store fail omni transaction hash open file fail, transaction hash:", hash)
		return err
	}
	defer fs.Close()

	// write transaction hash
	buf := bufio.NewWriter(fs)
	if _, err := fmt.Fprintln(buf, hash); nil != err {
		log.Log.Error(err, " store fail omni transaction hash write file fail, transaction hash:", hash)
		return err
	}
	result := buf.Flush()
	if nil == result {
		log.Log.Info("store fail omni transaction hash success, transaction hash:", hash)
	} else {
		log.Log.Error("store fail omni transaction hash fail, transaction hash:", hash)
	}
	return result
}

func BatchStoreFailTransactionHahs(allHash []string) error {
	transactionRepair.Rw.Lock()
	defer transactionRepair.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(transactionRepair.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if nil != err {
		log.Log.Error(err, " batch store fail omni transaction hash open file fail, transaction hashs:", allHash)
		return err
	}
	defer fs.Close()

	// write transaction hash
	buf := bufio.NewWriter(fs)
	for _, oneHash := range allHash {
		if _, err := fmt.Fprintln(buf, oneHash); nil != err {
			log.Log.Error(err, " batch store fail omni transaction hash write one omni transaction hash to file fail, one transaction hash:", oneHash)
		}
	}

	result := buf.Flush()
	if nil == result {
		log.Log.Info("batch store fail omni transaction hash success, omni transaction hashs:", allHash)
	} else {
		log.Log.Error("batch store fail omni transaction hash fail, omni transaction hashs:", allHash)
	}
	return result
}

func RepairFailTransactionHash() error {
	if !utility.IsFileExist(blockRepair.Filename) {
		return nil
	}

	transactionRepair.Rw.Lock()
	defer transactionRepair.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(transactionRepair.Filename, os.O_RDWR, 0666)
	if nil != err {
		log.Log.Error(err, " repair fail omni transaction hash open file fail")
		return err
	}
	defer fs.Close()

	// read file
	allFileHash := []string{}
	br := bufio.NewReader(fs)
	for {
		hash, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if nil != err {
			log.Log.Error(err, " read repair omni transaction file fail")
			return err
		}

		bExist := false
		for _, v := range allFileHash {
			if v == string(hash) {
				bExist = true
				break
			}
		}

		if !bExist {
			allFileHash = append(allFileHash, string(hash))
		}
	}
	fs.Truncate(0)

	// get real fail transaction hash
	err, realHash := getRealFailTransactionHash(allFileHash)
	if nil != err {
		log.Log.Error(err, " repairFailTransactionHash get real fail transaction hash fail")
		return err
	}

	if nil == realHash || 0 == len(realHash) {
		return nil
	}

	// store fail transaction
	allFailTrxCh := make(chan string, config.TrxGoroutineRuntime)
	go repairFailTransactionWithChannel(realHash, allFailTrxCh)

	// get fail transaction hash
	failTransactionHashFromChannel := []string{}
	for oneTrx := range allFailTrxCh {
		failTransactionHashFromChannel = append(failTransactionHashFromChannel, oneTrx)
	}

	// write fail transaction hash to file
	fs.Seek(0, 0)
	buf := bufio.NewWriter(fs)
	for _, oneFailHash := range failTransactionHashFromChannel {
		if _, err := fmt.Fprintln(buf, oneFailHash); nil != err {
			log.Log.Error(err, " repairFailTransactionHash store one fail omni transaction hash write to file fail, one fail omni transaction hash:", oneFailHash)
		}
	}
	return buf.Flush()
}

type TableTempFailOmniTransactionHash struct {
	Txid string `json:"txid"     		gorm:"column:hash;type:char(64);unique"`
}

func (t *TableTempFailOmniTransactionHash) TableName() string {
	return "t_temp_fail_omni_transaction_hash"
}

var TableTempFailOmniTransactionHashMutex sync.Mutex

func getRealFailTransactionHash(allTransactionHash []string) (error, []string) {
	if nil == allTransactionHash || 0 == len(allTransactionHash) {
		return nil, nil
	}

	// ues in when a few hash
	if len(allTransactionHash) <= database.MAX_WITh_IN {
		return simpleGetOmniTransactionHashs(allTransactionHash)
	}

	// lock temp table
	TableTempFailOmniTransactionHashMutex.Lock()
	defer TableTempFailOmniTransactionHashMutex.Unlock()

	// create temp table
	if database.Db.HasTable(&TableTempFailOmniTransactionHash{}) {
		if err := database.Db.Exec("DROP TABLE t_temp_fail_omni_transaction_hash;").Error; err != nil {
			log.Log.Error(err, " drop table t_temp_fail_omni_transaction_hash fail")
			return err, nil
		}
	}

	if err := database.Db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(&TableTempFailOmniTransactionHash{}).Error; err != nil {
		log.Log.Error(err, " create table t_temp_fail_omni_transaction_hash fail")
		return err, nil
	}
	// drop table
	defer database.Db.Exec("DROP TABLE t_temp_fail_omni_transaction_hash;")

	// insert data
	page := len(allTransactionHash) / database.MAX_WITH_INSERT
	for i := 0; i < page; i++ {
		begin := i * database.MAX_WITH_INSERT
		end := (i + 1) * database.MAX_WITH_INSERT
		pageHash := allTransactionHash[begin:end]
		insertSql := "insert into t_temp_fail_omni_transaction_hash(txid) values "
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
		insertSql := "insert into t_temp_fail_omni_transaction_hash(txid) values "
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
	deleteSql := "delete t1 from t_temp_fail_omni_transaction_hash t1, t_omni_transaction_info t2 where t1.hash=t2.txid;"
	if err := database.Db.Exec(deleteSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", deleteSql)
		return err, nil
	}

	// get real block hash
	var realTableHash []TableTempFailOmniTransactionHash
	if err := database.Db.Find(&realTableHash).Error; nil != err {
		log.Log.Error(err, " select * from t_temp_fail_omni_transaction_hash fail")
		return err, nil
	}

	resultHashs := []string{}
	for _, oneTableHash := range realTableHash {
		resultHashs = append(resultHashs, oneTableHash.Txid)
	}

	return nil, resultHashs
}

func repairFailTransactionWithChannel(realHash []string, failHash chan<- string) error {
	defer func() {
		// handle panic
		if err := recover(); err != nil {
			log.Log.Error(err, " panic occur when repair fail transaction with channel")
		}

		close(failHash)
	}()

	// store transaction
	wg := sync.WaitGroup{}
	trxCh := make(chan int, config.TrxGoroutineRuntime)
	taskFunc := func(oneIndex int, trx string, group *sync.WaitGroup) {
		defer func() {
			trxCh <- oneIndex
			group.Done()

			// handle panic
			if err := recover(); err != nil {
				log.Log.Error(err, " panic occur when repair fail omni transaction, omni transaction hash:", trx)
				failHash <- trx
			}
		}()

		if err := HandleOmniTransaction(trx, false); nil != err {
			log.Log.Error("repair fail omni transaction fail, omni transaction hash:", trx)
			failHash <- trx
		} else {
			log.Log.Info("repair fail omni transaction success, omni transaction hash:", trx)
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
