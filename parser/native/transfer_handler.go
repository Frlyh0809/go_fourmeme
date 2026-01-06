// parser/native/transfer_handler.go
package native

import (
	"context"
	"go_fourmeme/config"
	configentity "go_fourmeme/entity/config"
	"go_fourmeme/utils"
	"math/big"
	"sync"

	"go_fourmeme/log"
	"go_fourmeme/manager"

	"github.com/ethereum/go-ethereum/core/types"
)

const (
	MinTransferBNB = 10 // >10 BNB 标记
	//gasUsed = 0x5208 ≈ 21000 （标准转账 Gas）
	MaxGasForSimple = 30000 // 简单转账 GasUsed ≤30k
	MaxConcurrent   = 20    // 最大并发数
)

var (
	classPool *utils.WorkerPool
	poolOnce  sync.Once
)

type TransferHandler struct {
	pool *utils.WorkerPool // 共享全局 pool
}

func NewTransferHandler() *TransferHandler {

	return &TransferHandler{}
}

func getClassPool() *utils.WorkerPool {
	poolOnce.Do(func() {
		classPool = utils.NewWorkerPool(MaxConcurrent)
	})
	return classPool
}

func (h *TransferHandler) ParseLogMulti(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget) {
	go func() {
		pool := getClassPool()
		pool.Acquire()
		defer pool.Release()

		h.ParseLog(allLogs, receipt, target)
	}()
}
func (h *TransferHandler) ParseLog(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget) {
	if receipt == nil || len(allLogs) > 0 {
		return
	}

	ethClient := manager.GetEthClient()
	if ethClient == nil {
		return
	}

	if receipt.ContractAddress != config.AddrZero {
		log.LogWarn("block:%d hash:%s ContractAddress: %v", receipt.BlockNumber, receipt.TxHash.Hex(), receipt.ContractAddress.Hex())

		return // 创建合约
	}
	if len(receipt.Logs) > 0 || len(allLogs) > 0 {
		return // 有 log，通常 ERC20 或合约交互
	}
	if receipt.GasUsed > MaxGasForSimple {
		return // Gas 太高，不是简单转账
	}
	// 允许 Legacy (0x0) 和 DynamicFee (0x2)
	if receipt.Type != types.LegacyTxType && receipt.Type != types.DynamicFeeTxType {
		return
	}

	// 第二层：查询完整 Transaction 获取 Value
	tx, _, err := ethClient.TransactionByHash(context.Background(), receipt.TxHash)
	if err != nil || tx == nil {
		log.LogError("TransactionByHash err for hash:%s err:%v", receipt.TxHash.Hex(), err)
		return
	}
	value := tx.Value()
	if value.Sign() <= 0 {
		return // 无 BNB 转账
	}

	to := tx.To()
	if to == nil {
		return
	}

	amountBNB := new(big.Float).Quo(new(big.Float).SetInt(value), big.NewFloat(1e18))
	amountFloat, _ := amountBNB.Float64()
	if amountFloat <= MinTransferBNB {
		return
	}

	// 查询接收者 nonce
	nonce, err := ethClient.PendingNonceAt(context.Background(), *to)
	if err != nil {
		log.LogFatal("查询 nonce 失败 To: %s: %v", to.Hex(), err)
		return
	}

	from, _ := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	blockNum := receipt.BlockNumber.Uint64()
	if nonce == 0 {
		log.LogInfo("【大额 BNB 转入 新 地址】From: %s | To: %s | Amount: %.2f BNB | Tx: %s | Block: %d | Type: 0x%x | GasUsed: %d",
			from.Hex(), to.Hex(), amountFloat, receipt.TxHash.Hex(), blockNum, receipt.Type, receipt.GasUsed)
	} else {
		log.LogInfo("【大额 BNB 转入 老 地址】From: %s | To: %s | Amount: %.2f BNB | Tx: %s | Block: %d | Type: 0x%x | GasUsed: %d",
			from.Hex(), to.Hex(), amountFloat, receipt.TxHash.Hex(), blockNum, receipt.Type, receipt.GasUsed)
	}
}
