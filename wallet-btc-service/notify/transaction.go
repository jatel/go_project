package notify

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/database/tables"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
	"github.com/BlockABC/wallet-btc-service/omni"
	"github.com/BlockABC/wallet-btc-service/request"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Transaction struct {
	Txid        string   `json:"txid"`
	Hash        string   `json:"hash"`
	Version     int32    `json:"version"`
	Size        int32    `json:"size"`
	Vsize       int64    `json:"vsize"`
	Weight      int64    `json:"weight"`
	Locktime    int64    `json:"locktime"`
	Vin         []Input  `json:"vin"`
	Vout        []Output `json:"vout"`
	BlockHash   string   `json:"blockhash"`
	BlockTime   int64    `json:"blocktime"`
	Blockheight int32    `json:"blockheight"`
	Hex         string   `json:"hex"`
	ReceiveTime int64    `json:"receivetime"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

type Input struct {
	Coinbase    string    `json:"coinbase"`
	Txid        string    `json:"txid"`
	Vout        int64     `json:"vout"`
	ScriptSig   ScriptSig `json:"scriptSig"`
	Txinwitness []string  `json:"txinwitness"`
	Sequence    int64     `json:"sequence"`
}

type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	Type      string   `json:"type"`
	ReqSigs   int32    `json:"reqSigs"`
	Addresses []string `json:"addresses"`
}
type Output struct {
	Value        float64      `json:"value"`
	N            int64        `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

func isCoinbase(trx *Transaction) bool {
	return 1 == len(trx.Vin) && trx.Vin[0].Coinbase != ""
}

const COIN int64 = 100000000

func batchExecTransactionDbOperate(dbTx *gorm.DB, newTrx *Transaction, object []interface{}) error {
	page := len(object) / database.MAX_WITH_INSERT
	for i := 0; i < page; i++ {
		begin := i * database.MAX_WITH_INSERT
		end := (i + 1) * database.MAX_WITH_INSERT
		pageObject := object[begin:end]

		err, insertSql := generateOperateSql(newTrx, pageObject)
		if nil != err {
			log.Log.Error(err, " generate operate sql fail")
			return err
		}

		if err := dbTx.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			return err
		}
	}

	// others
	if 0 != len(object)%database.MAX_WITH_INSERT {
		begin := page * database.MAX_WITH_INSERT
		pageObject := object[begin:]
		err, insertSql := generateOperateSql(newTrx, pageObject)
		if nil != err {
			log.Log.Error(err, " generate operate sql fail")
			return err
		}

		if err := dbTx.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			return err
		}
	}

	return nil
}

func generateOperateSql(newTrx *Transaction, pageObject []interface{}) (err error, operateSql string) {
	switch pageObject[0].(type) {
	case Input:
		// generate vin insert sql
		operateSql = "insert into t_input_info(blockhash, time, hash, txid, vout, sequence, hex, asm, coinbase) values "
		for index, oneObject := range pageObject {
			newInput := oneObject.(Input)
			if 0 == index {
				operateSql += fmt.Sprintf(`('%s', %d, '%s', '%s', %d, %d, '%s', '%s', '%s')`,
					newTrx.BlockHash, newTrx.BlockTime, newTrx.Txid, newInput.Txid, newInput.Vout, newInput.Sequence, newInput.ScriptSig.Hex, newInput.ScriptSig.Asm, newInput.Coinbase)
			} else {
				operateSql += fmt.Sprintf(`,('%s', %d, '%s', '%s', %d, %d, '%s', '%s', '%s')`,
					newTrx.BlockHash, newTrx.BlockTime, newTrx.Txid, newInput.Txid, newInput.Vout, newInput.Sequence, newInput.ScriptSig.Hex, newInput.ScriptSig.Asm, newInput.Coinbase)
			}
		}

	case Output:
		// generate vout insert sql
		operateSql = "insert into t_output_info(blockhash, time, hash, value, n, hex, asm, type, reqSigs, `to`) values "
		for index, oneObject := range pageObject {
			newOutput := oneObject.(Output)
			var toAddress string = ""
			if nil != newOutput.ScriptPubKey.Addresses {
				toAddress = newOutput.ScriptPubKey.Addresses[0]
			}
			if 0 == index {
				operateSql += fmt.Sprintf(`('%s', %d, '%s', %d, %d, '%s', '%s', '%s', %d, '%s')`,
					newTrx.BlockHash, newTrx.BlockTime, newTrx.Txid, int64(newOutput.Value*100000000), newOutput.N, newOutput.ScriptPubKey.Hex, newOutput.ScriptPubKey.Asm,
					newOutput.ScriptPubKey.Type, newOutput.ScriptPubKey.ReqSigs, toAddress)
			} else {
				operateSql += fmt.Sprintf(`,('%s', %d, '%s', %d, %d, '%s', '%s', '%s', %d, '%s')`,
					newTrx.BlockHash, newTrx.BlockTime, newTrx.Txid, int64(newOutput.Value*100000000), newOutput.N, newOutput.ScriptPubKey.Hex, newOutput.ScriptPubKey.Asm,
					newOutput.ScriptPubKey.Type, newOutput.ScriptPubKey.ReqSigs, toAddress)
			}
		}

	case tables.TableOutputAddressInfo:
		// generate output address
		operateSql = "insert into t_output_address_info(blockhash, hash, n, address) values "
		for index, oneObject := range pageObject {
			oneOutputAddress := oneObject.(tables.TableOutputAddressInfo)
			if 0 == index {
				operateSql += fmt.Sprintf(`('%s', '%s', %d, '%s')`,
					oneOutputAddress.Blockhash, oneOutputAddress.Hash, oneOutputAddress.N, oneOutputAddress.Address)
			} else {
				operateSql += fmt.Sprintf(`,('%s', '%s', %d, '%s')`,
					oneOutputAddress.Blockhash, oneOutputAddress.Hash, oneOutputAddress.N, oneOutputAddress.Address)
			}
		}

	case string:
		// generate vout address info insert sql
		operateSql = "insert into t_transaction_input_output_address_info(blockhash, `time`, txid, `address`, isfrom) values "
		for index, oneObject := range pageObject {
			newAddress := oneObject.(string)
			if 0 == index {
				operateSql += fmt.Sprintf(`('%s', %d, '%s', '%s', %d)`, newTrx.BlockHash, newTrx.BlockTime, newTrx.Txid, newAddress, 0)
			} else {
				operateSql += fmt.Sprintf(`,('%s', %d, '%s', '%s', %d)`, newTrx.BlockHash, newTrx.BlockTime, newTrx.Txid, newAddress, 0)
			}
		}

	default:
		err = errors.New("invalid object type")

	}
	return
}

func SaveTransaction(newTrx *Transaction) error {
	dbTx := database.Db.Begin()
	if err := dbTx.Error; nil != err {
		log.Log.Error(err, " SaveTransaction start database transaction fail")
		return err
	}

	// handle transaction
	var oneCoinbase int8 = 0
	if isCoinbase(newTrx) {
		oneCoinbase = 1
	}
	oneTrx := tables.TableTransactionInfo{
		Iscoinbase:  oneCoinbase,
		Time:        newTrx.BlockTime,
		Blockhash:   newTrx.BlockHash,
		Blockheight: newTrx.Blockheight,
		Txid:        newTrx.Txid,
		Hash:        newTrx.Hash,
		Size:        newTrx.Size,
		Vsize:       newTrx.Vsize,
		Version:     newTrx.Version,
		Locktime:    newTrx.Locktime,
		Weight:      newTrx.Weight,
	}
	if err := dbTx.Create(&oneTrx).Error; nil != err {
		log.Log.Error(err, " insert into t_transaction_info fail when save one block transaction, block height: ", newTrx.Blockheight, ", transaction hash: ", newTrx.Txid)
		dbTx.Rollback()
		return err
	}

	// handle input
	inputObject := make([]interface{}, 0)
	for _, one := range newTrx.Vin {
		inputObject = append(inputObject, one)
	}
	if err := batchExecTransactionDbOperate(dbTx, newTrx, inputObject); nil != err {
		log.Log.Error(err, " handle transaction input fail")
		dbTx.Rollback()
		return err
	}

	// handle output
	outputObject := make([]interface{}, 0)
	for _, one := range newTrx.Vout {
		outputObject = append(outputObject, one)
	}
	if err := batchExecTransactionDbOperate(dbTx, newTrx, outputObject); nil != err {
		log.Log.Error(err, " handle transaction output fail")
		dbTx.Rollback()
		return err
	}

	// handle transaction address info
	outputAddressInfo := []string{}
	outputAddressInfoObject := make([]interface{}, 0)
	for _, oneOutput := range newTrx.Vout {
		var toAddress string = ""
		if nil == oneOutput.ScriptPubKey.Addresses || 0 == len(oneOutput.ScriptPubKey.Addresses) {
			continue
		} else {
			toAddress = oneOutput.ScriptPubKey.Addresses[0]
		}

		bFind := false
		for _, oneExistAddress := range outputAddressInfo {
			if oneExistAddress == toAddress {
				bFind = true
				break
			}
		}
		if !bFind {
			outputAddressInfo = append(outputAddressInfo, toAddress)
			outputAddressInfoObject = append(outputAddressInfoObject, toAddress)
		}
	}

	if err := batchExecTransactionDbOperate(dbTx, newTrx, outputAddressInfoObject); nil != err {
		log.Log.Error(err, " handle transaction input and output address info fail")
		dbTx.Rollback()
		return err
	}

	// handle address
	outputAddressObject := make([]interface{}, 0)
	for _, newOutput := range newTrx.Vout {
		for _, newAddress := range newOutput.ScriptPubKey.Addresses {
			oneAddress := tables.TableOutputAddressInfo{
				Blockhash: newTrx.BlockHash,
				Hash:      newTrx.Txid,
				N:         newOutput.N,
				Address:   newAddress,
			}
			outputAddressObject = append(outputAddressObject, oneAddress)
		}
	}

	if err := batchExecTransactionDbOperate(dbTx, newTrx, outputAddressObject); nil != err {
		log.Log.Error(err, " handle transaction output address fail")
		dbTx.Rollback()
		return err
	}

	// commit
	if err := dbTx.Commit().Error; nil != err {

		log.Log.Error(err, " save transaction commit fail, block height: ", newTrx.Blockheight, ", transaction hash: ", newTrx.Txid)
		dbTx.Rollback()
		return err
	}

	log.Log.Info("insert into t_transaction_info sucess, block height: ", newTrx.Blockheight, ", transaction hash: ", newTrx.Txid)

	if isCoinbase(newTrx) {
		oneTransactionNotify(newTrx)
	}

	return nil
}

var updateStateAndFromMutex sync.Mutex

func updateStateAndFrom(blockTrx []Transaction) error {
	// get table name
	if 0 == len(blockTrx) {
		return nil
	}

	var seedStr string
	for index, oneTrx := range blockTrx {
		oneStr := oneTrx.BlockHash + "_" + oneTrx.Txid
		if 0 == index {
			seedStr += oneStr
		} else {
			seedStr += "," + oneStr
		}
	}

	h := sha1.New()
	io.WriteString(h, seedStr)
	tableName := hex.EncodeToString(h.Sum(nil))

	// drop table if exist
	dropSql := fmt.Sprintf("drop table if exists `%s`;", tableName)
	if err := database.Db.Exec(dropSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", dropSql)
		return err
	}

	// create table
	createSql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (`txid` CHAR(64) NOT NULL DEFAULT '' COMMENT '交易哈希', `vout` BIGINT NOT NULL DEFAULT 0 COMMENT 'vout')ENGINE=INNODB DEFAULT CHARSET=utf8mb4;", tableName)
	if err := database.Db.Exec(createSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", createSql)
		return err
	}
	defer func() { // drop table
		finalDropSql := fmt.Sprintf("DROP TABLE `%s`;", tableName)
		if err := database.Db.Exec(finalDropSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", finalDropSql)
		}
	}()

	// all input
	allInput := []Input{}
	for _, oneTrx := range blockTrx {
		allInput = append(allInput, oneTrx.Vin...)
	}

	// insert data
	page := len(allInput) / database.MAX_WITH_INSERT
	for i := 0; i < page; i++ {
		begin := i * database.MAX_WITH_INSERT
		end := (i + 1) * database.MAX_WITH_INSERT
		pageVin := allInput[begin:end]
		insertSql := fmt.Sprintf("insert into `%s`(txid, vout) values ", tableName)
		for index, newInput := range pageVin {
			if 0 == index {
				insertSql += fmt.Sprintf(`('%s', %d)`, newInput.Txid, newInput.Vout)
			} else {
				insertSql += fmt.Sprintf(`,('%s', %d)`, newInput.Txid, newInput.Vout)
			}
		}
		insertSql += ";"
		if err := database.Db.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			return err
		}
	}

	if 0 != len(allInput)%database.MAX_WITH_INSERT {
		begin := page * database.MAX_WITH_INSERT
		pageVin := allInput[begin:]
		insertSql := fmt.Sprintf("insert into `%s`(txid, vout) values ", tableName)
		for index, newInput := range pageVin {
			if 0 == index {
				insertSql += fmt.Sprintf(`('%s', %d)`, newInput.Txid, newInput.Vout)
			} else {
				insertSql += fmt.Sprintf(`,('%s', %d)`, newInput.Txid, newInput.Vout)
			}
		}
		insertSql += ";"
		if err := database.Db.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			return err
		}
	}

	// In order to solve mysql Lock wait timeout
	updateStateAndFromMutex.Lock()
	defer updateStateAndFromMutex.Unlock()

	// update output state
	updateOutputSql := fmt.Sprintf("update t_output_info t1, `%s` t2 set t1.state=1 where t1.hash=t2.txid and t1.n=t2.vout;", tableName)

	if err := database.Db.Exec(updateOutputSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", updateOutputSql)
		return err
	}

	// update input from and value
	updateInputSql := fmt.Sprintf("update t_input_info t1, t_output_info t2, `%s` t3 set t1.from = t2.to, t1.value = t2.value where t1.txid=t2.hash and t1.vout=t2.n and t1.txid=t3.txid and t1.vout=t3.vout;", tableName)

	if err := database.Db.Exec(updateInputSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", updateInputSql)
		return err
	}

	// insert input address into t_transaction_input_output_address_info
	insertAddressSql := fmt.Sprintf("replace into t_transaction_input_output_address_info (blockhash, `time`, txid, `address`, isfrom) select distinct blockhash, `time`, `hash`, `from`, 1 from t_input_info where (txid, vout) in (select txid, vout from `%s`) and `from` != '';", tableName)
	if err := database.Db.Exec(insertAddressSql).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", insertAddressSql)
		return err
	}

	var transactionInfo string
	for index, oneTrx := range blockTrx {
		if 0 == index {
			transactionInfo += fmt.Sprintf("[{%d, %s}", oneTrx.Blockheight, oneTrx.Hash)
		} else {
			transactionInfo += fmt.Sprintf(", {%d, %s}", oneTrx.Blockheight, oneTrx.Hash)
		}
	}
	transactionInfo += "]"

	log.Log.Info("block transaction update t_output_info state and t_input_info from sucess, table name:", tableName, ", block height and transaction hashs: ", transactionInfo)

	return nil
}

var allUnfmdTrx []string

const (
	MAXTRANSACTIONLEN  = 500
	REVERSETRANSACTION = 100
)

func bitcoindTransactionNotify(c *gin.Context) {
	var newTransaction Transaction
	if err := c.BindJSON(&newTransaction); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	log.Log.Info("New transaction notifications received, transaction hash:", newTransaction.Txid)
	defer func() {
		// handle panic
		if err := recover(); err != nil {
			log.Log.Error(err, " panic occur when bitcoind transaction notify, transaction hash: ", newTransaction.Txid)
		}
	}()

	// handle not arrive unconfirmed omni transaction
	omni.HandleUnarriveOmniTrx()

	// only handle unconfirmed transaction
	if 0 != len(newTransaction.BlockHash) || newTransaction.Blockheight > 0 {
		c.JSON(http.StatusOK, gin.H{"result": "success"})
		return
	}

	// exist
	for _, oneHash := range allUnfmdTrx {
		if oneHash == newTransaction.Txid {
			c.JSON(http.StatusOK, gin.H{"result": "success"})
			return
		}
	}

	// handle transaction
	newTransaction.ReceiveTime = time.Now().Unix()
	if err := SaveUnconfirmedTransactionToRedis(&newTransaction); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	// apned all unconfirmed transaction
	if len(allUnfmdTrx) >= MAXTRANSACTIONLEN {
		allUnfmdTrx = allUnfmdTrx[len(allUnfmdTrx)-MAXTRANSACTIONLEN+REVERSETRANSACTION:]
	}
	allUnfmdTrx = append(allUnfmdTrx, newTransaction.Txid)

	// handle omni transaction info
	omni.HandleOmniTransaction(newTransaction.Txid, true)

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func oneTransactionNotify(newTrx *Transaction) error {
	bHashBlock := false
	if 0 != len(newTrx.BlockHash) || newTrx.Blockheight > 0 {
		if !isCoinbase(newTrx) {
			return nil
		}
		bHashBlock = true
	}

	err, allAddress := request.GetNotifyAddress()
	if nil != err {
		log.Log.Error("oneTransactionNotify get notify address fail, ", err)
		return err
	}

	if 0 == len(allAddress) {
		return nil
	}

	info := request.Push_list{}
	if 0 != len(newTrx.Vin) && !bHashBlock {
		err, bExist, oneRedisTransaction := GetRedisUnconfirmedTransactionByTxid(newTrx.Txid)
		if nil != err || !bExist {
			log.Log.Error(err, " oneTransactionNotify get redis unconfirmed transaction by txid fail, txid:", newTrx.Txid)
			if nil != err {
				return err
			} else {
				return errors.New("unconfirmed transaction not exist")
			}
		} else {
			// input
			for _, oneInput := range oneRedisTransaction.Vin {
				for _, oneAddress := range allAddress {
					if oneAddress.Name == oneInput.Address {
						oneNotify := request.NotifyInfo{
							Chain_type: oneAddress.Chain_type,
							Chain_id:   oneAddress.Chain_id,
							Msg_type:   2,
							Cid:        oneAddress.Cid,
							Msg_id:     fmt.Sprintf("%d_%s_%d", 2, oneInput.Txid, oneInput.Vout),
							Language:   oneAddress.Language,
							Token_name: "BTC",
							Name:       oneAddress.Name,
							Platform:   oneAddress.Platform,
						}
						info.List = append(info.List, oneNotify)
						break
					}
				}
			}

			// output
			for _, oneOutput := range oneRedisTransaction.Vout {
				for _, oneAddress := range allAddress {
					if nil != oneOutput.Addresses && 0 != len(oneOutput.Addresses) && oneOutput.Addresses[0] == oneAddress.Name {
						oneNotify := request.NotifyInfo{
							Chain_type: oneAddress.Chain_type,
							Chain_id:   oneAddress.Chain_id,
							Msg_type:   1,
							Cid:        oneAddress.Cid,
							Msg_id:     fmt.Sprintf("%d_%s_%d", 1, oneRedisTransaction.Txid, oneOutput.N),
							Language:   oneAddress.Language,
							Token_name: "BTC",
							Name:       oneAddress.Name,
							Platform:   oneAddress.Platform,
						}
						info.List = append(info.List, oneNotify)
						break
					}
				}
			}
		}
	}

	for _, oneOutput := range newTrx.Vout {
		for _, oneAddress := range allAddress {
			if nil != oneOutput.ScriptPubKey.Addresses && 0 != len(oneOutput.ScriptPubKey.Addresses) && oneOutput.ScriptPubKey.Addresses[0] == oneAddress.Name {
				oneNotify := request.NotifyInfo{
					Chain_type: oneAddress.Chain_type,
					Chain_id:   oneAddress.Chain_id,
					Msg_type:   1,
					Cid:        oneAddress.Cid,
					Msg_id:     fmt.Sprintf("%d_%s_%d", 1, newTrx.Txid, oneOutput.N),
					Language:   oneAddress.Language,
					Token_name: "BTC",
					Name:       oneAddress.Name,
					Platform:   oneAddress.Platform,
				}
				info.List = append(info.List, oneNotify)
				break
			}
		}
	}

	if 0 == len(info.List) {
		return nil
	}

	return request.NotifyTerminal(info)
}

func HandlePushTransaction(trx string) (error, int64) {
	result, err := jsonrpc.Call(1, "getrawtransaction", []interface{}{trx, true})
	if nil != err {
		log.Log.Error(err, " HandlePushTransaction jsonrpc call getrawtransaction fail, hash: ", trx)
		return err, 0
	}
	newTransaction := Transaction{}
	if err := json.Unmarshal(result, &newTransaction); nil != err {
		log.Log.Error(err, " HandlePushTransaction Unmarshal result to transaction struct fail")
		return err, 0
	}

	// handle transaction
	newTransaction.ReceiveTime = time.Now().Unix()
	go func() {
		if err := SaveUnconfirmedTransactionToRedis(&newTransaction); nil != err {
			log.Log.Error(err, " HandlePushTransaction save unconfirmed transaction to redis fail")
			return
		}

		log.Log.Info("HandlePushTransaction save unconfirmed transaction success, unconfirmed transaction hashs:", trx)
	}()

	return nil, newTransaction.ReceiveTime
}
