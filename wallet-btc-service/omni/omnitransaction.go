package omni

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/database/tables"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
	"github.com/BlockABC/wallet-btc-service/request"
)

type OmniTransaction struct {
	Txid                         string          `json:"txid"`
	ReceiveTime                  int64           `json:"receivetime"`
	Blockhash                    string          `json:"blockhash"`
	Blocktime                    int64           `json:"blocktime"`
	Block                        int32           `json:"block"`
	Positioninblock              int32           `json:"positioninblock"`
	Fee                          string          `json:"fee"`
	Sendingaddress               string          `json:"sendingaddress"`
	Purchases                    []OmniPurchase  `json:"purchases"`
	Referenceaddress             string          `json:"referenceaddress"`
	Type_int                     uint64          `json:"type_int"`
	Type                         string          `json:"type"`
	Ismine                       bool            `json:"ismine"`
	Version                      uint64          `json:"version"`
	Valid                        bool            `json:"valid"`
	Invalidreason                string          `json:"invalidreason"`
	Purchasedpropertyid          int64           `json:"purchasedpropertyid"`
	Purchasedpropertyname        string          `json:"purchasedpropertyname"`
	Purchasedpropertydivisible   bool            `json:"purchasedpropertydivisible"`
	Purchasedtokens              string          `json:"purchasedtokens"`
	Issuertokens                 string          `json:"issuertokens"`
	Propertyid                   uint64          `json:"propertyid"`
	Divisible                    bool            `json:"divisible"`
	Amount                       string          `json:"amount"`
	Totalstofee                  string          `json:"totalstofee"`
	Recipients                   []OmniRecipient `json:"recipients"`
	Ecosystem                    string          `json:"ecosystem"`
	Subsends                     []OmniSubsend   `json:"subsends"`
	Bitcoindesired               string          `json:"bitcoindesired"`
	Timelimit                    uint8           `json:"timelimit"`
	Feerequired                  string          `json:"feerequired"`
	Action                       string          `json:"action"`
	Propertyidforsale            uint64          `json:"propertyidforsale"`
	Propertyidforsaleisdivisible bool            `json:"propertyidforsaleisdivisible"`
	Amountforsale                string          `json:"amountforsale"`
	Propertyiddesired            uint64          `json:"propertyiddesired"`
	Propertyiddesiredisdivisible bool            `json:"propertyiddesiredisdivisible"`
	Amountdesired                string          `json:"amountdesired"`
	Unitprice                    string          `json:"unitprice"`
	Amountremaining              string          `json:"amountremaining"`
	Amounttofill                 string          `json:"amounttofill"`
	Status                       string          `json:"status"`
	Canceltxid                   string          `json:"canceltxid"`
	Matches                      []OmniTrade     `json:"matches"`
	Cancelledtransactions        []OmniCancel    `json:"cancelledtransactions"`
	Propertytype                 string          `json:"propertytype"`
	Category                     string          `json:"category"`
	Subcategory                  string          `json:"subcategory"`
	Propertyname                 string          `json:"propertyname"`
	Data                         string          `json:"data"`
	Url                          string          `json:"url"`
	Tokensperunit                string          `json:"tokensperunit"`
	Deadline                     int64           `json:"deadline"`
	Earlybonus                   uint8           `json:"earlybonus"`
	Percenttoissuer              uint8           `json:"percenttoissuer"`
	Featureid                    uint64          `json:"featureid"`
	Activationblock              uint64          `json:"activationblock"`
	Minimumversion               uint64          `json:"minimumversion"`
}

type OmniPurchase struct {
	Vout             int64  `json:"vout"`
	Amountpaid       string `json:"amountpaid"`
	Ismine           bool   `json:"ismine"`
	Referenceaddress string `json:"referenceaddress"`
	Propertyid       int64  `json:"propertyid"`
	Amountbought     string `json:"amountbought"`
	Valid            bool   `json:"valid"`
}

type OmniRecipient struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

type OmniSubsend struct {
	Propertyid uint64 `json:"propertyid"`
	Divisible  bool   `json:"divisible"`
	Amount     string `json:"amount"`
}

type OmniTrade struct {
	Txid           string `json:"txid"`
	Block          int    `json:"block"`
	Address        string `json:"address"`
	Amountsold     string `json:"amountsold"`
	Amountreceived string `json:"amountreceived"`
	Tradingfee     string `json:"tradingfee"`
}

type OmniCancel struct {
	Txid             string `json:"txid"`
	Propertyid       int64  `json:"propertyid"`
	Amountunreserved string `json:"amountunreserved"`
}

func ConvertTransactionToTableOmniTransactionInfo(newTransaction *OmniTransaction) *tables.TableOmniTransactionInfo {
	var ismine_int int8 = 0
	if newTransaction.Ismine {
		ismine_int = 1
	}
	var valid_int int8 = 0
	if newTransaction.Valid {
		valid_int = 1
	}

	var purchasedpropertydivisible_int int8 = 0
	if newTransaction.Purchasedpropertydivisible {
		purchasedpropertydivisible_int = 1
	}

	var divisible_int int8 = 0
	if newTransaction.Divisible {
		divisible_int = 1
	}

	var propertyidforsaleisdivisible_int int8 = 0
	if newTransaction.Propertyidforsaleisdivisible {
		propertyidforsaleisdivisible_int = 1
	}

	var propertyiddesiredisdivisible_int int8 = 0
	if newTransaction.Propertyiddesiredisdivisible {
		propertyiddesiredisdivisible_int = 1
	}

	return &tables.TableOmniTransactionInfo{
		Txid:                         newTransaction.Txid,
		Blockhash:                    newTransaction.Blockhash,
		Blocktime:                    newTransaction.Blocktime,
		Block:                        newTransaction.Block,
		Positioninblock:              newTransaction.Positioninblock,
		Fee:                          newTransaction.Fee,
		Sendingaddress:               newTransaction.Sendingaddress,
		Referenceaddress:             newTransaction.Referenceaddress,
		Type_int:                     newTransaction.Type_int,
		Type:                         newTransaction.Type,
		Ismine:                       ismine_int,
		Version:                      newTransaction.Version,
		Valid:                        valid_int,
		Invalidreason:                newTransaction.Invalidreason,
		Purchasedpropertyid:          newTransaction.Purchasedpropertyid,
		Purchasedpropertyname:        newTransaction.Purchasedpropertyname,
		Purchasedpropertydivisible:   purchasedpropertydivisible_int,
		Purchasedtokens:              newTransaction.Purchasedtokens,
		Issuertokens:                 newTransaction.Issuertokens,
		Propertyid:                   newTransaction.Propertyid,
		Divisible:                    divisible_int,
		Amount:                       newTransaction.Amount,
		Totalstofee:                  newTransaction.Totalstofee,
		Ecosystem:                    newTransaction.Ecosystem,
		Bitcoindesired:               newTransaction.Bitcoindesired,
		Timelimit:                    newTransaction.Timelimit,
		Feerequired:                  newTransaction.Feerequired,
		Action:                       newTransaction.Action,
		Propertyidforsale:            newTransaction.Propertyidforsale,
		Propertyidforsaleisdivisible: propertyidforsaleisdivisible_int,
		Amountforsale:                newTransaction.Amountforsale,
		Propertyiddesired:            newTransaction.Propertyiddesired,
		Propertyiddesiredisdivisible: propertyiddesiredisdivisible_int,
		Amountdesired:                newTransaction.Amountdesired,
		Unitprice:                    newTransaction.Unitprice,
		Amountremaining:              newTransaction.Amountremaining,
		Amounttofill:                 newTransaction.Amounttofill,
		Status:                       newTransaction.Status,
		Canceltxid:                   newTransaction.Canceltxid,
		Propertytype:                 newTransaction.Propertytype,
		Category:                     newTransaction.Category,
		Subcategory:                  newTransaction.Subcategory,
		Propertyname:                 newTransaction.Propertyname,
		Data:                         newTransaction.Data,
		Url:                          newTransaction.Url,
		Tokensperunit:                newTransaction.Tokensperunit,
		Deadline:                     newTransaction.Deadline,
		Earlybonus:                   newTransaction.Earlybonus,
		Percenttoissuer:              newTransaction.Percenttoissuer,
		Featureid:                    newTransaction.Featureid,
		Activationblock:              newTransaction.Activationblock,
		Minimumversion:               newTransaction.Minimumversion,
	}
}

func SaveOmniTransaction(newTransaction *OmniTransaction) error {
	dbTx := database.Db.Begin()
	if err := dbTx.Error; nil != err {
		log.Log.Error(err, " start database transaction fail")
		return err
	}

	oneOmniTrx := *ConvertTransactionToTableOmniTransactionInfo(newTransaction)

	if err := dbTx.Create(&oneOmniTrx).Error; nil != err {
		log.Log.Error(err, " insert into t_omni_transaction_info fail, omni transaction hash: ", oneOmniTrx.Txid)
		dbTx.Rollback()
		return err
	}

	// purchases
	if 0 != len(newTransaction.Purchases) {
		insertSql := generatePurchasesSql(newTransaction, newTransaction.Purchases)
		if err := dbTx.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			dbTx.Rollback()
			return err
		}
	}

	// recipients
	if 0 != len(newTransaction.Recipients) {
		insertSql := generateRecipientsSql(newTransaction, newTransaction.Recipients)
		if err := dbTx.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			dbTx.Rollback()
			return err
		}
	}

	// subsends
	if 0 != len(newTransaction.Subsends) {
		insertSql := generateSubsendSql(newTransaction, newTransaction.Subsends)
		if err := dbTx.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			dbTx.Rollback()
			return err
		}
	}

	// trades
	if 0 != len(newTransaction.Matches) {
		insertSql := generateTradesSql(newTransaction, newTransaction.Matches)
		if err := dbTx.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			dbTx.Rollback()
			return err
		}
	}

	// cancels
	if 0 != len(newTransaction.Cancelledtransactions) {
		insertSql := generateCancelsSql(newTransaction, newTransaction.Cancelledtransactions)
		if err := dbTx.Exec(insertSql).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", insertSql)
			dbTx.Rollback()
			return err
		}
	}

	// commit
	if err := dbTx.Commit().Error; nil != err {
		log.Log.Error(err, " save transaction commit fail")
		dbTx.Rollback()
		return err
	}

	return nil
}

func generatePurchasesSql(newTransaction *OmniTransaction, purchases []OmniPurchase) string {
	insertSql := "insert into t_omni_purchase_info(txid, vout, amountpaid, ismine, referenceaddress, propertyid, amountbought, valid) values "
	for index, onePurchase := range purchases {
		var valid_int int8 = 0
		if onePurchase.Valid {
			valid_int = 1
		}
		var ismine_int int8 = 0
		if onePurchase.Ismine {
			ismine_int = 1
		}
		if 0 == index {
			insertSql += fmt.Sprintf(`('%s', %d, '%s', %d, '%s', %d, '%s', %d)`, newTransaction.Txid, onePurchase.Vout, onePurchase.Amountpaid,
				ismine_int, onePurchase.Referenceaddress, onePurchase.Propertyid, onePurchase.Amountbought, valid_int)
		} else {
			insertSql += fmt.Sprintf(`,('%s', %d, '%s', %d, '%s', %d, '%s', %d)`, newTransaction.Txid, onePurchase.Vout, onePurchase.Amountpaid,
				ismine_int, onePurchase.Referenceaddress, onePurchase.Propertyid, onePurchase.Amountbought, valid_int)
		}
	}
	insertSql += ";"
	return insertSql
}

func generateRecipientsSql(newTransaction *OmniTransaction, recipients []OmniRecipient) string {
	insertSql := "insert into t_omni_recipient_info(txid, address, amount) values "
	for index, oneRecipient := range recipients {
		if 0 == index {
			insertSql += fmt.Sprintf(`('%s', '%s', '%s')`, newTransaction.Txid, oneRecipient.Address, oneRecipient.Amount)
		} else {
			insertSql += fmt.Sprintf(`,('%s', '%s', '%s')`, newTransaction.Txid, oneRecipient.Address, oneRecipient.Amount)
		}
	}
	insertSql += ";"
	return insertSql
}

func generateSubsendSql(newTransaction *OmniTransaction, subsends []OmniSubsend) string {
	insertSql := "insert into t_omni_subsend_info(txid, propertyid, divisible, amount) values "
	for index, onesubsend := range subsends {
		var divisible_int int8 = 0
		if onesubsend.Divisible {
			divisible_int = 1
		}
		if 0 == index {
			insertSql += fmt.Sprintf(`('%s', %d, %d, '%s')`, newTransaction.Txid, onesubsend.Propertyid, divisible_int, onesubsend.Amount)
		} else {
			insertSql += fmt.Sprintf(`,('%s', %d, %d, '%s')`, newTransaction.Txid, onesubsend.Propertyid, divisible_int, onesubsend.Amount)
		}
	}
	insertSql += ";"
	return insertSql
}

func generateTradesSql(newTransaction *OmniTransaction, trades []OmniTrade) string {
	insertSql := "insert into t_omni_trade_info(hash, txid, block, address, amountsold, amountreceived, tradingfee) values "
	for index, oneTrade := range trades {
		if 0 == index {
			insertSql += fmt.Sprintf(`('%s', '%s', %d, '%s', '%s', '%s', '%s')`, newTransaction.Txid, oneTrade.Txid, oneTrade.Block,
				oneTrade.Address, oneTrade.Amountsold, oneTrade.Amountreceived, oneTrade.Tradingfee)
		} else {
			insertSql += fmt.Sprintf(`,('%s', '%s', %d, '%s', '%s', '%s', '%s')`, newTransaction.Txid, oneTrade.Txid, oneTrade.Block,
				oneTrade.Address, oneTrade.Amountsold, oneTrade.Amountreceived, oneTrade.Tradingfee)
		}
	}
	insertSql += ";"
	return insertSql
}

func generateCancelsSql(newTransaction *OmniTransaction, cancels []OmniCancel) string {
	insertSql := "insert into t_omni_cancel_info(hash, txid, propertyid, amountunreserved) values "
	for index, oneCancel := range cancels {
		if 0 == index {
			insertSql += fmt.Sprintf(`('%s', '%s', %d, '%s')`, newTransaction.Txid, oneCancel.Txid, oneCancel.Propertyid, oneCancel.Amountunreserved)
		} else {
			insertSql += fmt.Sprintf(`,('%s', '%s', %d, '%s')`, newTransaction.Txid, oneCancel.Txid, oneCancel.Propertyid, oneCancel.Amountunreserved)
		}
	}
	insertSql += ";"
	return insertSql
}

var notArriveOmniTransaction []string
var notArriveOmniTransactionMutex sync.Mutex

func HandleOmniTransaction(oneHash string, bUnfimd bool) error {
	if 0 == len(oneHash) {
		return nil
	}

	// get omni transaction
	result, err := jsonrpc.OmniCall(1, "omni_gettransaction", []interface{}{oneHash})
	if nil != err {
		log.Log.Error(err, " HandleOmniTransaction jsonrpc OmniCall omni_gettransaction fail, transaction hash: ", oneHash)
		if IsNotOmniTransaction(err) {
			return nil
		}

		if bUnfimd && IsOmniNotFound(err) {
			bExist := false
			notArriveOmniTransactionMutex.Lock()
			for _, one := range notArriveOmniTransaction {
				if one == oneHash {
					bExist = true
					break
				}
			}
			if !bExist {
				notArriveOmniTransaction = append(notArriveOmniTransaction, oneHash)
			}
			notArriveOmniTransactionMutex.Unlock()
			return nil
		}

		return err
	}

	// unmarshal omni transaction
	newTransaction := OmniTransaction{}
	if err := json.Unmarshal(result, &newTransaction); nil != err {
		log.Log.Error(err, " HandleOmniTransaction Unmarshal result to OmniTransaction struct fail")
		return err
	}

	// save transaction
	taskFunc := SaveOmniTransaction
	if bUnfimd {
		taskFunc = SaveUnconfirmedOmniTransactionToRedis
		if 0 != len(newTransaction.Blockhash) || newTransaction.Block > 0 {
			log.Log.Error(newTransaction.Txid, " is not a unconfirmed omni transaction")
			return nil
		}
		newTransaction.ReceiveTime = time.Now().Unix()
	}

	if err := taskFunc(&newTransaction); nil != err {
		log.Log.Error(err, " HandleOmniTransaction save omni transaction fail, transaction hash:", oneHash)
		return err
	}

	// handle not arrive omni transaction
	notArriveOmniTransactionMutex.Lock()
	for index, oneHash := range notArriveOmniTransaction {
		if oneHash == newTransaction.Txid {
			notArriveOmniTransaction = append(notArriveOmniTransaction[:index], notArriveOmniTransaction[index+1:]...)
			break
		}
	}
	notArriveOmniTransactionMutex.Unlock()

	return nil
}

func IsNotOmniTransaction(err error) bool {
	type ResultErr struct {
		Code    int
		Message string
	}
	type resultRpcErr struct {
		Result string
		Error  ResultErr
		Id     int
	}
	info := err.Error()
	var resultInfo resultRpcErr
	err = json.Unmarshal([]byte(info), &resultInfo)
	if nil == err {
		if -5 == resultInfo.Error.Code &&
			(OmniErrorInfo(MP_TX_IS_NOT_OMNI_PROTOCOL) == resultInfo.Error.Message) {
			return true
		}
	}

	return false
}

func IsOmniNotFound(err error) bool {
	type ResultErr struct {
		Code    int
		Message string
	}
	type resultRpcErr struct {
		Result string
		Error  ResultErr
		Id     int
	}
	info := err.Error()
	var resultInfo resultRpcErr
	err = json.Unmarshal([]byte(info), &resultInfo)
	if nil == err {
		if -5 == resultInfo.Error.Code &&
			(OmniErrorInfo(MP_TX_NOT_FOUND) == resultInfo.Error.Message) {
			return true
		}
	}

	return false
}

func HandleUnarriveOmniTrx() {
	// copy slice
	allHash := make([]string, 0)
	notArriveOmniTransactionMutex.Lock()
	num := copy(allHash, notArriveOmniTransaction)
	notArriveOmniTransactionMutex.Unlock()

	// handle not arrive unconfirmed omni transaction
	if 0 == num {
		return
	}
	for _, oneHash := range allHash {
		HandleOmniTransaction(oneHash, true)
	}
}

func oneOmniTransactionNotify(newTrx *OmniTransaction) error {
	err, allAddress := request.GetNotifyAddress()
	if nil != err {
		log.Log.Error("oneOmniTransactionNotify get notify address fail, ", err)
		return err
	}

	for i := 0; i < len(allAddress); i++ {
		if 3 == allAddress[i].Platform {
			allAddress = append(allAddress[:i], allAddress[i+1:]...)
		}
	}

	if 0 == len(allAddress) {
		return nil
	}

	info := request.Push_list{}
	for _, oneAddress := range allAddress {
		if 0 != len(newTrx.Sendingaddress) && newTrx.Sendingaddress == oneAddress.Name {
			oneNotify := request.NotifyInfo{
				Chain_type: oneAddress.Chain_type,
				Chain_id:   oneAddress.Chain_id,
				Msg_type:   2,
				Cid:        oneAddress.Cid,
				Msg_id:     fmt.Sprintf("%s_from", newTrx.Txid),
				Language:   oneAddress.Language,
				Token_name: "USDT",
				Name:       oneAddress.Name,
			}
			info.List = append(info.List, oneNotify)
			break
		}
	}

	for _, oneAddress := range allAddress {
		if 0 != len(newTrx.Referenceaddress) && newTrx.Referenceaddress == oneAddress.Name {
			oneNotify := request.NotifyInfo{
				Chain_type: oneAddress.Chain_type,
				Chain_id:   oneAddress.Chain_id,
				Msg_type:   1,
				Cid:        oneAddress.Cid,
				Msg_id:     fmt.Sprintf("%s_to", newTrx.Txid),
				Language:   oneAddress.Language,
				Token_name: "USDT",
				Name:       oneAddress.Name,
			}
			info.List = append(info.List, oneNotify)
			break
		}
	}

	if 0 == len(info.List) {
		return nil
	}

	return request.NotifyTerminal(info)
}
