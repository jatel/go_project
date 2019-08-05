package notify

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/database"
	"github.com/BlockABC/wallet-btc-service/database/tables"
	"github.com/BlockABC/wallet-btc-service/omni"
	"github.com/gin-gonic/gin"
)

type Block struct {
	Version           int32         `json:"version"`
	Time              int64         `json:"time"`
	Nonce             int64         `json:"nonce"`
	Bits              string        `json:"bits"`
	Merkleroot        string        `json:"merkleroot"`
	Previousblockhash string        `json:"previousblockhash"`
	Hash              string        `json:"hash"`
	Height            int32         `json:"height"`
	NTx               int64         `json:"nTx"`
	Difficulty        float64       `json:"difficulty"`
	Size              int32         `json:"size"`
	Weight            int64         `json:"weight"`
	Mediantime        int64         `json:"mediantime"`
	Chainwork         string        `json:"chainwork"`
	Tx                []Transaction `json:"tx"`
}

const (
	MAXBLOCKLEN  = 500
	RESERVEBLOCK = 100
)

var (
	allBlockHash    []string
	blockManagechan = make(chan string, config.Cfg.BtcOpt.BlockGoroutineNum)
)

func newBlockCome(newBlock *Block) error {
	// exist
	for _, oneHash := range allBlockHash {
		if oneHash == newBlock.Hash {
			return fmt.Errorf("%s already exists", newBlock.Hash)
		}
	}

	// apned all block
	if len(allBlockHash) >= MAXBLOCKLEN {
		allBlockHash = allBlockHash[len(allBlockHash)-MAXBLOCKLEN+RESERVEBLOCK:]
	}
	allBlockHash = append(allBlockHash, newBlock.Hash)

	// write to channel
	blockManagechan <- newBlock.Hash

	// save block
	go func() {
		defer func() {
			<-blockManagechan

			// handle panic
			if err := recover(); err != nil {
				log.Log.Error(err, " panic occur when bitcoind block notify, block hash: ", newBlock.Hash)
			}
		}()

		// save block
		SaveBlock(newBlock)

		// process block omni transactions
		omni.HandleOmniBlock(newBlock.Height)
	}()

	return nil
}

func bitcoindBlockNotify(c *gin.Context) {
	var newBlock Block
	if err := c.BindJSON(&newBlock); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	log.Log.Info("New block notifications received, block hash:", newBlock.Hash)

	if err := newBlockCome(&newBlock); nil != err {
		c.JSON(http.StatusBadRequest, gin.H{"result": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func SaveBlock(newBlock *Block) error {
	// save block not update state and from
	if err := SaveBlockNotUpdateStateAndFrom(newBlock); nil != err {
		log.Log.Error(err, " save block not update state and from fail, block height: ", newBlock.Height, ", block hash: ", newBlock.Hash)
		return err
	}

	// update state and from
	updateStateAndFrom(newBlock.Tx)

	return nil
}

func SaveBlockNotUpdateStateAndFrom(newBlock *Block) error {
	// handle block
	oneBlock := tables.TableBlockInfo{
		Height:            newBlock.Height,
		Version:           int(newBlock.Version),
		Time:              newBlock.Time,
		Bits:              newBlock.Bits,
		Nonce:             newBlock.Nonce,
		Difficulty:        newBlock.Difficulty,
		Size:              newBlock.Size,
		Weight:            newBlock.Weight,
		Mediantime:        newBlock.Mediantime,
		Chainwork:         newBlock.Chainwork,
		Hash:              newBlock.Hash,
		Merkleroot:        newBlock.Merkleroot,
		Previousblockhash: newBlock.Previousblockhash,
		Ntx:               int64(len(newBlock.Tx)),
	}

	// save block fail
	if err := database.Db.Create(&oneBlock).Error; nil != err {
		log.Log.Error(err, " insert into t_block_info fail, block height: ", newBlock.Height, ", block hash: ", newBlock.Hash)
		StoreFailBlockHash(newBlock.Hash)
		return err
	}

	log.Log.Info("insert into t_block_info sucess, block height: ", newBlock.Height, ", block hash: ", newBlock.Hash)

	// set transaction blockhash, blockheight and blocktime
	for index, _ := range newBlock.Tx {
		newBlock.Tx[index].BlockHash = newBlock.Hash
		newBlock.Tx[index].BlockTime = newBlock.Time
		newBlock.Tx[index].Blockheight = newBlock.Height
	}

	// handle transaction
	wg := sync.WaitGroup{}
	trxCh := make(chan int, config.TrxGoroutineRuntime)
	for index, tx := range newBlock.Tx {
		if index >= config.TrxGoroutineRuntime {
			<-trxCh
		}
		wg.Add(1)
		go transactionTask(index, tx, trxCh, &wg)
	}
	wg.Wait()

	// delete unconfirmed transaction
	DeleteRedisBlockTransaction(newBlock)

	return nil
}

func transactionTask(txIndex int, oneTrx Transaction, ch chan<- int, group *sync.WaitGroup) {
	defer func() {
		ch <- txIndex
		group.Done()

		// handle panic
		if err := recover(); err != nil {
			log.Log.Error(err, " transactionTask panic when save one block transaction, block height:", oneTrx.Blockheight, ", transaction hash: ", oneTrx.Txid)
			StoreFailTransactionHash(oneTrx.Blockheight, oneTrx.Txid)
		}
	}()

	// save transaction
	if err := SaveTransaction(&oneTrx); nil != err {
		StoreFailTransactionHash(oneTrx.Blockheight, oneTrx.Txid)
		log.Log.Error(err, " save transaction fail when save block, block height: ", oneTrx.Blockheight, ", transaction hash: ", oneTrx.Txid)
	} else {
		log.Log.Info("save transaction success when save block, block height: ", oneTrx.Blockheight, ", transaction hash: ", oneTrx.Txid)
	}
}
