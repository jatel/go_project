package notify

import (
	"context"
	"net/http"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/gin-gonic/gin"
)

func StartHttpServer(cfg *config.Config, ctx context.Context) (err error) {
	//get router instance
	router := gin.Default()

	// handle new block
	router.POST("/bitcoind/block", bitcoindBlockNotify)

	// handle new transaction
	router.POST("/bitcoind/transaction", bitcoindTransactionNotify)

	// listen and server
	http.ListenAndServe(cfg.BtcOpt.ListenAddress, router)
	return nil
}
