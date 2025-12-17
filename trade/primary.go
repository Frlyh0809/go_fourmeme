// trade/primary.go
package trade

import (
	"context"
	"fmt"
	"go_fourmeme/database"
	"go_fourmeme/entity/po"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"go_fourmeme/config"
	configentity "go_fourmeme/entity/config"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/utils"
)

// BuyTokenViaManager 一级市场买入 (通过 Helper3 buyWithEth)
func BuyTokenViaManager(target *configentity.MonitorTarget, tokenAddr string) (string, error) {
	client := manager.GetEthClient()

	amountInWei := ToWei(target.BuyAmountBNB)
	if amountInWei.Sign() == 0 {
		return "", fmt.Errorf("买入金额为0")
	}

	// 滑点预估 (使用 calcTokenOut)
	minAmountOut, err := calcSlippageMinOut(client, tokenAddr, amountInWei, target.SlippageTolerance)
	if err != nil {
		return "", fmt.Errorf("滑点计算失败: %v", err)
	}

	helperABI := utils.GetABI("TokenManagerHelper")
	if helperABI == nil {
		return "", fmt.Errorf("Helper ABI 未加载")
	}

	// Pack buyWithEth: origin=0, token, to=wallet, funds=amountInWei, minAmount
	input, err := helperABI.Pack("buyWithEth",
		big.NewInt(0), // origin (requestId, 0 为默认)
		common.HexToAddress(tokenAddr),
		common.HexToAddress(config.BSCChain.WalletAddress), // to (接收者)
		amountInWei,
		minAmountOut,
	)
	if err != nil {
		return "", fmt.Errorf("Pack buyWithEth 失败: %v", err)
	}

	// 构建交易
	nonce, _ := client.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	helperAddr := common.HexToAddress(config.TokenManagerHelper3)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &helperAddr,
		Value:    amountInWei, // payable BNB
		Gas:      800000,
		GasPrice: gasPrice,
		Data:     input,
	})

	privateKey, err := getPrivateKey()
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(config.BSCChain.ChainID), privateKey)
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		database.SaveTxRecord(&po.TransactionRecord{
			Type:      "buy_primary",
			TokenAddr: tokenAddr,
			AmountIn:  amountInWei,
			Status:    "failed",
			ErrorMsg:  err.Error(),
			Timestamp: time.Now(),
		})
		return "", err
	}

	txHash := signedTx.Hash().Hex()
	log.LogInfo("一级买入提交成功 Tx: %s Token: %s Amount: %.6f BNB", txHash, tokenAddr[:10], target.BuyAmountBNB.Text('f', 6))

	// 等待收据 + 添加持仓
	receipt, err := waitForReceipt(client, signedTx.Hash())
	if err != nil {
		log.LogWarn("收据获取失败，使用估算: %v", err)
		addPositionFromEstimate(tokenAddr, txHash, target, minAmountOut)
		return txHash, nil
	}

	actualOut := extractTokenOutFromReceipt(receipt, tokenAddr)
	if actualOut == nil {
		actualOut = minAmountOut
	}
	addPositionFromReceipt(tokenAddr, txHash, target, actualOut)

	return txHash, nil
}

// SellTokenViaManager 一级市场卖出 (通过 Helper3 sellForEth)
func SellTokenViaManager(tokenAddr string, amountIn *big.Int, slippage float64) (string, error) {
	client := manager.GetEthClient()

	if amountIn.Sign() == 0 {
		return "", fmt.Errorf("卖出金额为0")
	}

	// 滑点预估 (calcEthOut)
	minFunds, err := calcSlippageMinOutForSell(client, tokenAddr, amountIn, slippage)
	if err != nil {
		return "", err
	}

	helperABI := utils.GetABI("TokenManagerHelper")
	input, err := helperABI.Pack("sellForEth",
		big.NewInt(0), // origin
		common.HexToAddress(tokenAddr),
		amountIn,
		minFunds,
	)
	if err != nil {
		return "", err
	}

	// 构建交易 (非 payable)
	nonce, _ := client.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	helperAddr := common.HexToAddress(config.TokenManagerHelper3)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &helperAddr,
		Value:    big.NewInt(0),
		Gas:      800000,
		GasPrice: gasPrice,
		Data:     input,
	})

	privateKey, err := getPrivateKey()
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(config.BSCChain.ChainID), privateKey)
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	txHash := signedTx.Hash().Hex()
	log.LogInfo("一级卖出提交成功 Tx: %s Token: %s Amount: %s", txHash, tokenAddr[:10], amountIn.String())

	return txHash, nil
}

// calcSlippageMinOutForSell 卖出滑点预估 (calcEthOut)
func calcSlippageMinOutForSell(client *ethclient.Client, tokenAddr string, amountIn *big.Int, slippage float64) (*big.Int, error) {
	helperABI := utils.GetABI("TokenManagerHelper")
	if helperABI == nil {
		return new(big.Int).Mul(amountIn, big.NewInt(int64(100-int(slippage*100)))), fmt.Errorf("Helper ABI 未加载")
	}

	input, err := helperABI.Pack("calcEthOut", common.HexToAddress(tokenAddr), amountIn)
	if err != nil {
		return nil, err
	}
	tokenManagerHelper3 := common.HexToAddress(config.TokenManagerHelper3)
	callMsg := ethereum.CallMsg{
		To:   &tokenManagerHelper3,
		Data: input,
	}

	output, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nil, err
	}

	var expectedOut *big.Int
	err = helperABI.UnpackIntoInterface(&expectedOut, "calcEthOut", output)
	if err != nil {
		return nil, err
	}

	minOut := new(big.Int).Mul(expectedOut, big.NewInt(int64(100-int(slippage*100))))
	minOut.Div(minOut, big.NewInt(100))
	return minOut, nil
}
