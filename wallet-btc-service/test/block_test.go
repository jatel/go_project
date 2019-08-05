package test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/BlockABC/wallet-btc-service/notify"
)

func TestBlock(t *testing.T) {
	rawData := []byte(`{"version":536870912,"time":1555587465,"nonce":1213753007,"bits":"172c071d","merkleroot":"a517dc26248fff42847376bac4212dbbd268ffa864e70597af4bf8b8363a7f44","preblock":"000000000000000000174484ff74b6f73734d2e4e99a2b1e6b6a2c40eb970b72","hash":"00000000000000000022c97df79ff34f1cd97eea1f1436ed7488c6953fcd3373","height":572157,"nTx":2160,"difficulty":6393023717201.863,"size":1209669,"weight":3998574,"mediantime":1555583678,"chainwork":"000000000000000000000000000000000000000005df29e5758dba5b3284d18a","txs":[{"txid":"441d3906087ec353297a701183cbb07e6d19dcb5572c9bfb5da00448751d341e","hash":"f53885b34d86e63f680ca443f7454cdb38b56f7131c5df83262766dbbab0447f","version":1,"size":330,"vsize":303,"weight":1212,"locktime":973614462,"vin":[{"coinbase":"03fdba082cfabe6d6d246dae150c3d42ad5934b00c8199f5199fd1f06553273b134fd977295154b62b10000000f09f909f00184d696e6564206279206631343230786474316c7331383035000000000000000000000000000000000000000500ca210000","sequence":0}],"vout":[{"value":13.11699371,"n":0,"scriptPubKey":{"asm":"OP_DUP OP_HASH160 c825a1ecf2a6830c4401620c3a16f1995057c2ab OP_EQUALVERIFY OP_CHECKSIG","hex":"76a914c825a1ecf2a6830c4401620c3a16f1995057c2ab88ac","reqSigs":1,"type":"pubkeyhash","addresses":["1KFHE7w8BhaENAswwryaoccDb6qcT6DbYY"]}},{"value":0.00000000,"n":1,"scriptPubKey":{"asm":"OP_RETURN aa21a9edb7e0b2be480ae72894bf3e79050e5f70510793ac3f2a2997e96e722875104cce 0000000000000000","hex":"6a24aa21a9edb7e0b2be480ae72894bf3e79050e5f70510793ac3f2a2997e96e722875104cce080000000000000000","type":"nulldata"}},{"value":0.00000000,"n":2,"scriptPubKey":{"asm":"OP_RETURN 52534b424c4f434b3a5e7a1f2c410ec3cf2f65e4c68eaaa6d4720f8266d8f0b5fb75c2632bd5808b4d","hex":"6a4c2952534b424c4f434b3a5e7a1f2c410ec3cf2f65e4c68eaaa6d4720f8266d8f0b5fb75c2632bd5808b4d","type":"nulldata"}}],"hex":"010000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff6403fdba082cfabe6d6d246dae150c3d42ad5934b00c8199f5199fd1f06553273b134fd977295154b62b10000000f09f909f00184d696e6564206279206631343230786474316c7331383035000000000000000000000000000000000000000500ca2100000000000003abf12e4e000000001976a914c825a1ecf2a6830c4401620c3a16f1995057c2ab88ac00000000000000002f6a24aa21a9edb7e0b2be480ae72894bf3e79050e5f70510793ac3f2a2997e96e722875104cce08000000000000000000000000000000002c6a4c2952534b424c4f434b3a5e7a1f2c410ec3cf2f65e4c68eaaa6d4720f8266d8f0b5fb75c2632bd5808b4d012000000000000000000000000000000000000000000000000000000000000000007e2d083a"},{"txid":"5d01a545cdaa49bb8b7d9e87c361b9c8f1e8957f9dd3ea8efebab21e740d140d","hash":"41d067d743f27db9254a1e3de0bbf2c97f1e49792df76775db9e4aba4c4edb61","version":2,"size":247,"vsize":166,"weight":661,"locktime":572156,"vin":[{"txid":"9b874f673a82dfdb3cb7749130542a09c11bf7d3325dfe14a6df77a811991f43","vout":0,"scriptSig":{"asm":"001454155c39b63deb07e59eefd450dea7ffade7e512","hex":"16001454155c39b63deb07e59eefd450dea7ffade7e512"},"txinwitness":["304402206516e15dc213174efb31ff59a944786b398e7a9cd68ad9a8266883b2048adc54022063eea42bf4f424923dc36352837bc7b987a70905954bb3253f1f100daa732f9501","023933e382eda53641220dfa5a767d129b26844d805606a883b4c464efec55e5a2"],"sequence":4294967294}],"vout":[{"value":0.88183451,"n":0,"scriptPubKey":{"asm":"OP_HASH160 463f1e26b4600fabcd498cb55103e77d70078b2a OP_EQUAL","hex":"a914463f1e26b4600fabcd498cb55103e77d70078b2a87","reqSigs":1,"type":"scripthash","addresses":["386SqUL262Bmnw6tzxGjgqtfZkJXChyBJV"]}},{"value":6.20938637,"n":1,"scriptPubKey":{"asm":"OP_HASH160 fd5995aa9b55a9d0c4458a404f64050216390686 OP_EQUAL","hex":"a914fd5995aa9b55a9d0c4458a404f6405021639068687","reqSigs":1,"type":"scripthash","addresses":["3QncCXUzA9XxsAJKjWyMCnNGLzUCKt3j3N"]}}],"hex":"02000000000101431f9911a877dfa614fe5d32d3f71bc1092a54309174b73cdbdf823a674f879b000000001716001454155c39b63deb07e59eefd450dea7ffade7e512feffffff029b9241050000000017a914463f1e26b4600fabcd498cb55103e77d70078b2a878dc502250000000017a914fd5995aa9b55a9d0c4458a404f64050216390686870247304402206516e15dc213174efb31ff59a944786b398e7a9cd68ad9a8266883b2048adc54022063eea42bf4f424923dc36352837bc7b987a70905954bb3253f1f100daa732f950121023933e382eda53641220dfa5a767d129b26844d805606a883b4c464efec55e5a2fcba0800"}]}`)
	var oneBlock notify.Block
	if err := json.Unmarshal(rawData, &oneBlock); nil != err {
		fmt.Println(err)
	}
	fmt.Println(oneBlock)

	fmt.Println("")

	fmt.Println(fmt.Sprintf("%f", oneBlock.Difficulty))
}

func convertValue(value float64) (int64, error) {
	strValue := strconv.FormatFloat(value, 'f', 30, 64)
	fmt.Println(strValue)
	all := strings.Split(strValue, ".")
	head := all[0]
	fmt.Println(head)
	other := all[1][0:8]
	fmt.Println(other)
	return strconv.ParseInt(head+other, 10, 64)
}

func TestTransaction(t *testing.T) {
	rawData := []byte(`{"txid":"86d98c59c4c4c27cccfab314e2a43418600e4d229e18d982556e46e83f062f68","hash":"e6db134cd83af7e4b5e0b29907133d304719033c2d7e09335c6bba901ef8c9ef","version":1,"size":1132,"vsize":561,"weight":2242,"locktime":0,"vin":[{"txid":"80fa1a41b3bdc09adb791a2389dd3fc5ebd1689dd56d77707efc049c0cbd2182","vout":1,"scriptSig":{"asm":"00202035e78a6b34d583ad0a874b80ee41a0507b92cf0b671656108f594c7f75ca90","hex":"2200202035e78a6b34d583ad0a874b80ee41a0507b92cf0b671656108f594c7f75ca90"},"txinwitness":["","3045022100eb7390e1b9aa6a24ca7c3a519f9991a27195a5228042b08da506d61c4b6e3923022047233739f820983b19caf99a2c3f5a2e9d241c56e594980ea5db0be16120b68a01","30450221008f406e69f7824f592c03d1fe58a509102d2022e7e893fa028bd160d9b062ad630220294edaeda674dbd088c64da11e41bf37fdd8b910d908868690ce8c1720d571d901","5221029103d1dfbbee9ea5249ee0b03ca59e08291ce34a7467513edf8ea767b5aa26382103dce07bea5905a1c3e70f86c1f74f0e98e7cf3b6f5d02226a4c531c9e930c613b210268d8878afaf4b55118519d8520fe0db27f9596a812d4378f4bf4a96a5333694653ae"],"sequence":4294967295},{"txid":"7e67b8ce7cd8d7078d7a432795115319dfed10a24cc2c3579ab6be73de2acf5e","vout":2,"scriptSig":{"asm":"002065d416c48a8072e0ac51c2d111eb194f009caef0332446c1bf2097316cf07fa9","hex":"22002065d416c48a8072e0ac51c2d111eb194f009caef0332446c1bf2097316cf07fa9"},"txinwitness":["","304402203790092bdd19287fde498042da8090259cd94c5993efff3b4348dc3eb48ab40f0220516913cc32ad720ea9e009414ea051c5aee19b8553f40c620429da263f84e8da01","304402205c53ab18f74405280c90dde36ccc6c0d42e3fee9f8e6218dd68494c3b52040ba02205d79c591ee034cf9c09e345c0f78688a75bcbf4cae020819805c30a181585a1301","522102f44abcf9e23c9a460da309ccca56c619c04eed3bde2c2cff5e7d78fbcd980b9c2103c9443cf3047bb6c2c82f1b0c44c36109cdc3d0d601d16d1189a1602bf8d1a0a02103bfe867059274412412e088af5572b92168c2ef495cfe6c9b7a753a009eb37c4853ae"],"sequence":4294967295},{"txid":"7e67b8ce7cd8d7078d7a432795115319dfed10a24cc2c3579ab6be73de2acf5e","vout":0,"scriptSig":{"asm":"00202035e78a6b34d583ad0a874b80ee41a0507b92cf0b671656108f594c7f75ca90","hex":"2200202035e78a6b34d583ad0a874b80ee41a0507b92cf0b671656108f594c7f75ca90"},"txinwitness":["","3045022100d3f7cc3ee8d6c9952c045fddf9096ee32931b8f44d321ec60c5e072a1a23801002204204692bf8022b541154b73cee9a414efad9facfa520660e8877e194122d8c8401","3045022100f8677fcc9d1f572100a37ea325b75839fa8e5518faf049e609ed4ef495b6ee46022008dd2fb313469c6bbfbc62d3e9779fd13da06bd1fafda5a0829a1f8887df5c7101","5221029103d1dfbbee9ea5249ee0b03ca59e08291ce34a7467513edf8ea767b5aa26382103dce07bea5905a1c3e70f86c1f74f0e98e7cf3b6f5d02226a4c531c9e930c613b210268d8878afaf4b55118519d8520fe0db27f9596a812d4378f4bf4a96a5333694653ae"],"sequence":4294967295}],"vout":[{"value":1.12486721,"n":0,"scriptPubKey":{"asm":"OP_HASH160 b8a9a8ba8cf965b7df6b05afd948e53c351b2c0d OP_EQUAL","hex":"a914b8a9a8ba8cf965b7df6b05afd948e53c351b2c0d87","reqSigs":1,"type":"scripthash","addresses":["3JXRVxhrk2o9f4w3cQchBLwUeegJBj6BEp"]}},{"value":6.64002000,"n":1,"scriptPubKey":{"asm":"OP_DUP OP_HASH160 37ea25eb33e4e5cff431b56b119ba27869eb1a8d OP_EQUALVERIFY OP_CHECKSIG","hex":"76a91437ea25eb33e4e5cff431b56b119ba27869eb1a8d88ac","reqSigs":1,"type":"pubkeyhash","addresses":["166efSwdeNS6WVb5LQiaCrmp6tTAEdHf9F"]}},{"value":0.60459499,"n":2,"scriptPubKey":{"asm":"OP_DUP OP_HASH160 e2091e35218d3513534afc0c6fe7842029461081 OP_EQUALVERIFY OP_CHECKSIG","hex":"76a914e2091e35218d3513534afc0c6fe784202946108188ac","reqSigs":1,"type":"pubkeyhash","addresses":["1McAdtEQvHUVRfqDQFe16moYJbua32PbRm"]}},{"value":1.12486720,"n":3,"scriptPubKey":{"asm":"OP_HASH160 1988a27e3c2df4ddee7fad5a2303d086179b2a30 OP_EQUAL","hex":"a9141988a27e3c2df4ddee7fad5a2303d086179b2a3087","reqSigs":1,"type":"scripthash","addresses":["3422VtS7UtCvXYxoXMVp6eZupR252z85oC"]}}],"hex":"010000000001038221bd0c9c04fc7e70776dd59d68d1ebc53fdd89231a79db9ac0bdb3411afa8001000000232200202035e78a6b34d583ad0a874b80ee41a0507b92cf0b671656108f594c7f75ca90ffffffff5ecf2ade73beb69a57c3c24ca210eddf1953119527437a8d07d7d87cceb8677e020000002322002065d416c48a8072e0ac51c2d111eb194f009caef0332446c1bf2097316cf07fa9ffffffff5ecf2ade73beb69a57c3c24ca210eddf1953119527437a8d07d7d87cceb8677e00000000232200202035e78a6b34d583ad0a874b80ee41a0507b92cf0b671656108f594c7f75ca90ffffffff044169b4060000000017a914b8a9a8ba8cf965b7df6b05afd948e53c351b2c0d87d0dd9327000000001976a91437ea25eb33e4e5cff431b56b119ba27869eb1a8d88aceb899a03000000001976a914e2091e35218d3513534afc0c6fe784202946108188ac4069b4060000000017a9141988a27e3c2df4ddee7fad5a2303d086179b2a30870400483045022100eb7390e1b9aa6a24ca7c3a519f9991a27195a5228042b08da506d61c4b6e3923022047233739f820983b19caf99a2c3f5a2e9d241c56e594980ea5db0be16120b68a014830450221008f406e69f7824f592c03d1fe58a509102d2022e7e893fa028bd160d9b062ad630220294edaeda674dbd088c64da11e41bf37fdd8b910d908868690ce8c1720d571d901695221029103d1dfbbee9ea5249ee0b03ca59e08291ce34a7467513edf8ea767b5aa26382103dce07bea5905a1c3e70f86c1f74f0e98e7cf3b6f5d02226a4c531c9e930c613b210268d8878afaf4b55118519d8520fe0db27f9596a812d4378f4bf4a96a5333694653ae040047304402203790092bdd19287fde498042da8090259cd94c5993efff3b4348dc3eb48ab40f0220516913cc32ad720ea9e009414ea051c5aee19b8553f40c620429da263f84e8da0147304402205c53ab18f74405280c90dde36ccc6c0d42e3fee9f8e6218dd68494c3b52040ba02205d79c591ee034cf9c09e345c0f78688a75bcbf4cae020819805c30a181585a130169522102f44abcf9e23c9a460da309ccca56c619c04eed3bde2c2cff5e7d78fbcd980b9c2103c9443cf3047bb6c2c82f1b0c44c36109cdc3d0d601d16d1189a1602bf8d1a0a02103bfe867059274412412e088af5572b92168c2ef495cfe6c9b7a753a009eb37c4853ae0400483045022100d3f7cc3ee8d6c9952c045fddf9096ee32931b8f44d321ec60c5e072a1a23801002204204692bf8022b541154b73cee9a414efad9facfa520660e8877e194122d8c8401483045022100f8677fcc9d1f572100a37ea325b75839fa8e5518faf049e609ed4ef495b6ee46022008dd2fb313469c6bbfbc62d3e9779fd13da06bd1fafda5a0829a1f8887df5c7101695221029103d1dfbbee9ea5249ee0b03ca59e08291ce34a7467513edf8ea767b5aa26382103dce07bea5905a1c3e70f86c1f74f0e98e7cf3b6f5d02226a4c531c9e930c613b210268d8878afaf4b55118519d8520fe0db27f9596a812d4378f4bf4a96a5333694653ae00000000"}`)
	var oneTransaction notify.Transaction

	if err := json.Unmarshal(rawData, &oneTransaction); nil != err {
		fmt.Println(err)
	}

	fmt.Println(oneTransaction)
	notify.SaveUnconfirmedTransactionToRedis(&oneTransaction)
}