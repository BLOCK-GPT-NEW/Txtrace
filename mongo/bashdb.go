package mongo

// Databse 1, store the basic transaction metadata
type Transac struct {
	// Transaction
	// Tx_BlockHash string
	Tx_BlockNum uint64
	Tx_FromAddr string
	Tx_ToAddr   string
	Tx_Gas      uint64
	// Tx_GasPrice  string
	Tx_Hash  string
	Tx_Input string
	Tx_Nonce uint64
	Tx_Index int
	Tx_Value string

	Tx_Trace           string
	Log_Trace          string
	Re_contractAddress string
	// Re_CumulativeGasUsed string
	// Re_GasUsed           string
	Re_Status uint64
	// Re_FailReason        string

	Re_Log_Address string
	Re_Log_Topics  string
	Re_Log_Data    string
}

type Log struct {
	Tx_Hash   string
	Log_Trace string
}

// 默认为50， 这里改成1仅为了测试
var BashNum int = 50
var BashTxs = make([]interface{}, BashNum)
var BashLogs = make([]interface{}, BashNum)
var CurrentNum int = 0
var Current_Log_Num int = 0
