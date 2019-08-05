package test

import (
	"fmt"
	"testing"

	"github.com/BlockABC/wallet-btc-service/request"
)

func TestBlock(t *testing.T) {
	ids := []string{"bitcoin", "tether"}
	err, result := request.GetPrice(ids)
	if nil != err {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
}
