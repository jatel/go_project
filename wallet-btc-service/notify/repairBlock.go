package notify

import (
	"bufio"
	"encoding/json"
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
	"github.com/BlockABC/wallet-btc-service/database/tables"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
)

type BlockRepair struct {
	Filename string
	Rw       sync.RWMutex
}

func NewBlockRepair() *BlockRepair {
	baseDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	baseDir = strings.Replace(baseDir, "\\", "/", -1)
	return &BlockRepair{Filename: baseDir + "/repairBlock"}
}

func (repair *BlockRepair) StoreFailHash(hash string) error {
	repair.Rw.Lock()
	defer repair.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(repair.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if nil != err {
		log.Log.Error(err, " store fail block hash open file fail, block hash:", hash)
		return err
	}
	defer fs.Close()

	// write block hash
	buf := bufio.NewWriter(fs)
	if _, err := fmt.Fprintln(buf, hash); nil != err {
		log.Log.Error(err, " store fail block hash write file fail, block hash:", hash)
		return err
	}
	result := buf.Flush()
	if nil == result {
		log.Log.Info("store fail block hash success, block hash:", hash)
	} else {
		log.Log.Error("store fail block hash fail, block hash:", hash)
	}
	return result
}

func (repair *BlockRepair) BatchStoreFailHash(allHash []string) error {
	repair.Rw.Lock()
	defer repair.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(repair.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if nil != err {
		log.Log.Error(err, "batch store block open file fail, block hashs:", allHash)
		return err
	}
	defer fs.Close()

	// write block hash
	buf := bufio.NewWriter(fs)
	for _, oneHash := range allHash {
		if _, err := fmt.Fprintln(buf, oneHash); nil != err {
			log.Log.Error(err, " batch store fail block hash write one block hash to file fail, one block hash:", oneHash)
		}
	}
	result := buf.Flush()
	if nil == result {
		log.Log.Info("batch store fail block hash success, block hashs:", allHash)
	} else {
		log.Log.Error("batch store fail block hash fail, block hashs:", allHash)
	}
	return result
}

func (repair *BlockRepair) RepairAllItems() error {
	if !utility.IsFileExist(repair.Filename) {
		return nil
	}

	repair.Rw.Lock()

	// open file
	fs, err := os.Open(repair.Filename)
	if nil != err {
		log.Log.Error(err, " RepairAllItems open block file fail")
		repair.Rw.Unlock()
		return err
	}

	// read file
	allHash := []string{}
	br := bufio.NewReader(fs)
	for {
		hash, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if nil != err {
			log.Log.Error(err, " read repair block file fail")
			repair.Rw.Unlock()
			fs.Close()
			return err
		}

		bExist := false
		for _, v := range allHash {
			if v == string(hash) {
				bExist = true
				break
			}
		}
		if !bExist {
			allHash = append(allHash, string(hash))
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

	// get real fail block hash
	err, realHashs := getRealBlockHashs(allHash)
	if nil != err {
		repair.BatchStoreFailHash(allHash)
		return err
	}

	if nil == realHashs || len(realHashs) == 0 {
		return nil
	}
	log.Log.Info("repair fail block get real block hash success, real fail block hashs:", realHashs)

	return repair.repairBlocks(realHashs)
}

type TableTempBlockHash struct {
	Hash string `json:"hash"     		gorm:"column:hash;type:char(64);unique"`
}

func (t *TableTempBlockHash) TableName() string {
	return "t_temp_block_hash"
}

var TempBlockHashMutex sync.Mutex

func getRealBlockHashs(allHash []string) (error, []string) {
	// ues in when a few hash
	if len(allHash) <= database.MAX_WITh_IN {
		return simpleGetRealBlockHashs(allHash)
	}

	//lock temp table
	TempBlockHashMutex.Lock()
	defer TempBlockHashMutex.Unlock()

	// create temp table
	if database.Db.HasTable(&TableTempBlockHash{}) {
		if err := database.Db.Exec("DROP TABLE t_temp_block_hash;").Error; err != nil {
			log.Log.Error(err, " drop table t_temp_block_hash fail")
			return err, nil
		}
	}

	if err := database.Db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(&TableTempBlockHash{}).Error; err != nil {
		log.Log.Error(err, " create table t_temp_block_hash fail")
		return err, nil
	}
	// drop table
	defer database.Db.Exec("DROP TABLE t_temp_block_hash;")

	// insert data
	page := len(allHash) / database.MAX_WITH_INSERT
	for i := 0; i < page; i++ {
		begin := i * database.MAX_WITH_INSERT
		end := (i + 1) * database.MAX_WITH_INSERT
		pageHash := allHash[begin:end]
		insertSql := "insert into t_temp_block_hash(hash) values "
		for index, blockhash := range pageHash {
			if 0 == index {
				insertSql += fmt.Sprintf(`('%s')`, blockhash)
			} else {
				insertSql += fmt.Sprintf(`,('%s')`, blockhash)
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
		insertSql := "insert into t_temp_block_hash(hash) values "
		for index, blockhash := range pageHash {
			if 0 == index {
				insertSql += fmt.Sprintf(`('%s')`, blockhash)
			} else {
				insertSql += fmt.Sprintf(`,('%s')`, blockhash)
			}
		}
		insertSql += ";"
		if err := database.Db.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			return err, nil
		}
	}

	// delete repetition data
	deleteSql := "delete t1 from t_temp_block_hash t1, t_block_info t2 where t1.hash=t2.hash;"
	if err := database.Db.Exec(deleteSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", deleteSql)
		return err, nil
	}

	// get real block hash
	var realHash []TableTempBlockHash
	if err := database.Db.Find(&realHash).Error; nil != err {
		log.Log.Error(err, " select * from t_temp_block_hash fail")
		return err, nil
	}

	var result []string
	for _, one := range realHash {
		result = append(result, one.Hash)
	}
	return nil, result
}

func simpleGetRealBlockHashs(allHash []string) (error, []string) {
	var result []tables.TableBlockInfo
	if err := database.Db.Model(&tables.TableBlockInfo{}).Where("`hash` IN (?)", allHash).Find(&result).Error; nil != err {
		log.Log.Error(err, " select * from t_block_info fail")
		return err, nil
	}

	realHash := []string{}
	for _, oneHash := range allHash {
		bFind := false
		for _, oneBlock := range result {
			if oneHash == oneBlock.Hash {
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

func (repair *BlockRepair) repairBlocks(allHash []string) error {
	allBlockCh := make(chan Block, config.Cfg.BtcOpt.BlockGoroutineNum)
	go repairBlockTask(allHash, allBlockCh)

	// get success block from channel
	allBlock := []Block{}
	for oneBlock := range allBlockCh {
		allBlock = append(allBlock, oneBlock)
	}

	// update state and from
	allTrx := []Transaction{}
	for _, oneBlock := range allBlock {
		allTrx = append(allTrx, oneBlock.Tx...)
	}
	return updateStateAndFrom(allTrx)
}

func repairBlockTask(allHash []string, blockCh chan<- Block) error {
	defer func() {
		// close channel
		close(blockCh)

		// handle panic
		if err := recover(); err != nil {
			log.Log.Error(err, " panic occur when run repair block task")
		}
	}()

	// save one block
	wg := sync.WaitGroup{}
	trxCh := make(chan int, config.Cfg.BtcOpt.BlockGoroutineNum)
	taskFunc := func(blockIndex int, block string, group *sync.WaitGroup) {
		defer func() {
			trxCh <- blockIndex
			group.Done()

			// handle panic
			if err := recover(); err != nil {
				log.Log.Error(err, " panic occur when repair blocks, block hash: ", block)
				StoreFailBlockHash(block)
			}
		}()

		if err := GetBlockAndStoreWithChannle(block, blockCh); nil != err {
			log.Log.Error("repair block fail, block hash:", block)
			StoreFailBlockHash(block)
		} else {
			log.Log.Info("repair block success, block hash:", block)
		}
	}
	for index, oneBlock := range allHash {
		if index >= config.Cfg.BtcOpt.BlockGoroutineNum {
			<-trxCh
		}

		wg.Add(1)
		go taskFunc(index, oneBlock, &wg)
	}
	wg.Wait()

	return nil
}

func GetBlockAndStoreNotUpdateStateAndFrom(hash string) error {
	result, err := jsonrpc.Call(1, "getblock", []interface{}{hash, 2})
	if nil != err {
		log.Log.Error(err, " GetBlockAndStoreNotUpdateStateAndFrom jsonrpc call getblock fail, block hash: ", hash)
		return err
	}
	newBlock := Block{}
	if err := json.Unmarshal(result, &newBlock); nil != err {
		log.Log.Error(err, " GetBlockAndStoreNotUpdateStateAndFrom Unmarshal result to block struct fail")
		return err
	}

	return SaveBlockNotUpdateStateAndFrom(&newBlock)
}

func GetBlockAndStoreWithChannle(hash string, blockCh chan<- Block) error {
	result, err := jsonrpc.Call(1, "getblock", []interface{}{hash, 2})
	if nil != err {
		log.Log.Error(err, " GetBlockAndStoreNotUpdateStateAndFrom jsonrpc call getblock fail, block hash: ", hash)
		return err
	}
	newBlock := Block{}
	if err := json.Unmarshal(result, &newBlock); nil != err {
		log.Log.Error(err, " GetBlockAndStoreNotUpdateStateAndFrom Unmarshal result to block struct fail")
		return err
	}

	saveErr := SaveBlockNotUpdateStateAndFrom(&newBlock)
	if nil == saveErr {
		blockCh <- newBlock
	}

	return saveErr
}
