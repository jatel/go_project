package notify

var RepairMag RepairManage

type RepairManage struct {
	Block       *BlockRepair
	Transaction *TransactionRepair
}

func InitRepairManager() error {
	RepairMag.Block = NewBlockRepair()
	RepairMag.Transaction = NewTransactionRepair()
	return nil
}

func StoreFailBlockHash(hash string) error {
	return RepairMag.Block.StoreFailHash(hash)
}

func BatchStoreBlockHash(allHash []string) error {
	return RepairMag.Block.BatchStoreFailHash(allHash)
}

func RepairBlocks() error {
	return RepairMag.Block.RepairAllItems()
}

func StoreFailTransactionHash(height int32, hash string) error {
	return RepairMag.Transaction.StoreFailHash(height, hash)
}

func BatchStoreTransactionHash(allHeight []int32, allHash []string) error {
	return RepairMag.Transaction.BatchStoreFailHash(allHeight, allHash)
}

func RepairTransactions() error {
	return RepairMag.Transaction.RepairAllItems()
}

