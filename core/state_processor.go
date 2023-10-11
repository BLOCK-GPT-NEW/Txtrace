// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/consensus/misc/eip4844"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/mongo"
	"github.com/ethereum/go-ethereum/params"
	// [swx]
)

//[end]

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error) {
	var (
		receipts    types.Receipts
		usedGas     = new(uint64)
		header      = block.Header()
		blockHash   = block.Hash()
		blockNumber = block.Number()
		allLogs     []*types.Log
		gp          = new(GasPool).AddGas(block.GasLimit())
	)
	// Mutate the block and state according to any hard-fork specs
	if p.config.DAOForkSupport && p.config.DAOForkBlock != nil && p.config.DAOForkBlock.Cmp(block.Number()) == 0 {
		misc.ApplyDAOHardFork(statedb)
	}
	var (
		context = NewEVMBlockContext(header, p.bc, nil)
		vmenv   = vm.NewEVM(context, vm.TxContext{}, statedb, p.config, cfg)
		signer  = types.MakeSigner(p.config, header.Number, header.Time)
	)
	if beaconRoot := block.BeaconRoot(); beaconRoot != nil {
		ProcessBeaconBlockRoot(*beaconRoot, vmenv, statedb)
	}
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		msg, err := TransactionToMessage(tx, signer, header.BaseFee)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("could not apply tx %d [%v]: %w", i, tx.Hash().Hex(), err)
		}
		statedb.SetTxContext(tx.Hash(), i)
		receipt, err := applyTransaction(msg, p.config, gp, statedb, blockNumber, blockHash, tx, usedGas, vmenv)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("could not apply tx %d [%v]: %w", i, tx.Hash().Hex(), err)
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}
	// Fail if Shanghai not enabled and len(withdrawals) is non-zero.
	withdrawals := block.Withdrawals()
	if len(withdrawals) > 0 && !p.config.IsShanghai(block.Number(), block.Time()) {
		return nil, nil, 0, errors.New("withdrawals before shanghai")
	}
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	p.engine.Finalize(p.bc, header, statedb, block.Transactions(), block.Uncles(), withdrawals)

	return receipts, allLogs, *usedGas, nil
}

func applyTransaction(msg *Message, config *params.ChainConfig, gp *GasPool, statedb *state.StateDB, blockNumber *big.Int, blockHash common.Hash, tx *types.Transaction, usedGas *uint64, evm *vm.EVM) (*types.Receipt, error) {

	// cnz 处理TxHashGlobal
	mongo.TxHashGlobal.Reset()
	mongo.TxHashGlobal.WriteString(tx.Hash().Hex())
	//[swx]
	// Check for nil pointers to avoid nil pointer dereference

	if msg == nil || config == nil || gp == nil || statedb == nil || blockNumber == nil || tx == nil || usedGas == nil || evm == nil {
		log.Println("Error: state_processor.go applyTransaction line 118 received a nil parameter")
		return nil, errors.New("received a nil parameter")
	}

	mongo.TraceGlobal.Reset()
	mongo.TxVMErr = ""
	//[end]

	// Create a new context to be used in the EVM environment.
	txContext := NewEVMTxContext(msg)
	evm.Reset(txContext, statedb)

	// Apply the transaction to the current state (included in the env).
	result, err := ApplyMessage(evm, msg, gp)
	if err != nil {
		return nil, err
	}
	//[swx]
	mongo.TraceGlobal.Reset()
	//[end]

	// Update the state with pending changes.
	var root []byte
	if config.IsByzantium(blockNumber) {
		statedb.Finalise(true)
	} else {
		root = statedb.IntermediateRoot(config.IsEIP158(blockNumber)).Bytes()
	}
	*usedGas += result.UsedGas

	// Create a new receipt for the transaction, storing the intermediate root and gas used
	// by the tx.
	receipt := &types.Receipt{Type: tx.Type(), PostState: root, CumulativeGasUsed: *usedGas}
	if result.Failed() {
		receipt.Status = types.ReceiptStatusFailed
	} else {
		receipt.Status = types.ReceiptStatusSuccessful
	}
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = result.UsedGas

	if tx.Type() == types.BlobTxType {
		receipt.BlobGasUsed = uint64(len(tx.BlobHashes()) * params.BlobTxBlobGasPerBlob)
		receipt.BlobGasPrice = eip4844.CalcBlobFee(*evm.Context.ExcessBlobGas)
	}

	// If the transaction created a contract, store the creation address in the receipt.
	var toAddress string
	if msg.To == nil {
		// receipt.ContractAddress = crypto.CreateAddress(evm.TxContext.Origin, tx.Nonce())
		contractAddress := crypto.CreateAddress(evm.TxContext.Origin, tx.Nonce())
		receipt.ContractAddress = contractAddress
		toAddress = contractAddress.Hex()
		//[swx]
	} else {
		toAddress = msg.To.Hex()
	}

	// Set the receipt logs and create the bloom filter.
	receipt.Logs = statedb.GetLogs(tx.Hash(), blockNumber.Uint64(), blockHash)
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	receipt.BlockHash = blockHash
	receipt.BlockNumber = blockNumber
	receipt.TransactionIndex = uint(statedb.TxIndex())
	// cnz
	/*
		var log_data []byte
		var log_hash []string
		// 遍历 Receipt 结构体中的 Logs 切片
		for _, logEntry := range receipt.Logs {
			// logEntry.Data =
			// 提取每个 Log 结构体的 字段值，并将其添加到 切片中
			log_data = append(log_data, logEntry.Data...)
			log_hash = append(log_hash, logEntry.Address.Hex())
		}
	*/
	//[swx]
	// Check if ClientGlobal is nil and try to reconnect
	if mongo.ClientGlobal == nil {
		var recon_err error
		// Code to re-initialize the MongoDB client goes here
		// ...
		if recon_err != nil {
			log.Printf("Failed to reconnect to MongoDB: %v", recon_err)
		}
	}

	// 处理input
	// 处理data
	var res1 []string
	var resString string
	// 遍历 []byte 中的每个字节
	for _, b := range tx.Data() {
		// 使用 encoding/hex 包将字节转换为十六进制字符串
		hexString := hex.EncodeToString([]byte{b})
		// 将结果添加到 res1 切片中
		res1 = append(res1, hexString)
	}
	resString = strings.Join(res1, ",")
	/*// 将 res1 切片中的十六进制字符串连接为一个字符串
	finalResult := ""
	for _, hexStr := range res1 {
		finalResult += hexStr
	}*/

	// 构造交易结构体
	mongo.BashTxs[mongo.CurrentNum] = mongo.Transac{
		// Tx_BlockHash: blockHash.Hex(),
		// Tx_BlockNum:  blockNumber.Uint64(),
		Tx_FromAddr: msg.From.Hex(),
		Tx_Gas:      fmt.Sprint(result.UsedGas),
		// Tx_GasPrice:  msg.GasPrice.String(),
		Tx_Hash:  tx.Hash().Hex(),
		Tx_Input: resString,
		// Tx_Nonce:     tx.Nonce(),
		Tx_ToAddr: toAddress, // Will be empty if contract creation
		Tx_Index:  fmt.Sprint(statedb.TxIndex()),
		Tx_Value:  msg.Value.String(),

		// Tx_Trace:             vm.OpCode.String(),
		Re_contractAddress: receipt.ContractAddress.Hex(),
		// Re_CumulativeGasUsed: fmt.Sprint(receipt.CumulativeGasUsed),
		// Re_GasUsed:           fmt.Sprint(receipt.GasUsed),
		Re_Status: fmt.Sprint(receipt.Status),

		// Re_Logs: log_data,
		// Re_Hash: log_hash,
	}

	if mongo.CurrentNum != mongo.BashNum-1 {
		mongo.CurrentNum = mongo.CurrentNum + 1
	} else {
		collection := mongo.ClientGlobal.Database("geth").Collection("transaction")
		_, err := collection.InsertMany(context.Background(), mongo.BashTxs)
		if err != nil {
			// 日志记录或错误处理
			log.Printf("Failed to insert transactions: %v", err)
			// Convert the failed transaction data to JSON and write to an error file
			for _, txInterface := range mongo.BashTxs {
				if tx, ok := txInterface.(mongo.Transac); ok {
					json_tx, json_err := json.Marshal(tx)
					if json_err != nil {
						// Assuming ErrorFile is a global variable for error logging
						mongo.ErrorFile.WriteString(fmt.Sprintf("Transaction;%s;%s\n", tx.Tx_Hash, json_err))
					}
					mongo.ErrorFile.WriteString(fmt.Sprintf("Transaction|%s|%s\n", string(json_tx), err))
				} else {
					mongo.ErrorFile.WriteString(fmt.Sprintf("Failed to assert type for transaction: %v\n", txInterface))
				}
			}
		}
		mongo.CurrentNum = 0
	}

	//[end]

	return receipt, err
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, error) {
	msg, err := TransactionToMessage(tx, types.MakeSigner(config, header.Number, header.Time), header.BaseFee)
	if err != nil {
		return nil, err
	}
	// Create a new context to be used in the EVM environment
	blockContext := NewEVMBlockContext(header, bc, author)
	vmenv := vm.NewEVM(blockContext, vm.TxContext{BlobHashes: tx.BlobHashes()}, statedb, config, cfg)

	return applyTransaction(msg, config, gp, statedb, header.Number, header.Hash(), tx, usedGas, vmenv)
}

// ProcessBeaconBlockRoot applies the EIP-4788 system call to the beacon block root
// contract. This method is exported to be used in tests.
func ProcessBeaconBlockRoot(beaconRoot common.Hash, vmenv *vm.EVM, statedb *state.StateDB) {
	// If EIP-4788 is enabled, we need to invoke the beaconroot storage contract with
	// the new root
	msg := &Message{
		From:      params.SystemAddress,
		GasLimit:  30_000_000,
		GasPrice:  common.Big0,
		GasFeeCap: common.Big0,
		GasTipCap: common.Big0,
		To:        &params.BeaconRootsStorageAddress,
		Data:      beaconRoot[:],
	}
	vmenv.Reset(NewEVMTxContext(msg), statedb)
	statedb.AddAddressToAccessList(params.BeaconRootsStorageAddress)
	_, _, _ = vmenv.Call(vm.AccountRef(msg.From), *msg.To, msg.Data, 30_000_000, common.Big0)
	statedb.Finalise(true)
}
