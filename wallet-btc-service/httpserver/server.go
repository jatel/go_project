package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/database/tables"
	"github.com/BlockABC/wallet-btc-service/innererror"
	"github.com/BlockABC/wallet-btc-service/jsonrpc"
	"github.com/BlockABC/wallet-btc-service/notify"
	"github.com/BlockABC/wallet-btc-service/omni"
	"github.com/BlockABC/wallet-btc-service/request"
	"github.com/gin-gonic/gin"
)

func StartHttpServer(cfg *config.Config, ctx context.Context) (err error) {
	//get router instance
	router := gin.Default()

	router.Use(cors())

	// handle get address info
	router.POST("/address", getAddressInfo)

	// handle get address transactions
	router.POST("/address/transactions", getAddressTransactions)

	// handle get transaction info
	router.GET("/transaction/:txid", getTransaction)

	// handle get address unspent output
	router.POST("/address/unspents", getUnspents)

	// handle send raw transaction
	router.POST("/send_raw_transaction", sendRawTransaction)

	// handle get recommended fee rates
	router.GET("/recommended_fee_rates", getFeeRate)

	// balance
	router.POST("/address/usdt/balances", getUsdtBalances)

	// record
	router.POST("/address/usdt/transactions", getAddressUsdtTransactions)

	// handle get transaction info
	router.GET("/usdt/:txid", getUsdtTransaction)

	// listen and server
	http.ListenAndServe(cfg.BtcOpt.ApiServerAddress, router)
	return nil
}

func getAddressInfo(c *gin.Context) {
	type info struct {
		Address string `json:"address"`
		Tx_used bool   `json:"tx_used"`
		Balance string `json:"balance"`
	}

	type data struct {
		Decimals  int     `json:"decimals"`
		Usd_price float64 `json:"usd_price"`
		Info      []info  `json:"info"`
	}

	type msg struct {
		Errno  int    `json:"errno"`
		Errmsg string `json:"errmsg"`
		Data   data   `json:"data"`
	}

	var noError innererror.ErrCode = innererror.ErrNoError
	resultMsg := msg{
		Errno:  noError.Value(),
		Errmsg: noError.ErrorInfo(),
		Data:   data{},
	}

	// get address list
	var addressBatch []string
	if err := c.BindJSON(&addressBatch); nil != err {
		var errorCode innererror.ErrCode = innererror.ErrDecodeError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	// real address
	addressReal := make([]string, 0)
	for _, oneAddress := range addressBatch {
		if !IsValidAddress(oneAddress) {
			continue
		}

		bExist := false
		for _, oneReal := range addressReal {
			if oneAddress == oneReal {
				bExist = true
				break
			}
		}

		if !bExist {
			addressReal = append(addressReal, oneAddress)
		}
	}

	// parameter invalid
	if 0 == len(addressReal) {
		var errorCode innererror.ErrCode = innererror.ErrInvalidParaError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	log.Log.Info("getAddressInfo parameter address:", addressReal)

	// get count
	addressUsed := make(map[string]bool)
	for _, oneAddress := range addressReal {
		var outputAddress []tables.TableOutputAddressInfo
		if err := database.Db.Where("address = ?", oneAddress).Limit(1).Find(&outputAddress).Error; nil != err {
			log.Log.Error(err, "  select * from t_output_address_info fail")
			var errorCode innererror.ErrCode = innererror.ErrSQLError
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusOK, resultMsg)
			return
		}
		if 0 != len(outputAddress) {
			addressUsed[oneAddress] = true
		} else {
			addressUsed[oneAddress] = false
		}
	}

	err, unfmdUsed := notify.GetRedisUsedInfoByAddress(addressReal)
	if nil != err {
		log.Log.Error(err, " GetRedisUsedInfoByAddress fail")
		var errorCode innererror.ErrCode = innererror.ErrUnknown
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	for unfmdAddress, unfmdUsed := range unfmdUsed {
		if unfmdUsed {
			addressUsed[unfmdAddress] = unfmdUsed
		}
	}

	// balance
	address := ""
	bFirst := true
	for oneAddress, bUsed := range addressUsed {
		if bUsed {
			if bFirst {
				address += "'" + oneAddress + "'"
				bFirst = false
			} else {
				address += ",'" + oneAddress + "'"
			}
		}
	}
	type OneBalance struct {
		Balance int64
		Address string
	}
	balance := make([]OneBalance, 0)
	if "" != address {
		selectBalanceSql := fmt.Sprintf("select sum(value) as balance, `to` as address from t_output_info  where `to` in (%s) and state=0 group by `to`", address)
		log.Log.Info("getAddressInfo select balance sql:", selectBalanceSql)
		if err := database.Db.Raw(selectBalanceSql).Scan(&balance).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", selectBalanceSql)
			var errorCode innererror.ErrCode = innererror.ErrSQLError
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusOK, resultMsg)
			return
		}
	}

	// unconfirmed balance
	err, unfmdBalance := notify.GetRedisUnconfirmedTransactionBalanceByAddress(addressReal)
	if nil != err {
		log.Log.Error(err, " GetRedisUnconfirmedTransactionBalanceByAddress fail")
		var errorCode innererror.ErrCode = innererror.ErrUnknown
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	for oneUnfmdAddress, oneUnfmdBalance := range unfmdBalance {
		bFind := false
		for index, oneBalance := range balance {
			if oneUnfmdAddress == oneBalance.Address {
				balance[index].Balance += oneUnfmdBalance
				bFind = true
				break
			}
		}

		if !bFind {
			balance = append(balance, OneBalance{oneUnfmdBalance, oneUnfmdAddress})
		}
	}

	// price
	ids := []string{"bitcoin"}
	err, resultPrice := request.GetPrice(ids)
	if nil != err {
		var errorCode innererror.ErrCode = innererror.ErrRPCCallError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}
	price := resultPrice["bitcoin"]
	resultMsg.Data.Usd_price = price
	resultMsg.Data.Decimals = 8

	// result
	for oneAddress, bUsed := range addressUsed {
		oneData := info{
			Address: oneAddress,
			Tx_used: bUsed,
			Balance: "0",
		}

		if bUsed {
			for _, oneBalance := range balance {
				if oneAddress == oneBalance.Address {
					oneData.Balance = fmt.Sprintf("%d", oneBalance.Balance)
					break
				}
			}
		}
		resultMsg.Data.Info = append(resultMsg.Data.Info, oneData)
	}

	log.Log.Info("getAddressInfo result:", resultMsg)

	c.JSON(http.StatusOK, resultMsg)
	return
}

func getAddressTransactions(c *gin.Context) {
	// message type
	type pagination struct {
		Size int `json:"size"`
		Page int `json:"page"`
	}

	type data struct {
		Pagination pagination        `json:"pagination"`
		List       []transactionInfo `json:"list"`
	}

	type msg struct {
		Errno  int    `json:"errno"`
		Errmsg string `json:"errmsg"`
		Data   data   `json:"data"`
	}

	// init message
	var noError innererror.ErrCode = innererror.ErrNoError
	resultMsg := msg{
		Errno:  noError.Value(),
		Errmsg: noError.ErrorInfo(),
		Data:   data{},
	}

	// request info
	type request struct {
		Size      int
		Page      int
		Addresses []string
	}

	var oneRequest request
	if err := c.BindJSON(&oneRequest); nil != err {
		var errorCode innererror.ErrCode = innererror.ErrDecodeError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	// real address
	addressReal := make([]string, 0)
	for _, oneAddress := range oneRequest.Addresses {
		if !IsValidAddress(oneAddress) {
			continue
		}

		bExist := false
		for _, oneReal := range addressReal {
			if oneAddress == oneReal {
				bExist = true
				break
			}
		}

		if !bExist {
			addressReal = append(addressReal, oneAddress)
		}
	}

	// parameter invalid
	if 0 == len(addressReal) {
		var errorCode innererror.ErrCode = innererror.ErrInvalidParaError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	log.Log.Info("getAddressTransactions parameter address:", addressReal)

	var address string
	for index, oneAddress := range addressReal {
		if 0 == index {
			address += `'` + oneAddress + `'`
		} else {
			address += `,'` + oneAddress + `'`
		}
	}

	// transaction count
	type hashAndTime struct {
		Txid   string
		Txtime int64
	}

	// unconfirmed transaction count
	err, unfmdTransaction := notify.GetRedisUnconfirmedTransactionByAddress(addressReal)
	if nil != err {
		log.Log.Error(err, " GetRedisUnconfirmedTransactionByAddress fail")
		var errorCode innererror.ErrCode = innererror.ErrUnknown
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	// get transaction
	pageInfo := pagination{
		Size: oneRequest.Size,
		Page: oneRequest.Page,
	}
	resultMsg.Data.Pagination = pageInfo

	if oneRequest.Page <= 0 {
		log.Log.Error(fmt.Sprintf("page number %d, size %d out of range", oneRequest.Page, oneRequest.Size))
		var errorCode innererror.ErrCode = innererror.ErrOutOfRangeError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	// all unconfirmed and confirmed transactions
	txUnfmdRedisTrx := []notify.RedisTransaction{}
	txIds := []string{}
	begin := (oneRequest.Page - 1) * oneRequest.Size
	end := oneRequest.Page * oneRequest.Size
	limitBegin := 0
	limitEnd := 0
	if end <= len(unfmdTransaction) {
		for i := begin; i < end; i++ {
			txUnfmdRedisTrx = append(txUnfmdRedisTrx, unfmdTransaction[i])
		}
	} else if begin < len(unfmdTransaction) {
		for i := begin; i < len(unfmdTransaction); i++ {
			txUnfmdRedisTrx = append(txUnfmdRedisTrx, unfmdTransaction[i])
		}
		limitBegin = 0
		limitEnd = oneRequest.Size - len(txUnfmdRedisTrx)
	} else {
		limitBegin = begin - len(unfmdTransaction)
		limitEnd = oneRequest.Size
	}

	if 0 != limitEnd {
		type transactionId struct {
			Txid string
		}
		allTransactionId := []transactionId{}
		txidSelectSql := fmt.Sprintf("select distinct txid from t_transaction_input_output_address_info where `address` in (%s) order by time desc limit %d, %d;", address, limitBegin, limitEnd)
		log.Log.Info("getAddressTransactions select transaction id sql:", txidSelectSql)
		if err := database.Db.Raw(txidSelectSql).Scan(&allTransactionId).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", txidSelectSql)
			var errorCode innererror.ErrCode = innererror.ErrSQLError
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusOK, resultMsg)
			return
		}
		for _, oneTableTxid := range allTransactionId {
			txIds = append(txIds, oneTableTxid.Txid)
		}
	}

	// unconfirmed transaction
	list := []transactionInfo{}
	if 0 != len(txUnfmdRedisTrx) {
		for _, oneRedis := range txUnfmdRedisTrx {
			list = append(list, convertRedisTransactionToTransactionInfo(&oneRedis))
		}
	}

	// transaction
	if 0 != len(txIds) {
		var strTxIds string
		for index, oneTxId := range txIds {
			if 0 == index {
				strTxIds += "'" + oneTxId + "'"
			} else {
				strTxIds += ",'" + oneTxId + "'"
			}
		}

		var pageTransaction []tables.TableTransactionInfo
		txSelectSql := fmt.Sprintf("select * from t_transaction_info where txid in (%s);", strTxIds)
		log.Log.Info("getAddressTransactions select transaction sql:", txSelectSql)
		if err := database.Db.Raw(txSelectSql).Scan(&pageTransaction).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", txSelectSql)
			var errorCode innererror.ErrCode = innererror.ErrSQLError
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusOK, resultMsg)
			return
		}

		var pageInput []tables.TableInputInfo
		inputSelectSql := fmt.Sprintf("select * from t_input_info where hash in (%s);", strTxIds)
		log.Log.Info("getAddressTransactions select transaction input sql:", inputSelectSql)
		if err := database.Db.Raw(inputSelectSql).Scan(&pageInput).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", inputSelectSql)
			var errorCode innererror.ErrCode = innererror.ErrSQLError
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusOK, resultMsg)
			return
		}

		var pageOutput []tables.TableOutputInfo
		outputSelectSql := fmt.Sprintf("select * from t_output_info where hash in (%s);", strTxIds)
		log.Log.Info("getAddressTransactions select transaction output sql:", outputSelectSql)
		if err := database.Db.Raw(outputSelectSql).Scan(&pageOutput).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", outputSelectSql)
			var errorCode innererror.ErrCode = innererror.ErrSQLError
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusOK, resultMsg)
			return
		}

		allTransactionInfo := convertToTransactionInfo(pageTransaction, pageInput, pageOutput)

		for _, onrTxid := range txIds {
			for _, oneTrx := range allTransactionInfo {
				if onrTxid == oneTrx.Txid {
					list = append(list, oneTrx)
				}
			}
		}
	}

	resultMsg.Data.List = list

	log.Log.Info("getAddressTransactions result:", resultMsg)

	c.JSON(http.StatusOK, resultMsg)
	return
}

func getTransaction(c *gin.Context) {
	// message type
	type msg struct {
		Errno  int             `json:"errno"`
		Errmsg string          `json:"errmsg"`
		Data   transactionInfo `json:"data"`
	}

	// init message
	var noError innererror.ErrCode = innererror.ErrNoError
	resultMsg := msg{
		Errno:  noError.Value(),
		Errmsg: noError.ErrorInfo(),
	}

	txid := c.Param("txid")

	log.Log.Info("getTransaction txid:", txid)

	// get unconfirmed from redis
	err, bExist, oneRedisTransaction := notify.GetRedisUnconfirmedTransactionByTxid(txid)
	if nil != err {
		log.Log.Error(err, " GetRedisUnconfirmedTransactionByTxid fail")
		var errorCode innererror.ErrCode = innererror.ErrUnknown
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	if bExist {
		resultMsg.Data = convertRedisTransactionToTransactionInfo(oneRedisTransaction)
		log.Log.Info("getTransaction result:", resultMsg)
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	// transaction
	var oneTransaction []tables.ViewTransactionInfo
	if err := database.Db.Where("transactionhash = ?", txid).Find(&oneTransaction).Error; nil != err {
		log.Log.Error(err, " select * from v_transaction_info fail")
		var errorCode innererror.ErrCode = innererror.ErrSQLError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	if 0 != len(oneTransaction) {
		resultMsg.Data = convertViewTransactionInfo(oneTransaction)
		log.Log.Info("getTransaction result:", resultMsg)
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	var errorCode innererror.ErrCode = innererror.ErrInvalidParaError
	resultMsg.Errno = errorCode.Value()
	resultMsg.Errmsg = errorCode.ErrorInfo()
	c.JSON(http.StatusOK, resultMsg)
}

func getUnspents(c *gin.Context) {
	type unspent struct {
		Address    string `json:"address"`
		Txid       string `json:"txid"`
		Vout_index int64  `json:"vout_index"`
		Value      string `json:"value"`
	}

	type msg struct {
		Errno  int       `json:"errno"`
		Errmsg string    `json:"errmsg"`
		Data   []unspent `json:"data"`
	}

	var noError innererror.ErrCode = innererror.ErrNoError
	resultMsg := msg{
		Errno:  noError.Value(),
		Errmsg: noError.ErrorInfo(),
		Data:   []unspent{},
	}

	// get address list
	var addressBatch []string
	if err := c.BindJSON(&addressBatch); nil != err {
		var errorCode innererror.ErrCode = innererror.ErrDecodeError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	// real address
	addressReal := make([]string, 0)
	for _, oneAddress := range addressBatch {
		if !IsValidAddress(oneAddress) {
			continue
		}

		bExist := false
		for _, oneReal := range addressReal {
			if oneAddress == oneReal {
				bExist = true
				break
			}
		}

		if !bExist {
			addressReal = append(addressReal, oneAddress)
		}
	}

	// parameter invalid
	if 0 == len(addressReal) {
		var errorCode innererror.ErrCode = innererror.ErrInvalidParaError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	log.Log.Info("getUnspents parameter address:", addressReal)

	var address string
	for index, oneAddress := range addressReal {
		if 0 == index {
			address += `'` + oneAddress + `'`
		} else {
			address += `,'` + oneAddress + `'`
		}
	}

	// get unspent
	type unspentTable struct {
		Address    string
		Txid       string
		Vout_index int64
		Value_int  int64
	}
	unspentsTable := []unspentTable{}
	unspentSql := fmt.Sprintf("select `to` as address, hash as txid, n as vout_index, value as value_int from t_output_info where `to` in(%s) and state = 0;", address)
	log.Log.Info("getUnspents select unspent sql:", unspentSql)
	if err := database.Db.Raw(unspentSql).Scan(&unspentsTable).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", unspentSql)
		var errorCode innererror.ErrCode = innererror.ErrSQLError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	err, unfmdUnspents := notify.GetRedisUnconfirmedTransactionUnspentByAddress(addressReal)
	if nil != err {
		log.Log.Error(err, " GetRedisUnconfirmedTransactionByTxid fail")
		var errorCode innererror.ErrCode = innererror.ErrUnknown
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	// all unspent
	for _, oneUnfmd := range unfmdUnspents {
		unspentsTable = append(unspentsTable, unspentTable{oneUnfmd.Address, oneUnfmd.Txid, oneUnfmd.N, oneUnfmd.Value})
	}

	// result
	for _, oneTable := range unspentsTable {
		oneUnspent := unspent{
			Address:    oneTable.Address,
			Txid:       oneTable.Txid,
			Vout_index: oneTable.Vout_index,
			Value:      fmt.Sprintf("%d", oneTable.Value_int),
		}
		resultMsg.Data = append(resultMsg.Data, oneUnspent)
	}

	log.Log.Info("getUnspents result:", resultMsg)

	c.JSON(http.StatusOK, resultMsg)
	return
}

func sendRawTransaction(c *gin.Context) {
	type data struct {
		Txid        string `json:"txid"`
		Receivetime string `json:"receivetime"`
	}

	type msg struct {
		Errno  int    `json:"errno"`
		Errmsg string `json:"errmsg"`
		Data   data   `json:"data"`
	}

	var noError innererror.ErrCode = innererror.ErrNoError
	resultMsg := msg{
		Errno:  noError.Value(),
		Errmsg: noError.ErrorInfo(),
	}

	// get request
	type request struct {
		Tx string
	}
	var oneRequest request
	if err := c.BindJSON(&oneRequest); nil != err {
		var errorCode innererror.ErrCode = innererror.ErrDecodeError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	log.Log.Info("sendRawTransaction parameter hex transaction info:", oneRequest)

	// send raw transaction
	result, err := jsonrpc.Call(1, "sendrawtransaction", []interface{}{oneRequest.Tx})
	if nil != err {
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
		if nil != err {
			var errorCode innererror.ErrCode = innererror.ErrRPCCallError
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusOK, resultMsg)
			return
		} else {
			resultMsg.Errno = resultInfo.Error.Code
			resultMsg.Errmsg = resultInfo.Error.Message
			c.JSON(http.StatusOK, resultMsg)
			return
		}
	} else {
		var txid string
		if err := json.Unmarshal(result, &txid); nil != err {
			var errorCode innererror.ErrCode = innererror.ErrRPCCallError
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusOK, resultMsg)
			return
		} else {
			resultMsg.Data.Txid = txid
			pushErr, receiveTime := notify.HandlePushTransaction(txid)
			if nil != pushErr {
				var errorCode innererror.ErrCode = innererror.ErrRPCCallError
				resultMsg.Errno = errorCode.Value()
				resultMsg.Errmsg = errorCode.ErrorInfo()
				c.JSON(http.StatusOK, resultMsg)
			}
			resultMsg.Data.Receivetime = time.Unix(receiveTime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			log.Log.Info("sendRawTransaction result:", resultMsg)
			c.JSON(http.StatusOK, resultMsg)
			return
		}
	}

}

func getFeeRate(c *gin.Context) {
	type levelFee struct {
		Weight int    `json:"weight"`
		Value  string `json:"value"`
	}

	type msg struct {
		Errno  int        `json:"errno"`
		Errmsg string     `json:"errmsg"`
		Data   []levelFee `json:"data"`
	}

	var noError innererror.ErrCode = innererror.ErrNoError
	resultMsg := msg{
		Errno:  noError.Value(),
		Errmsg: noError.ErrorInfo(),
	}

	// send raw transaction
	result, err := jsonrpc.Call(1, "estimatesmartfee", []interface{}{config.Cfg.Number.BlockNum})
	if nil != err {
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
		if nil != err {
			var errorCode innererror.ErrCode = innererror.ErrRPCCallError
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusBadRequest, resultMsg)
			return
		} else {
			resultMsg.Errno = resultInfo.Error.Code
			resultMsg.Errmsg = resultInfo.Error.Message
			c.JSON(http.StatusBadRequest, resultMsg)
			return
		}
	} else {
		type blockFee struct {
			Feerate float64
			Blocks  int
		}

		var oneBlockFee blockFee
		if err := json.Unmarshal(result, &oneBlockFee); nil != err {
			var errorCode innererror.ErrCode = innererror.ErrRPCCallError
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusBadRequest, resultMsg)
			return
		} else {
			for i := 4; i >= 1; i-- {
				var oneLevelFee levelFee
				if 1 == i {
					oneLevelFee = levelFee{
						Weight: i,
						Value:  fmt.Sprintf("%d", int64(oneBlockFee.Feerate*float64(notify.COIN))),
					}
				}

				// normal
				if 2 == i {
					oneLevelFee = levelFee{
						Weight: i,
						Value:  fmt.Sprintf("%d", int64(oneBlockFee.Feerate*config.Cfg.Number.Normal*float64(notify.COIN))),
					}
				}

				// priority
				if 3 == i {
					oneLevelFee = levelFee{
						Weight: i,
						Value:  fmt.Sprintf("%d", int64(oneBlockFee.Feerate*config.Cfg.Number.Priority*float64(notify.COIN))),
					}
				}

				// quick
				if 4 == i {
					oneLevelFee = levelFee{
						Weight: i,
						Value:  fmt.Sprintf("%d", int64(oneBlockFee.Feerate*config.Cfg.Number.Quick*float64(notify.COIN))),
					}
				}

				resultMsg.Data = append(resultMsg.Data, oneLevelFee)
			}
			c.JSON(http.StatusOK, resultMsg)
			return
		}
	}
}

func getUsdtBalances(c *gin.Context) {
	type info struct {
		Address  string `json:"address"`
		Balance  string `json:"balance"`
		Reserved string `json:"reserved"`
		Frozen   string `json:"frozen"`
	}

	type data struct {
		Usd_price float64 `json:"usd_price"`
		Decimals  int     `json:"decimals"`
		Info      []info  `json:"info"`
	}

	type msg struct {
		Errno  int    `json:"errno"`
		Errmsg string `json:"errmsg"`
		Data   data   `json:"data"`
	}

	var noError innererror.ErrCode = innererror.ErrNoError
	resultMsg := msg{
		Errno:  noError.Value(),
		Errmsg: noError.ErrorInfo(),
		Data:   data{},
	}

	// get address list
	var addressBatch []string
	if err := c.BindJSON(&addressBatch); nil != err {
		log.Log.Error(err, " getUsdtBalances Unmarshal request body fail")
		var errorCode innererror.ErrCode = innererror.ErrDecodeError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusBadRequest, resultMsg)
		return
	}

	// real address
	addressReal := make([]string, 0)
	for _, oneAddress := range addressBatch {
		if !IsValidAddress(oneAddress) {
			continue
		}

		bExist := false
		for _, oneReal := range addressReal {
			if oneAddress == oneReal {
				bExist = true
				break
			}
		}

		if !bExist {
			addressReal = append(addressReal, oneAddress)
		}
	}

	// parameter invalid
	if 0 == len(addressReal) {
		var errorCode innererror.ErrCode = innererror.ErrInvalidParaError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	log.Log.Info("getUsdtBalances address: ", addressReal)

	// price
	ids := []string{"tether"}
	err, resultPrice := request.GetPrice(ids)
	if nil != err {
		log.Log.Error(err, " getUsdtBalances get price fail")
		var errorCode innererror.ErrCode = innererror.ErrRPCCallError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}
	price := resultPrice["tether"]
	resultMsg.Data.Usd_price = price
	resultMsg.Data.Decimals = 8

	// get balance
	type addressBalance struct {
		Propertyid int    `json:"propertyid"`
		Name       string `json:"name"`
		Balance    string `json:"balance"`
		Reserved   string `json:"reserved"`
		Frozen     string `json:"frozen"`
	}

	resultData := make([]info, 0)
	for _, oneAddress := range addressReal {
		result, err := jsonrpc.OmniCall(1, "omni_getallbalancesforaddress", []interface{}{oneAddress})
		if nil != err {
			log.Log.Error(err, " getUsdtBalances jsonrpc OmniCall omni_getallbalancesforaddress fail, address: ", oneAddress)
			continue
		}
		allBalance := make([]addressBalance, 0)
		if err := json.Unmarshal(result, &allBalance); nil != err {
			log.Log.Error(err, " getUsdtBalances Unmarshal result to addressBalance struct fail")
			continue
		}

		for _, oneBalance := range allBalance {
			if 31 == oneBalance.Propertyid {
				oneData := info{
					Address:  oneAddress,
					Balance:  oneBalance.Balance,
					Reserved: oneBalance.Reserved,
					Frozen:   oneBalance.Frozen,
				}
				resultData = append(resultData, oneData)
				break
			}
		}
	}

	// result
	resultMsg.Data.Info = resultData
	log.Log.Info("getUsdtBalances result: ", resultMsg)
	c.JSON(http.StatusOK, resultMsg)
	return
}

func getAddressUsdtTransactions(c *gin.Context) {
	// message type
	type pagination struct {
		Size int `json:"size"`
		Page int `json:"page"`
	}

	type record struct {
		Type   string      `json:"type"`
		Detail interface{} `json:"detail"`
	}

	type data struct {
		Pagination pagination `json:"pagination"`
		List       []record   `json:"list"`
	}

	type msg struct {
		Errno  int    `json:"errno"`
		Errmsg string `json:"errmsg"`
		Data   data   `json:"data"`
	}

	// init message
	var noError innererror.ErrCode = innererror.ErrNoError
	resultMsg := msg{
		Errno:  noError.Value(),
		Errmsg: noError.ErrorInfo(),
	}

	// request info
	type request struct {
		Size      int      `json:"size"`
		Page      int      `json:"page"`
		Addresses []string `json:"addresses"`
		Type      []string `json:"type"`
	}

	var oneRequest request
	if err := c.BindJSON(&oneRequest); nil != err {
		var errorCode innererror.ErrCode = innererror.ErrDecodeError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusBadRequest, resultMsg)
		return
	}

	// real address
	addressReal := make([]string, 0)
	for _, oneAddress := range oneRequest.Addresses {
		if !IsValidAddress(oneAddress) {
			continue
		}

		bExist := false
		for _, oneReal := range addressReal {
			if oneAddress == oneReal {
				bExist = true
				break
			}
		}

		if !bExist {
			addressReal = append(addressReal, oneAddress)
		}
	}

	// parameter invalid
	if 0 == len(addressReal) {
		var errorCode innererror.ErrCode = innererror.ErrInvalidParaError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	log.Log.Info("getAddressUsdtTransactions address: ", addressReal)

	var address string
	for index, oneAddress := range addressReal {
		if 0 == index {
			address += `'` + oneAddress + `'`
		} else {
			address += `,'` + oneAddress + `'`
		}
	}

	var strType string
	for index, oneType := range oneRequest.Type {
		if 0 == index {
			strType += `'` + oneType + `'`
		} else {
			strType += `,'` + oneType + `'`
		}
	}

	// unconfirmed transaction count
	err, unfmdTransaction := omni.GetRedisUnconfirmedOmniTransactionByAddressAndType(addressReal, oneRequest.Type)
	if nil != err {
		log.Log.Error(err, " GetRedisUnconfirmedTransactionByAddress fail")
		var errorCode innererror.ErrCode = innererror.ErrUnknown
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	// get transaction
	pageInfo := pagination{
		Size: oneRequest.Size,
		Page: oneRequest.Page,
	}
	resultMsg.Data.Pagination = pageInfo

	if oneRequest.Page <= 0 {
		log.Log.Error(fmt.Sprintf("page number %d, size %d out of range", oneRequest.Page, oneRequest.Size))
		var errorCode innererror.ErrCode = innererror.ErrOutOfRangeError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	// all unconfirmed and confirmed transactions
	txUnfmdRedisTrx := []omni.OmniTransaction{}
	begin := (oneRequest.Page - 1) * oneRequest.Size
	end := oneRequest.Page * oneRequest.Size
	limitBegin := 0
	limitEnd := 0
	if end <= len(unfmdTransaction) {
		for i := begin; i < end; i++ {
			txUnfmdRedisTrx = append(txUnfmdRedisTrx, unfmdTransaction[i])
		}
	} else if begin < len(unfmdTransaction) {
		for i := begin; i < len(unfmdTransaction); i++ {
			txUnfmdRedisTrx = append(txUnfmdRedisTrx, unfmdTransaction[i])
		}
		limitBegin = 0
		limitEnd = oneRequest.Size - len(txUnfmdRedisTrx)
	} else {
		limitBegin = begin - len(unfmdTransaction)
		limitEnd = oneRequest.Size
	}

	list := []record{}
	if 0 != len(txUnfmdRedisTrx) {
		for _, oneRedis := range txUnfmdRedisTrx {
			if err, typeInfo, receiveTime, detail := omni.ConvertToRecord(omni.ConvertTransactionToTableOmniTransactionInfo(&oneRedis), true, &oneRedis); nil != err {
				log.Log.Error(err, " ConvertToRecord fail, omni transaction hash:", oneRedis.Txid)
			} else {
				msgErr, message := omni.ConvertToMessage(detail, receiveTime)
				if nil != msgErr {
					var errorCode innererror.ErrCode = innererror.ErrUnknown
					resultMsg.Errno = errorCode.Value()
					resultMsg.Errmsg = errorCode.ErrorInfo()
					c.JSON(http.StatusOK, resultMsg)
					return
				}
				one := record{
					Type:   typeInfo,
					Detail: message,
				}
				list = append(list, one)
			}
		}
	}

	if 0 != limitEnd {
		allOmniTransaction := []tables.TableOmniTransactionInfo{}
		selectSql := fmt.Sprintf("select * from t_omni_transaction_info where (`sendingaddress` in (%s) or `referenceaddress` in (%s)) and `type` in (%s) and propertyid = 31 order by blocktime desc limit %d, %d;", address, address, strType, limitBegin, limitEnd)
		log.Log.Info("get omni transaction by address and type select transaction sql:", selectSql)
		if err := database.Db.Raw(selectSql).Scan(&allOmniTransaction).Error; nil != err {
			log.Log.Error(err, " exec sql fail: ", selectSql)
			var errorCode innererror.ErrCode = innererror.ErrSQLError
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusOK, resultMsg)
			return
		}

		for _, oneTrx := range allOmniTransaction {
			if err, typeInfo, receiveTime, detail := omni.ConvertToRecord(&oneTrx, false, nil); nil != err {
				log.Log.Error(err, " ConvertToRecord fail, omni transaction hash:", oneTrx.Txid)
			} else {
				msgErr, message := omni.ConvertToMessage(detail, receiveTime)
				if nil != msgErr {
					var errorCode innererror.ErrCode = innererror.ErrUnknown
					resultMsg.Errno = errorCode.Value()
					resultMsg.Errmsg = errorCode.ErrorInfo()
					c.JSON(http.StatusOK, resultMsg)
					return
				}
				one := record{
					Type:   typeInfo,
					Detail: message,
				}
				list = append(list, one)
			}
		}
	}

	resultMsg.Data.List = list
	log.Log.Info("getAddressUsdtTransactions result: ", resultMsg)
	c.JSON(http.StatusOK, resultMsg)
	return
}

func getUsdtTransaction(c *gin.Context) {
	type usdtTransaction struct {
		Type   string      `json:"type"`
		Detail interface{} `json:"detail"`
	}

	// message type
	type msg struct {
		Errno  int             `json:"errno"`
		Errmsg string          `json:"errmsg"`
		Data   usdtTransaction `json:"data"`
	}

	// init message
	var noError innererror.ErrCode = innererror.ErrNoError
	resultMsg := msg{
		Errno:  noError.Value(),
		Errmsg: noError.ErrorInfo(),
	}

	// unconfirmed transaction
	txid := c.Param("txid")

	log.Log.Info("getUsdtTransaction txid:", txid)

	err, bExist, oneRedisOmniTransaction := omni.GetRedisUnconfirmedOmniTransactionByTxid(txid)
	if nil != err {
		log.Log.Error(err, " GetRedisUnconfirmedOmniTransactionByTxid fail")
		var errorCode innererror.ErrCode = innererror.ErrUnknown
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	if bExist {
		if err, typeInfo, receiveTime, detail := omni.ConvertToRecord(omni.ConvertTransactionToTableOmniTransactionInfo(oneRedisOmniTransaction), true, oneRedisOmniTransaction); nil != err {
			log.Log.Error(err, " ConvertToRecord fail, omni unconfirmed transaction hash:", oneRedisOmniTransaction.Txid)
			var errorCode innererror.ErrCode = innererror.ErrUnknown
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusOK, resultMsg)
			return
		} else {
			msgErr, message := omni.ConvertToMessage(detail, receiveTime)
			if nil != msgErr {
				var errorCode innererror.ErrCode = innererror.ErrUnknown
				resultMsg.Errno = errorCode.Value()
				resultMsg.Errmsg = errorCode.ErrorInfo()
				c.JSON(http.StatusOK, resultMsg)
				return
			}
			one := usdtTransaction{
				Type:   typeInfo,
				Detail: message,
			}
			resultMsg.Data = one
			log.Log.Info("getUsdtTransaction result:", resultMsg)
			c.JSON(http.StatusOK, resultMsg)
			return
		}
	}

	pageTransaction := []tables.TableOmniTransactionInfo{}
	txSelectSql := fmt.Sprintf("select * from t_omni_transaction_info where txid='%s';", txid)
	if err := database.Db.Raw(txSelectSql).Scan(&pageTransaction).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", txSelectSql)
		var errorCode innererror.ErrCode = innererror.ErrSQLError
		resultMsg.Errno = errorCode.Value()
		resultMsg.Errmsg = errorCode.ErrorInfo()
		c.JSON(http.StatusOK, resultMsg)
		return
	}

	if 0 != len(pageTransaction) {
		if err, typeInfo, receiveTime, detail := omni.ConvertToRecord(&pageTransaction[0], false, nil); nil != err {
			log.Log.Error(err, " ConvertToRecord fail, omni transaction hash:", pageTransaction[0].Txid)
			var errorCode innererror.ErrCode = innererror.ErrUnknown
			resultMsg.Errno = errorCode.Value()
			resultMsg.Errmsg = errorCode.ErrorInfo()
			c.JSON(http.StatusOK, resultMsg)
			return
		} else {
			msgErr, message := omni.ConvertToMessage(detail, receiveTime)
			if nil != msgErr {
				var errorCode innererror.ErrCode = innererror.ErrUnknown
				resultMsg.Errno = errorCode.Value()
				resultMsg.Errmsg = errorCode.ErrorInfo()
				c.JSON(http.StatusOK, resultMsg)
				return
			}
			one := usdtTransaction{
				Type:   typeInfo,
				Detail: message,
			}
			resultMsg.Data = one
			log.Log.Info("getUsdtTransaction result:", resultMsg)
			c.JSON(http.StatusOK, resultMsg)
			return
		}
	}

	var errorCode innererror.ErrCode = innererror.ErrInvalidParaError
	resultMsg.Errno = errorCode.Value()
	resultMsg.Errmsg = errorCode.ErrorInfo()
	c.JSON(http.StatusOK, resultMsg)
}
