package omni

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/database/tables"
)

const (
	MSC_TYPE_SIMPLE_SEND               = 0
	MSC_TYPE_RESTRICTED_SEND           = 2
	MSC_TYPE_SEND_TO_OWNERS            = 3
	MSC_TYPE_SEND_ALL                  = 4
	MSC_TYPE_SAVINGS_MARK              = 10
	MSC_TYPE_SAVINGS_COMPROMISED       = 11
	MSC_TYPE_RATELIMITED_MARK          = 12
	MSC_TYPE_AUTOMATIC_DISPENSARY      = 15
	MSC_TYPE_TRADE_OFFER               = 20
	MSC_TYPE_ACCEPT_OFFER_BTC          = 22
	MSC_TYPE_METADEX_TRADE             = 25
	MSC_TYPE_METADEX_CANCEL_PRICE      = 26
	MSC_TYPE_METADEX_CANCEL_PAIR       = 27
	MSC_TYPE_METADEX_CANCEL_ECOSYSTEM  = 28
	MSC_TYPE_NOTIFICATION              = 31
	MSC_TYPE_OFFER_ACCEPT_A_BET        = 40
	MSC_TYPE_CREATE_PROPERTY_FIXED     = 50
	MSC_TYPE_CREATE_PROPERTY_VARIABLE  = 51
	MSC_TYPE_PROMOTE_PROPERTY          = 52
	MSC_TYPE_CLOSE_CROWDSALE           = 53
	MSC_TYPE_CREATE_PROPERTY_MANUAL    = 54
	MSC_TYPE_GRANT_PROPERTY_TOKENS     = 55
	MSC_TYPE_REVOKE_PROPERTY_TOKENS    = 56
	MSC_TYPE_CHANGE_ISSUER_ADDRESS     = 70
	MSC_TYPE_ENABLE_FREEZING           = 71
	MSC_TYPE_DISABLE_FREEZING          = 72
	MSC_TYPE_FREEZE_PROPERTY_TOKENS    = 185
	MSC_TYPE_UNFREEZE_PROPERTY_TOKENS  = 186
	OMNICORE_MESSAGE_TYPE_DEACTIVATION = 65533
	OMNICORE_MESSAGE_TYPE_ACTIVATION   = 65534
	OMNICORE_MESSAGE_TYPE_ALERT        = 65535
)

func GetOmniTransactionType(txType int) string {
	switch txType {
	case MSC_TYPE_SIMPLE_SEND:
		return "Simple Send"
	case MSC_TYPE_RESTRICTED_SEND:
		return "Restricted Send"
	case MSC_TYPE_SEND_TO_OWNERS:
		return "Send To Owners"
	case MSC_TYPE_SEND_ALL:
		return "Send All"
	case MSC_TYPE_SAVINGS_MARK:
		return "Savings"
	case MSC_TYPE_SAVINGS_COMPROMISED:
		return "Savings COMPROMISED"
	case MSC_TYPE_RATELIMITED_MARK:
		return "Rate-Limiting"
	case MSC_TYPE_AUTOMATIC_DISPENSARY:
		return "Automatic Dispensary"
	case MSC_TYPE_TRADE_OFFER:
		return "DEx Sell Offer"
	case MSC_TYPE_METADEX_TRADE:
		return "MetaDEx trade"
	case MSC_TYPE_METADEX_CANCEL_PRICE:
		return "MetaDEx cancel-price"
	case MSC_TYPE_METADEX_CANCEL_PAIR:
		return "MetaDEx cancel-pair"
	case MSC_TYPE_METADEX_CANCEL_ECOSYSTEM:
		return "MetaDEx cancel-ecosystem"
	case MSC_TYPE_ACCEPT_OFFER_BTC:
		return "DEx Accept Offer"
	case MSC_TYPE_CREATE_PROPERTY_FIXED:
		return "Create Property - Fixed"
	case MSC_TYPE_CREATE_PROPERTY_VARIABLE:
		return "Create Property - Variable"
	case MSC_TYPE_PROMOTE_PROPERTY:
		return "Promote Property"
	case MSC_TYPE_CLOSE_CROWDSALE:
		return "Close Crowdsale"
	case MSC_TYPE_CREATE_PROPERTY_MANUAL:
		return "Create Property - Manual"
	case MSC_TYPE_GRANT_PROPERTY_TOKENS:
		return "Grant Property Tokens"
	case MSC_TYPE_REVOKE_PROPERTY_TOKENS:
		return "Revoke Property Tokens"
	case MSC_TYPE_CHANGE_ISSUER_ADDRESS:
		return "Change Issuer Address"
	case MSC_TYPE_ENABLE_FREEZING:
		return "Enable Freezing"
	case MSC_TYPE_DISABLE_FREEZING:
		return "Disable Freezing"
	case MSC_TYPE_FREEZE_PROPERTY_TOKENS:
		return "Freeze Property Tokens"
	case MSC_TYPE_UNFREEZE_PROPERTY_TOKENS:
		return "Unfreeze Property Tokens"
	case MSC_TYPE_NOTIFICATION:
		return "Notification"
	case OMNICORE_MESSAGE_TYPE_ALERT:
		return "ALERT"
	case OMNICORE_MESSAGE_TYPE_DEACTIVATION:
		return "Feature Deactivation"
	case OMNICORE_MESSAGE_TYPE_ACTIVATION:
		return "Feature Activation"

	default:
		return "* unknown type *"
	}
}

type DExPurchase struct {
	Txid           string         `json:"txid"`
	Type           string         `json:"type"`
	Sendingaddress string         `json:"sendingaddress"`
	Purchases      []OmniPurchase `json:"purchases"`
	Blockhash      string         `json:"blockhash"`
	Blocktime      int64          `json:"blocktime"`
	Block          int32          `json:"block"`
}

type OmniCommon struct {
	Txid             string `json:"txid"`
	Fee              string `json:"fee"`
	Sendingaddress   string `json:"sendingaddress"`
	Referenceaddress string `json:"referenceaddress"`
	Ismine           bool   `json:"ismine"`
	Version          uint64 `json:"version"`
	Type_int         uint64 `json:"type_int"`
	Type             string `json:"type"`
	Valid            bool   `json:"valid"`
	Invalidreason    string `json:"invalidreason"`
	Blockhash        string `json:"blockhash"`
	Blocktime        int64  `json:"blocktime"`
	Positioninblock  int32  `json:"positioninblock"`
	Block            int32  `json:"block"`
}

type SimpleSend struct {
	OmniCommon
	Propertyid uint64 `json:"propertyid"`
	Divisible  bool   `json:"divisible"`
	Amount     string `json:"amount"`
}

type CrowdsalePurchase struct {
	SimpleSend
	Purchasedpropertyid        int64  `json:"purchasedpropertyid"`
	Purchasedpropertyname      string `json:"purchasedpropertyname"`
	Purchasedpropertydivisible bool   `json:"purchasedpropertydivisible"`
	Purchasedtokens            string `json:"purchasedtokens"`
	Issuertokens               string `json:"issuertokens"`
}

type SendToOwners struct {
	SimpleSend
	Totalstofee string          `json:"totalstofee"`
	Recipients  []OmniRecipient `json:"recipients"`
}

type SendAll struct {
	OmniCommon
	Ecosystem string        `json:"ecosystem"`
	Subsends  []OmniSubsend `json:"subsends"`
}

type DExSellOffer struct {
	SimpleSend
	Bitcoindesired string `json:"bitcoindesired"`
	Timelimit      uint8  `json:"timelimit"`
	Feerequired    string `json:"feerequired"`
	Action         string `json:"action"`
}

type MetaDExTrade struct {
	OmniCommon
	Propertyidforsale            uint64      `json:"propertyidforsale"`
	Propertyidforsaleisdivisible bool        `json:"propertyidforsaleisdivisible"`
	Amountforsale                string      `json:"amountforsale"`
	Propertyiddesired            uint64      `json:"propertyiddesired"`
	Propertyiddesiredisdivisible bool        `json:"propertyiddesiredisdivisible"`
	Amountdesired                string      `json:"amountdesired"`
	Unitprice                    string      `json:"unitprice"`
	Amountremaining              string      `json:"amountremaining"`
	Amounttofill                 string      `json:"amounttofill"`
	Status                       string      `json:"status"`
	Canceltxid                   string      `json:"canceltxid"`
	Matches                      []OmniTrade `json:"matches"`
}

type MetaDExCancelPrice struct {
	OmniCommon
	Propertyidforsale            uint64       `json:"propertyidforsale"`
	Propertyidforsaleisdivisible bool         `json:"propertyidforsaleisdivisible"`
	Amountforsale                string       `json:"amountforsale"`
	Propertyiddesired            uint64       `json:"propertyiddesired"`
	Propertyiddesiredisdivisible bool         `json:"propertyiddesiredisdivisible"`
	Amountdesired                string       `json:"amountdesired"`
	Unitprice                    string       `json:"unitprice"`
	Cancelledtransactions        []OmniCancel `json:"cancelledtransactions"`
}

type MetaDExCancelPair struct {
	OmniCommon
	Propertyidforsale     uint64       `json:"propertyidforsale"`
	Propertyiddesired     uint64       `json:"propertyiddesired"`
	Cancelledtransactions []OmniCancel `json:"cancelledtransactions"`
}

type MetaDExCancelEcosystem struct {
	OmniCommon
	Ecosystem             string       `json:"ecosystem"`
	Cancelledtransactions []OmniCancel `json:"cancelledtransactions"`
}

type DExAcceptOffer struct {
	SimpleSend
}

type CreatePropertyFixed struct {
	OmniCommon
	Propertyid   uint64 `json:"propertyid"`
	Divisible    bool   `json:"divisible"`
	Ecosystem    string `json:"ecosystem"`
	Propertytype string `json:"propertytype"`
	Category     string `json:"category"`
	Subcategory  string `json:"subcategory"`
	Propertyname string `json:"propertyname"`
	Data         string `json:"data"`
	Url          string `json:"url"`
	Amount       string `json:"amount"`
}

type CreatePropertyVariable struct {
	CreatePropertyFixed
	Propertyiddesired uint64 `json:"propertyiddesired"`
	Tokensperunit     string `json:"tokensperunit"`
	Deadline          int64  `json:"deadline"`
	Earlybonus        uint8  `json:"earlybonus"`
	Percenttoissuer   uint8  `json:"percenttoissuer"`
}

type CreatePropertyManual struct {
	CreatePropertyFixed
}

type CloseCrowdsale struct {
	OmniCommon
	Propertyid uint64 `json:"propertyid"`
	Divisible  bool   `json:"divisible"`
}

type GrantPropertyTokens struct {
	SimpleSend
}

type RevokePropertyTokens struct {
	SimpleSend
}

type ChangeIssuerAddress struct {
	CloseCrowdsale
}

type EnableFreezing struct {
	OmniCommon
	Propertyid uint64 `json:"propertyid"`
}

type DisableFreezing struct {
	EnableFreezing
}

type FreezePropertyTokens struct {
	EnableFreezing
}

type UnfreezePropertyTokens struct {
	EnableFreezing
}

type FeatureActivation struct {
	OmniCommon
	Featureid       uint64 `json:"featureid"`
	Activationblock uint64 `json:"activationblock"`
	Minimumversion  uint64 `json:"minimumversion"`
}

func ConvertToDExPurchase(trx *tables.TableOmniTransactionInfo, allPurchase []tables.TableOmniPurchaseInfo) *DExPurchase {
	one := DExPurchase{
		Txid:           trx.Txid,
		Type:           trx.Type,
		Sendingaddress: trx.Sendingaddress,
		Purchases:      []OmniPurchase{},
		Blockhash:      trx.Blockhash,
		Blocktime:      trx.Blocktime,
		Block:          trx.Block,
	}

	for _, onePurchaseTable := range allPurchase {
		ismine := false
		if 1 == onePurchaseTable.Ismine {
			ismine = true
		}
		valid := false
		if 1 == onePurchaseTable.Valid {
			valid = true
		}
		onePurchase := OmniPurchase{
			Vout:             onePurchaseTable.Vout,
			Amountpaid:       onePurchaseTable.Amountpaid,
			Ismine:           ismine,
			Referenceaddress: onePurchaseTable.Referenceaddress,
			Propertyid:       onePurchaseTable.Propertyid,
			Amountbought:     onePurchaseTable.Amountbought,
			Valid:            valid,
		}
		one.Purchases = append(one.Purchases, onePurchase)
	}

	return &one
}

func ConvertToOmniCommon(trx *tables.TableOmniTransactionInfo) *OmniCommon {
	ismine := false
	if 1 == trx.Ismine {
		ismine = true
	}

	valid := false
	if 1 == trx.Valid {
		valid = true
	}

	one := OmniCommon{
		Txid:             trx.Txid,
		Fee:              trx.Fee,
		Sendingaddress:   trx.Sendingaddress,
		Referenceaddress: trx.Referenceaddress,
		Ismine:           ismine,
		Version:          trx.Version,
		Type_int:         trx.Type_int,
		Type:             trx.Type,
		Valid:            valid,
		Invalidreason:    trx.Invalidreason,
		Blockhash:        trx.Blockhash,
		Blocktime:        trx.Blocktime,
		Positioninblock:  trx.Positioninblock,
		Block:            trx.Block,
	}
	return &one
}

func ConvertToSimpleSend(trx *tables.TableOmniTransactionInfo) *SimpleSend {
	divisible := false
	if 1 == trx.Divisible {
		divisible = true
	}
	one := SimpleSend{
		OmniCommon: *ConvertToOmniCommon(trx),
		Propertyid: trx.Propertyid,
		Divisible:  divisible,
		Amount:     trx.Amount,
	}

	return &one
}

func ConvertToCrowdsalePurchase(trx *tables.TableOmniTransactionInfo) *CrowdsalePurchase {
	purchasedpropertydivisible := false
	if 1 == trx.Purchasedpropertydivisible {
		purchasedpropertydivisible = true
	}
	one := CrowdsalePurchase{
		SimpleSend:                 *ConvertToSimpleSend(trx),
		Purchasedpropertyid:        trx.Purchasedpropertyid,
		Purchasedpropertyname:      trx.Purchasedpropertyname,
		Purchasedpropertydivisible: purchasedpropertydivisible,
		Purchasedtokens:            trx.Purchasedtokens,
		Issuertokens:               trx.Issuertokens,
	}
	return &one
}

func ConvertToSendToOwners(trx *tables.TableOmniTransactionInfo, allRecipient []tables.TableOmniRecipientInfo) *SendToOwners {
	one := SendToOwners{
		SimpleSend:  *ConvertToSimpleSend(trx),
		Totalstofee: trx.Totalstofee,
		Recipients:  []OmniRecipient{},
	}

	for _, oneRecipientTable := range allRecipient {
		oneRecipient := OmniRecipient{
			Address: oneRecipientTable.Address,
			Amount:  oneRecipientTable.Amount,
		}
		one.Recipients = append(one.Recipients, oneRecipient)
	}
	return &one
}

func ConvertToSendAll(trx *tables.TableOmniTransactionInfo, allSubsend []tables.TableOmniSubsendInfo) *SendAll {
	one := SendAll{
		OmniCommon: *ConvertToOmniCommon(trx),
		Ecosystem:  trx.Ecosystem,
		Subsends:   []OmniSubsend{},
	}

	for _, oneSubsendTable := range allSubsend {
		divisible := false
		if 1 == oneSubsendTable.Divisible {
			divisible = true
		}
		oneSubsend := OmniSubsend{
			Propertyid: trx.Propertyid,
			Divisible:  divisible,
			Amount:     trx.Amount,
		}
		one.Subsends = append(one.Subsends, oneSubsend)
	}
	return &one
}

func ConvertToDExSellOffer(trx *tables.TableOmniTransactionInfo) *DExSellOffer {
	one := DExSellOffer{
		SimpleSend:     *ConvertToSimpleSend(trx),
		Bitcoindesired: trx.Bitcoindesired,
		Timelimit:      trx.Timelimit,
		Feerequired:    trx.Feerequired,
		Action:         trx.Action,
	}

	return &one
}

func ConvertToMetaDExTrade(trx *tables.TableOmniTransactionInfo, allTrade []tables.TableOmniTradeInfo) *MetaDExTrade {
	propertyidforsaleisdivisible := false
	if 1 == trx.Propertyidforsaleisdivisible {
		propertyidforsaleisdivisible = true
	}

	propertyiddesiredisdivisible := false
	if 1 == trx.Propertyiddesiredisdivisible {
		propertyiddesiredisdivisible = true
	}

	one := MetaDExTrade{
		OmniCommon:                   *ConvertToOmniCommon(trx),
		Propertyidforsale:            trx.Propertyidforsale,
		Propertyidforsaleisdivisible: propertyidforsaleisdivisible,
		Amountforsale:                trx.Amountforsale,
		Propertyiddesired:            trx.Propertyiddesired,
		Propertyiddesiredisdivisible: propertyiddesiredisdivisible,
		Amountdesired:                trx.Amountdesired,
		Unitprice:                    trx.Unitprice,
		Amountremaining:              trx.Amountremaining,
		Amounttofill:                 trx.Amounttofill,
		Status:                       trx.Status,
		Canceltxid:                   trx.Canceltxid,
		Matches:                      []OmniTrade{},
	}

	for _, oneTradeTable := range allTrade {
		oneTrade := OmniTrade{
			Txid:           oneTradeTable.Txid,
			Block:          oneTradeTable.Block,
			Address:        oneTradeTable.Address,
			Amountsold:     oneTradeTable.Amountsold,
			Amountreceived: oneTradeTable.Amountreceived,
			Tradingfee:     oneTradeTable.Tradingfee,
		}
		one.Matches = append(one.Matches, oneTrade)
	}

	return &one
}

func ConvertToMetaDExCancelPrice(trx *tables.TableOmniTransactionInfo, allCancel []tables.TableOmniCancelInfo) *MetaDExCancelPrice {
	propertyidforsaleisdivisible := false
	if 1 == trx.Propertyidforsaleisdivisible {
		propertyidforsaleisdivisible = true
	}

	propertyiddesiredisdivisible := false
	if 1 == trx.Propertyiddesiredisdivisible {
		propertyiddesiredisdivisible = true
	}

	one := MetaDExCancelPrice{
		OmniCommon:                   *ConvertToOmniCommon(trx),
		Propertyidforsale:            trx.Propertyidforsale,
		Propertyidforsaleisdivisible: propertyidforsaleisdivisible,
		Amountforsale:                trx.Amountforsale,
		Propertyiddesired:            trx.Propertyiddesired,
		Propertyiddesiredisdivisible: propertyiddesiredisdivisible,
		Amountdesired:                trx.Amountdesired,
		Unitprice:                    trx.Unitprice,
		Cancelledtransactions:        []OmniCancel{},
	}

	for _, oneCancelTable := range allCancel {
		oneCancel := OmniCancel{
			Txid:             oneCancelTable.Txid,
			Propertyid:       oneCancelTable.Propertyid,
			Amountunreserved: oneCancelTable.Amountunreserved,
		}
		one.Cancelledtransactions = append(one.Cancelledtransactions, oneCancel)
	}

	return &one
}

func ConvertToMetaDExCancelPair(trx *tables.TableOmniTransactionInfo, allCancel []tables.TableOmniCancelInfo) *MetaDExCancelPair {
	one := MetaDExCancelPair{
		OmniCommon:            *ConvertToOmniCommon(trx),
		Propertyidforsale:     trx.Propertyidforsale,
		Propertyiddesired:     trx.Propertyiddesired,
		Cancelledtransactions: []OmniCancel{},
	}

	for _, oneCancelTable := range allCancel {
		oneCancel := OmniCancel{
			Txid:             oneCancelTable.Txid,
			Propertyid:       oneCancelTable.Propertyid,
			Amountunreserved: oneCancelTable.Amountunreserved,
		}
		one.Cancelledtransactions = append(one.Cancelledtransactions, oneCancel)
	}

	return &one
}

func ConvertToMetaDExCancelEcosystem(trx *tables.TableOmniTransactionInfo, allCancel []tables.TableOmniCancelInfo) *MetaDExCancelEcosystem {
	one := MetaDExCancelEcosystem{
		OmniCommon:            *ConvertToOmniCommon(trx),
		Ecosystem:             trx.Ecosystem,
		Cancelledtransactions: []OmniCancel{},
	}

	for _, oneCancelTable := range allCancel {
		oneCancel := OmniCancel{
			Txid:             oneCancelTable.Txid,
			Propertyid:       oneCancelTable.Propertyid,
			Amountunreserved: oneCancelTable.Amountunreserved,
		}
		one.Cancelledtransactions = append(one.Cancelledtransactions, oneCancel)
	}

	return &one
}

func ConvertToDExAcceptOffer(trx *tables.TableOmniTransactionInfo) *DExAcceptOffer {
	one := DExAcceptOffer{
		SimpleSend: *ConvertToSimpleSend(trx),
	}

	return &one
}

func ConvertToCreatePropertyFixed(trx *tables.TableOmniTransactionInfo) *CreatePropertyFixed {
	divisible := false
	if 1 == trx.Divisible {
		divisible = true
	}
	one := CreatePropertyFixed{
		OmniCommon:   *ConvertToOmniCommon(trx),
		Propertyid:   trx.Propertyid,
		Divisible:    divisible,
		Ecosystem:    trx.Ecosystem,
		Propertytype: trx.Propertytype,
		Category:     trx.Category,
		Subcategory:  trx.Subcategory,
		Propertyname: trx.Propertyname,
		Data:         trx.Data,
		Url:          trx.Url,
		Amount:       trx.Amount,
	}

	return &one
}

func ConvertToCreatePropertyVariable(trx *tables.TableOmniTransactionInfo) *CreatePropertyVariable {
	one := CreatePropertyVariable{
		CreatePropertyFixed: *ConvertToCreatePropertyFixed(trx),
		Propertyiddesired:   trx.Propertyiddesired,
		Tokensperunit:       trx.Tokensperunit,
		Deadline:            trx.Deadline,
		Earlybonus:          trx.Earlybonus,
		Percenttoissuer:     trx.Percenttoissuer,
	}

	return &one
}

func ConvertToCreatePropertyManual(trx *tables.TableOmniTransactionInfo) *CreatePropertyManual {
	one := CreatePropertyManual{
		CreatePropertyFixed: *ConvertToCreatePropertyFixed(trx),
	}

	return &one
}

func ConvertToCloseCrowdsale(trx *tables.TableOmniTransactionInfo) *CloseCrowdsale {
	divisible := false
	if 1 == trx.Divisible {
		divisible = true
	}

	one := CloseCrowdsale{
		OmniCommon: *ConvertToOmniCommon(trx),
		Propertyid: trx.Propertyid,
		Divisible:  divisible,
	}

	return &one
}

func ConvertToGrantPropertyTokens(trx *tables.TableOmniTransactionInfo) *GrantPropertyTokens {
	one := GrantPropertyTokens{
		SimpleSend: *ConvertToSimpleSend(trx),
	}

	return &one
}

func ConvertToRevokePropertyTokens(trx *tables.TableOmniTransactionInfo) *RevokePropertyTokens {
	one := RevokePropertyTokens{
		SimpleSend: *ConvertToSimpleSend(trx),
	}

	return &one
}

func ConvertToChangeIssuerAddress(trx *tables.TableOmniTransactionInfo) *ChangeIssuerAddress {
	one := ChangeIssuerAddress{
		CloseCrowdsale: *ConvertToCloseCrowdsale(trx),
	}

	return &one
}

func ConvertToEnableFreezing(trx *tables.TableOmniTransactionInfo) *EnableFreezing {
	one := EnableFreezing{
		OmniCommon: *ConvertToOmniCommon(trx),
		Propertyid: trx.Propertyid,
	}

	return &one
}

func ConvertToDisableFreezing(trx *tables.TableOmniTransactionInfo) *DisableFreezing {
	one := DisableFreezing{
		EnableFreezing: *ConvertToEnableFreezing(trx),
	}

	return &one
}

func ConvertToFreezePropertyTokens(trx *tables.TableOmniTransactionInfo) *FreezePropertyTokens {
	one := FreezePropertyTokens{
		EnableFreezing: *ConvertToEnableFreezing(trx),
	}

	return &one
}

func ConvertToUnfreezePropertyTokens(trx *tables.TableOmniTransactionInfo) *UnfreezePropertyTokens {
	one := UnfreezePropertyTokens{
		EnableFreezing: *ConvertToEnableFreezing(trx),
	}

	return &one
}

func ConvertToFeatureActivation(trx *tables.TableOmniTransactionInfo) *FeatureActivation {
	one := FeatureActivation{
		OmniCommon:      *ConvertToOmniCommon(trx),
		Featureid:       trx.Featureid,
		Activationblock: trx.Activationblock,
		Minimumversion:  trx.Minimumversion,
	}

	return &one
}

func GetPurchase(txid string, bUnfmd bool, oneRedisUnfmdTransaction *OmniTransaction) (errInfo error, allPurchase []tables.TableOmniPurchaseInfo) {
	if bUnfmd {
		for _, onePurchase := range oneRedisUnfmdTransaction.Purchases {
			var ismine int8 = 0
			if onePurchase.Ismine {
				ismine = 1
			}
			var isValid int8 = 0
			if onePurchase.Valid {
				isValid = 1
			}

			oneTable := tables.TableOmniPurchaseInfo{
				Txid:             oneRedisUnfmdTransaction.Txid,
				Vout:             onePurchase.Vout,
				Amountpaid:       onePurchase.Amountpaid,
				Ismine:           ismine,
				Referenceaddress: onePurchase.Referenceaddress,
				Propertyid:       onePurchase.Propertyid,
				Amountbought:     onePurchase.Amountbought,
				Valid:            isValid,
			}
			allPurchase = append(allPurchase, oneTable)
		}
		return
	}

	txSelectSql := fmt.Sprintf("select * from t_omni_purchase_info where txid = '%s';", txid)
	if err := database.Db.Raw(txSelectSql).Scan(&allPurchase).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", txSelectSql)
		errInfo = err
		return
	}
	return
}

func GetRecipientInfo(txid string, bUnfmd bool, oneRedisUnfmdTransaction *OmniTransaction) (errInfo error, allRecipientInfo []tables.TableOmniRecipientInfo) {
	if bUnfmd {
		for _, oneRecipientInfo := range oneRedisUnfmdTransaction.Recipients {
			oneTable := tables.TableOmniRecipientInfo{
				Txid:    oneRedisUnfmdTransaction.Txid,
				Address: oneRecipientInfo.Address,
				Amount:  oneRecipientInfo.Amount,
			}
			allRecipientInfo = append(allRecipientInfo, oneTable)
		}
		return
	}
	txSelectSql := fmt.Sprintf("select * from t_omni_recipient_info where txid = '%s';", txid)
	if err := database.Db.Raw(txSelectSql).Scan(&allRecipientInfo).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", txSelectSql)
		errInfo = err
		return
	}
	return
}

func GetSubsend(txid string, bUnfmd bool, oneRedisUnfmdTransaction *OmniTransaction) (errInfo error, allSubsend []tables.TableOmniSubsendInfo) {
	if bUnfmd {
		for _, oneSubsend := range oneRedisUnfmdTransaction.Subsends {
			var divisible int8 = 0
			if oneSubsend.Divisible {
				divisible = 1
			}
			oneTable := tables.TableOmniSubsendInfo{
				Txid:       oneRedisUnfmdTransaction.Txid,
				Propertyid: int64(oneSubsend.Propertyid),
				Divisible:  divisible,
				Amount:     oneSubsend.Amount,
			}
			allSubsend = append(allSubsend, oneTable)
		}
		return
	}
	txSelectSql := fmt.Sprintf("select * from t_omni_subsend_info where txid = '%s';", txid)
	if err := database.Db.Raw(txSelectSql).Scan(&allSubsend).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", txSelectSql)
		errInfo = err
		return
	}
	return
}

func GetTrade(txid string, bUnfmd bool, oneRedisUnfmdTransaction *OmniTransaction) (errInfo error, allTrade []tables.TableOmniTradeInfo) {
	if bUnfmd {
		for _, oneTrade := range oneRedisUnfmdTransaction.Matches {
			oneTable := tables.TableOmniTradeInfo{
				Hash:           oneRedisUnfmdTransaction.Txid,
				Txid:           oneTrade.Txid,
				Block:          oneTrade.Block,
				Address:        oneTrade.Address,
				Amountsold:     oneTrade.Amountsold,
				Amountreceived: oneTrade.Amountreceived,
				Tradingfee:     oneTrade.Tradingfee,
			}
			allTrade = append(allTrade, oneTable)
		}
		return
	}
	txSelectSql := fmt.Sprintf("select * from t_omni_trade_info where hash = '%s';", txid)
	if err := database.Db.Raw(txSelectSql).Scan(&allTrade).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", txSelectSql)
		errInfo = err
		return
	}
	return
}

func GetCancel(txid string, bUnfmd bool, oneRedisUnfmdTransaction *OmniTransaction) (errInfo error, allCancel []tables.TableOmniCancelInfo) {
	if bUnfmd {
		for _, oneCancel := range oneRedisUnfmdTransaction.Cancelledtransactions {
			oneTable := tables.TableOmniCancelInfo{
				Hash:             oneRedisUnfmdTransaction.Txid,
				Txid:             oneCancel.Txid,
				Propertyid:       oneCancel.Propertyid,
				Amountunreserved: oneCancel.Amountunreserved,
			}
			allCancel = append(allCancel, oneTable)
		}
		return
	}
	txSelectSql := fmt.Sprintf("select * from t_omni_cancel_info where hash = '%s';", txid)
	if err := database.Db.Raw(txSelectSql).Scan(&allCancel).Error; nil != err {
		log.Log.Error(err, " exec sql fail: ", txSelectSql)
		errInfo = err
		return
	}
	return
}

func ConvertToRecord(trx *tables.TableOmniTransactionInfo, bUnfmd bool, oneRedisUnfmdTransaction *OmniTransaction) (errInfo error, trxType string, receiveTime int64, trxDetail interface{}) {
	trxType = trx.Type
	if bUnfmd {
		receiveTime = oneRedisUnfmdTransaction.ReceiveTime
	}

	switch trxType {
	case "DEx Purchase":
		err, allPurchase := GetPurchase(trx.Txid, bUnfmd, oneRedisUnfmdTransaction)
		if nil != err {
			errInfo = err
			return
		}
		result := ConvertToDExPurchase(trx, allPurchase)
		trxDetail = *result
	case "Simple Send":
		result := ConvertToSimpleSend(trx)
		trxDetail = *result
	case "Send To Owners":
		err, allRecipient := GetRecipientInfo(trx.Txid, bUnfmd, oneRedisUnfmdTransaction)
		if nil != err {
			errInfo = err
			return
		}
		result := ConvertToSendToOwners(trx, allRecipient)
		trxDetail = *result
	case "Send All":
		err, allSubsend := GetSubsend(trx.Txid, bUnfmd, oneRedisUnfmdTransaction)
		if nil != err {
			errInfo = err
			return
		}
		result := ConvertToSendAll(trx, allSubsend)
		trxDetail = *result
	case "DEx Sell Offer":
		result := ConvertToDExSellOffer(trx)
		trxDetail = *result
	case "MetaDEx trade":
		err, allTrade := GetTrade(trx.Txid, bUnfmd, oneRedisUnfmdTransaction)
		if nil != err {
			errInfo = err
			return
		}
		result := ConvertToMetaDExTrade(trx, allTrade)
		trxDetail = *result
	case "MetaDEx cancel-price":
		err, allCancel := GetCancel(trx.Txid, bUnfmd, oneRedisUnfmdTransaction)
		if nil != err {
			errInfo = err
			return
		}
		result := ConvertToMetaDExCancelPrice(trx, allCancel)
		trxDetail = *result
	case "MetaDEx cancel-pair":
		err, allCancel := GetCancel(trx.Txid, bUnfmd, oneRedisUnfmdTransaction)
		if nil != err {
			errInfo = err
			return
		}
		result := ConvertToMetaDExCancelPair(trx, allCancel)
		trxDetail = *result
	case "MetaDEx cancel-ecosystem":
		err, allCancel := GetCancel(trx.Txid, bUnfmd, oneRedisUnfmdTransaction)
		if nil != err {
			errInfo = err
			return
		}
		result := ConvertToMetaDExCancelEcosystem(trx, allCancel)
		trxDetail = *result
	case "DEx Accept Offer":
		result := ConvertToDExAcceptOffer(trx)
		trxDetail = *result
	case "Create Property - Fixed":
		result := ConvertToCreatePropertyFixed(trx)
		trxDetail = *result
	case "Create Property - Variable":
		result := ConvertToCreatePropertyVariable(trx)
		trxDetail = *result
	case "Close Crowdsale":
		result := ConvertToCloseCrowdsale(trx)
		trxDetail = *result
	case "Create Property - Manual":
		result := ConvertToCreatePropertyManual(trx)
		trxDetail = *result
	case "Grant Property Tokens":
		result := ConvertToGrantPropertyTokens(trx)
		trxDetail = *result
	case "Revoke Property Tokens":
		result := ConvertToRevokePropertyTokens(trx)
		trxDetail = *result
	case "Change Issuer Address":
		result := ConvertToChangeIssuerAddress(trx)
		trxDetail = *result
	case "Enable Freezing":
		result := ConvertToEnableFreezing(trx)
		trxDetail = *result
	case "Disable Freezing":
		result := ConvertToDisableFreezing(trx)
		trxDetail = *result
	case "Freeze Property Tokens":
		result := ConvertToFreezePropertyTokens(trx)
		trxDetail = *result
	case "Unfreeze Property Tokens":
		result := ConvertToUnfreezePropertyTokens(trx)
		trxDetail = *result
	case "Feature Activation":
		result := ConvertToFeatureActivation(trx)
		trxDetail = *result

	default:
		result := ConvertToOmniCommon(trx)
		trxDetail = *result
	}
	return
}

func ConvertToMessage(record interface{}, receiveTime int64) (error, interface{}) {
	err, blockHeight := GetOmniBlockHeight()
	if nil != err {
		return err, nil
	}

	recordType := reflect.TypeOf(record)
	recordValue := reflect.ValueOf(record)

	type detail struct {
		Info          interface{} `json:"info"`
		Confirmations int32       `json:"confirmations"`
		Receivetime   string      `json:"receivetime"`
		Blocktime     string      `json:"blocktime"`
	}

	result := detail{Info: record}
	for i := 0; i < recordType.NumField(); i++ {
		key := recordType.Field(i)

		// OmniCommon
		if key.Name == "OmniCommon" {
			omniCommonInfo, ok := recordValue.Field(i).Interface().(OmniCommon)
			if !ok {
				return errors.New("wrong type"), nil
			}

			if omniCommonInfo.Block > 0 {
				result.Confirmations = blockHeight - omniCommonInfo.Block
			} else {
				result.Confirmations = -1
			}

			if omniCommonInfo.Blocktime > 0 {
				result.Blocktime = time.Unix(omniCommonInfo.Blocktime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			} else {
				result.Receivetime = time.Unix(receiveTime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			}
		}

		// SimpleSend
		if key.Name == "SimpleSend" {
			SimpleSendInfo, ok := recordValue.Field(i).Interface().(SimpleSend)
			if !ok {
				return errors.New("wrong type"), nil
			}

			if SimpleSendInfo.Block > 0 {
				result.Confirmations = blockHeight - SimpleSendInfo.Block
			} else {
				result.Confirmations = -1
			}

			if SimpleSendInfo.Blocktime > 0 {
				result.Blocktime = time.Unix(SimpleSendInfo.Blocktime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			} else {
				result.Receivetime = time.Unix(receiveTime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			}
		}

		// EnableFreezing
		if key.Name == "EnableFreezing" {
			EnableFreezingInfo, ok := recordValue.Field(i).Interface().(EnableFreezing)
			if !ok {
				return errors.New("wrong type"), nil
			}

			if EnableFreezingInfo.Block > 0 {
				result.Confirmations = blockHeight - EnableFreezingInfo.Block
			} else {
				result.Confirmations = -1
			}

			if EnableFreezingInfo.Blocktime > 0 {
				result.Blocktime = time.Unix(EnableFreezingInfo.Blocktime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			} else {
				result.Receivetime = time.Unix(receiveTime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			}
		}

		// CloseCrowdsale
		if key.Name == "CloseCrowdsale" {
			CloseCrowdsaleInfo, ok := recordValue.Field(i).Interface().(CloseCrowdsale)
			if !ok {
				return errors.New("wrong type"), nil
			}

			if CloseCrowdsaleInfo.Block > 0 {
				result.Confirmations = blockHeight - CloseCrowdsaleInfo.Block
			} else {
				result.Confirmations = -1
			}

			if CloseCrowdsaleInfo.Blocktime > 0 {
				result.Blocktime = time.Unix(CloseCrowdsaleInfo.Blocktime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			} else {
				result.Receivetime = time.Unix(receiveTime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			}
		}

		// CreatePropertyFixed
		if key.Name == "CreatePropertyFixed" {
			CreatePropertyFixedInfo, ok := recordValue.Field(i).Interface().(CreatePropertyFixed)
			if !ok {
				return errors.New("wrong type"), nil
			}

			if CreatePropertyFixedInfo.Block > 0 {
				result.Confirmations = blockHeight - CreatePropertyFixedInfo.Block
			} else {
				result.Confirmations = -1
			}

			if CreatePropertyFixedInfo.Blocktime > 0 {
				result.Blocktime = time.Unix(CreatePropertyFixedInfo.Blocktime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			} else {
				result.Receivetime = time.Unix(receiveTime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			}
		}

		// Block
		if key.Name == "Block" {
			height, ok := recordValue.Field(i).Interface().(int32)
			if !ok {
				return errors.New("wrong type"), nil
			}

			if height > 0 {
				result.Confirmations = blockHeight - height
			} else {
				result.Confirmations = -1
			}
		}

		// Blocktime
		if key.Name == "Blocktime" {
			blockTime, ok := recordValue.Field(i).Interface().(int64)
			if !ok {
				return errors.New("wrong type"), nil
			}

			if blockTime > 0 {
				result.Blocktime = time.Unix(blockTime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			} else {
				result.Blocktime = time.Unix(receiveTime, 0).UTC().Format("2006-01-02T15:04:05.999999-0700")
			}
		}
	}
	return nil, result
}
