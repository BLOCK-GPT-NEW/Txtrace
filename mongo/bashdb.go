package mongo

// Databse 1, store the basic transaction metadata
type Transac struct {
	// Transaction
	// Tx_BlockHash string
	Tx_BlockNum uint64
	Tx_FromAddr string
	Tx_Gas      uint64
	// Tx_GasPrice  string
	Tx_Hash   string
	Tx_Input  string
	Tx_Nonce  uint64
	Tx_ToAddr string
	Tx_Index  string
	Tx_Value  string

	Tx_Trace string

	Re_contractAddress string
	// Re_CumulativeGasUsed string
	// Re_GasUsed           string
	Re_Status string
	// Re_FailReason        string

	Re_Log_Address string
	Re_Log_Topics  string
	Re_Log_Data    string
}

/* type Trace struct {
	Tx_Trace string
	Tx_Hash  string
}
*/
// 默认为50， 这里改成1仅为了测试
var BashNum int = 1
var BashTxs = make([]interface{}, BashNum)
var CurrentNum int = 0
