package mongo

import (
	"bytes"
	"context"
	"log"
	"os"

	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ClientGlobal *mongo.Client // 替代原来的 *mgo.Session
var TraceGlobal = bytes.NewBuffer(make([]byte, 0, 10485760))

// 私有包级变量
var traceGlobalMutex sync.Mutex

// Exported functions to control the access to the mutex
// LockTraceGlobalMutex 提供了一个锁定互斥锁的公共方法
func LockTraceGlobalMutex() {
    traceGlobalMutex.Lock()
}

// UnlockTraceGlobalMutex 提供了一个解锁互斥锁的公共方法
func UnlockTraceGlobalMutex() {
    traceGlobalMutex.Unlock()
}

var LogGlobal = bytes.NewBuffer(make([]byte, 0, 10485760))
var CurrentTx string
var CurrentBlockNum uint64
var TxVMErr string
var ErrorFile *os.File

func InitMongoDb() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://b515:WdnPQyE9jzFitkMs@10.12.46.32:27020")

	// Connect to MongoDB
	ClientGlobal, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Check the connection
	err = ClientGlobal.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	ErrorFile, err = os.OpenFile("db_error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open error file: %v", err)
	}
}
