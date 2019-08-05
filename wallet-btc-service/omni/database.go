package omni

import (
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/database/tables"
)

func GetOmniDbMaxHeight() (errInfo error, blockheight int32) {
	type maxHeight struct {
		End int32
	}

	var result maxHeight
	selectSql := "select max(block) as end from t_omni_transaction_info;"
	if err := database.Db.Raw(selectSql).Scan(&result).Error; nil != err {
		errInfo = err
		return
	}

	return nil, result.End
}

func IsOmniTransactionExist(hash string) (errInfo error, bExist bool) {
	var count int
	err := database.Db.Model(&tables.TableOmniTransactionInfo{}).Where(&tables.TableOmniTransactionInfo{Txid: hash}).Count(&count).Error
	if err != nil {
		errInfo = err
		return
	}

	if count > 0 {
		bExist = true
	}

	return
}
