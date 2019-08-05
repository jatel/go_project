package test

func TestStoreFile(t *testing.T) {
	// set block repair begin
	var blockHeight int32 = 0
	if err, height := filestore.RepairStoreInstance.GetBlockBegin(); nil == err {
		fmt.Println("block begin:", height)
		blockHeight = height
	} else {
		fmt.Println(err)
	}

	// set omni repair begin
	var omniHeight int32 = 0
	if err, height := filestore.RepairStoreInstance.GetOmniBegin(); nil == err {
		fmt.Println("omni begin:", height)
		omniHeight = height
	}else {
		fmt.Println(err)
	}
		
	if err := filestore.RepairStoreInstance.SaveBlockBegin(blockHeight+10); nil != err {
		fmt.Println(err)
	}

	if err := filestore.RepairStoreInstance.SaveOmniBegin(omniHeight+20); nil != err {
		fmt.Println(err)
	}
}