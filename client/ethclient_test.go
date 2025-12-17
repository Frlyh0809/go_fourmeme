// client/ethclient_test.go
package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"go_fourmeme/config"

	//"fmt"
	"math/big"
	"testing"
	"time"

	"go_fourmeme/client" // 替换为你的实际包路径

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	//"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	//"go_fourmeme/log"    // 如果没有自定义 log，可直接用 fmt.Printf
)

// 一些常用测试地址
var (
	testAddress       = common.HexToAddress("0x000000000000000000000000000000000000dEaD")
	pancakeRouterTest = common.HexToAddress("0x9Ac64Cc6e4415144C455BD8E4837Fea55603e5C3") // 测试网 PancakeRouter
	wbnbTest          = common.HexToAddress("0xae13d989daC2f0dEBFf460aC112a837C89bAa7cd")
	sampleTxHash      = common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef") // 可替换为真实存在的交易
)

// 统一的 JSON 打印函数
func printJSON(t *testing.T, name string, result interface{}, err error) {
	if err != nil {
		t.Errorf("✗ %s 失败: %v", name, err)
		return
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	t.Logf("✓ %s 成功，结果:\n%s", name, string(jsonBytes))
}

// 创建客户端的公共函数（每个测试独立创建，避免状态污染）
func newTestClient(t *testing.T) *ethclient.Client {
	cli, err := client.NewEthClient()
	if err != nil {
		t.Fatalf("创建 ethclient 失败: %v", err)
	}
	return cli
}

// ==================== 独立测试函数 ====================

func TestChainID(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	chainID, err := cli.ChainID(context.Background())
	printJSON(t, "ChainID", map[string]string{"chainID": chainID.String()}, err)
}

func TestBlockNumber(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	blockNumber, err := cli.BlockNumber(context.Background())
	printJSON(t, "BlockNumber", blockNumber, err)
}

func TestBalanceAtLatest(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	balance, err := cli.BalanceAt(context.Background(), testAddress, nil)
	printJSON(t, "BalanceAt (最新区块)", balance.String(), err)
}

func TestBalanceAtSpecificBlock(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	blockNum := big.NewInt(30_000_000)
	balance, err := cli.BalanceAt(context.Background(), testAddress, blockNum)
	printJSON(t, "BalanceAt (区块 30000000)", balance.String(), err)
}

func TestNonceAt(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	nonce, err := cli.NonceAt(context.Background(), testAddress, nil)
	printJSON(t, "NonceAt (最新)", nonce, err)
}

func TestSuggestGasPrice(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	gasPrice, err := cli.SuggestGasPrice(context.Background())
	printJSON(t, "SuggestGasPrice", gasPrice.String(), err)
}

func TestBlockByNumber(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	block, err := cli.BlockByNumber(context.Background(), big.NewInt(31_000_000))
	if err != nil {
		printJSON(t, "BlockByNumber", nil, err)
		return
	}
	result := map[string]interface{}{
		"Number":   block.Number().Uint64(),
		"Hash":     block.Hash().Hex(),
		"TxCount":  len(block.Transactions()),
		"Time":     time.Unix(int64(block.Time()), 0).Format(time.RFC3339),
		"GasUsed":  block.GasUsed(),
		"GasLimit": block.GasLimit(),
	}
	printJSON(t, "BlockByNumber (31000000)", result, nil)
}

func TestTransactionByHash(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	// 请替换为测试网上真实存在的交易 hash
	txHash := common.HexToHash("0xYourRealTestnetTxHashHere")
	tx, isPending, err := cli.TransactionByHash(context.Background(), txHash)
	if err != nil {
		printJSON(t, "TransactionByHash", nil, err)
		return
	}
	result := map[string]interface{}{
		"Hash":      tx.Hash().Hex(),
		"Value":     tx.Value().String(),
		"Gas":       tx.Gas(),
		"GasPrice":  tx.GasPrice().String(),
		"To":        addrToString(tx.To()),
		"IsPending": isPending,
	}
	printJSON(t, "TransactionByHash", result, nil)
}

func TestCodeAt(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	code, err := cli.CodeAt(context.Background(), pancakeRouterTest, nil)
	result := map[string]interface{}{
		"CodeLength": len(code),
		"HasCode":    len(code) > 0,
	}
	printJSON(t, "CodeAt (PancakeRouter)", result, err)
}

func TestCallContract(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	// 示例：调用 PancakeRouter getAmountsOut (1 BNB -> WBNB)
	// data 可通过 abi.Pack 生成，这里手动示例
	callMsg := ethereum.CallMsg{
		To:   &pancakeRouterTest,
		Data: common.FromHex("0xd06ca61f0000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000ae13d989dac2f0debff460ac112a837c89baa7cd"),
	}
	output, err := cli.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		printJSON(t, "CallContract", nil, err)
		return
	}
	result := map[string]interface{}{
		"OutputHex": common.Bytes2Hex(output),
		"Length":    len(output),
	}
	printJSON(t, "CallContract (getAmountsOut 示例)", result, nil)
}

func TestPendingBalanceAt(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	balance, err := cli.PendingBalanceAt(context.Background(), testAddress)
	printJSON(t, "PendingBalanceAt", balance.String(), err)
}

func TestPendingNonceAt(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	nonce, err := cli.PendingNonceAt(context.Background(), testAddress)
	printJSON(t, "PendingNonceAt", nonce, err)
}

func TestHeaderByNumber(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	header, err := cli.HeaderByNumber(context.Background(), nil) // 最新
	if err != nil {
		printJSON(t, "HeaderByNumber", nil, err)
		return
	}
	result := map[string]interface{}{
		"Number":   header.Number.Uint64(),
		"Hash":     header.Hash().Hex(),
		"Time":     time.Unix(int64(header.Time), 0).Format(time.RFC3339),
		"GasLimit": header.GasLimit,
	}
	printJSON(t, "HeaderByNumber (最新)", result, nil)
}

func TestFilterLogs(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	var addresses []common.Address
	fourmemeManagers := []string{config.DefaultFourmemeManager, "0x7d6c5429A39B8414609e8D257BEDA23525884444"}
	// Fourmeme Manager 地址
	for _, mgr := range fourmemeManagers {
		addresses = append(addresses, common.HexToAddress(mgr))
	}

	topicsArr := []string{
		//TODO 多topic时 FilterLogs失效
		//"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"}

	var topics [][]common.Hash
	for _, t := range topicsArr {
		topics = append(topics, []common.Hash{common.HexToHash(t)})
	}

	fromBlock := big.NewInt(71710960)
	toBlock := big.NewInt(71710970)

	t.Logf("查询区块范围: %d ~ %d", fromBlock.Uint64(), toBlock.Uint64())

	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: addresses,
		Topics:    topics,
	}

	logs, err := cli.FilterLogs(context.Background(), query) // 最新
	if err != nil {
		printJSON(t, "HeaderByNumber", nil, err)
		return
	}

	result := map[string]interface{}{
		"BlockRange": fmt.Sprintf("%d ~ %d", fromBlock.Uint64(), toBlock.Uint64()),
		"Addresses":  len(addresses),
		"Topics":     len(topics),
		"LogsFound":  len(logs),
	}

	if len(logs) > 0 {
		var previewList []map[string]string
		for i, l := range logs {
			if i >= 10 {
				break
			}
			previewList = append(previewList, map[string]string{
				"TxHash":   l.TxHash.Hex(),
				"Address":  l.Address.Hex(),
				"Topic0":   l.Topics[0].Hex(),
				"BlockNum": fmt.Sprintf("%d", l.BlockNumber),
				"LogIndex": fmt.Sprintf("%d", l.Index),
			})
		}
		result["PreviewLogs"] = previewList
	}

	printJSON(t, "FilterLogs", result, nil)

	if len(logs) == 0 {
		t.Log("未找到日志，建议检查：1. 区块范围是否包含事件 2. 节点是否同步 3. Addresses/Topics 是否正确")
	}
}

// 辅助函数
func addrToString(addr *common.Address) string {
	if addr == nil {
		return "<nil>"
	}
	return addr.Hex()
}
