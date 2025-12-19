// trade/secondary.go
package trade

import (
	"context"
	"fmt"
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

// BuyTokenSecondary 二级市场买入 (PancakeSwap swapExactETHForTokens)
func BuyTokenSecondary(target *configentity.MonitorTarget, tokenAddr string) (string, error) {
	client := manager.GetEthClient()

	amountInWei := ToWei(target.BuyAmountBNB)
	if amountInWei.Sign() == 0 {
		return "", fmt.Errorf("买入金额为0")
	}

	// 滑点预估 (Pancake getAmountsOut)
	minAmountOut, err := calcSecondaryMinOut(client, amountInWei, tokenAddr, target.SlippageTolerance)
	if err != nil {
		return "", err
	}

	pancakeABI := utils.GetABI("PancakeRouter")
	path := []common.Address{config.WBNB, common.HexToAddress(tokenAddr)}
	deadline := big.NewInt(time.Now().Add(20 * time.Minute).Unix())

	input, err := pancakeABI.Pack("swapExactETHForTokens",
		minAmountOut,
		path,
		common.HexToAddress(config.BSCChain.WalletAddress),
		deadline,
	)
	if err != nil {
		return "", err
	}

	// 构建交易
	nonce, _ := client.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	gasPrice, _ := client.SuggestGasPrice(context.Background())

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &config.PancakeRouter,
		Value:    amountInWei, // payable BNB
		Gas:      300000,
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
	log.LogInfo("二级买入提交成功 Tx: %s Token: %s Amount: %s BNB", txHash, tokenAddr[:10], target.BuyAmountBNB.Text('f', 6))

	// 添加持仓 (二级无 Manager Transfer，估算 tokenOut)
	addPositionFromEstimate(tokenAddr, txHash, target, minAmountOut)

	return txHash, nil
}

// SellTokenSecondary 二级市场卖出 (swapExactTokensForETH)
func SellTokenSecondary(tokenAddr string, tokenAmount *big.Int, slippage float64) (string, error) {
	client := manager.GetEthClient()

	if tokenAmount.Sign() == 0 {
		return "", fmt.Errorf("卖出金额为0")
	}

	// 先 approve token to PancakeRouter (ERC20 approve)
	err := approveToken(client, tokenAddr, tokenAmount)
	if err != nil {
		return "", err
	}

	// 滑点预估
	minAmountOut, err := calcSecondaryMinOutForSell(client, tokenAmount, tokenAddr, slippage)
	if err != nil {
		return "", err
	}

	pancakeABI := utils.GetABI("PancakeRouter")
	path := []common.Address{common.HexToAddress(tokenAddr), config.WBNB}
	deadline := big.NewInt(time.Now().Add(20 * time.Minute).Unix())

	input, err := pancakeABI.Pack("swapExactTokensForETH",
		tokenAmount,
		minAmountOut,
		path,
		common.HexToAddress(config.BSCChain.WalletAddress),
		deadline,
	)
	if err != nil {
		return "", err
	}

	// 构建交易 (非 payable)
	nonce, _ := client.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	gasPrice, _ := client.SuggestGasPrice(context.Background())

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &config.PancakeRouter,
		Value:    big.NewInt(0),
		Gas:      300000,
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
	log.LogInfo("二级卖出提交成功 Tx: %s Token: %s Amount: %s", txHash, tokenAddr[:10], tokenAmount.String())

	return txHash, nil
}

// approveToken ERC20 approve to PancakeRouter
func approveToken(client *ethclient.Client, tokenAddr string, amount *big.Int) error {
	erc20ABI := utils.GetABI("ERC20")
	if erc20ABI == nil {
		return fmt.Errorf("ERC20 ABI 未加载")
	}

	input, err := erc20ABI.Pack("approve",
		config.PancakeRouter,
		amount,
	)
	if err != nil {
		return err
	}

	nonce, _ := client.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	gasPrice, _ := client.SuggestGasPrice(context.Background())
	tokenAddress := common.HexToAddress(tokenAddr)
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &tokenAddress,
		Value:    big.NewInt(0),
		Gas:      100000,
		GasPrice: gasPrice,
		Data:     input,
	})

	privateKey, err := getPrivateKey()
	if err != nil {
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(config.BSCChain.ChainID), privateKey)
	if err != nil {
		return err
	}

	return client.SendTransaction(context.Background(), signedTx)
}

// calcSecondaryMinOut 买入滑点预估 (getAmountsOut)
func calcSecondaryMinOut(client *ethclient.Client, amountIn *big.Int, tokenAddr string, slippage float64) (*big.Int, error) {
	pancakeABI := utils.GetABI("PancakeRouter")
	path := []common.Address{config.WBNB, common.HexToAddress(tokenAddr)}

	input, err := pancakeABI.Pack("getAmountsOut", amountIn, path)
	if err != nil {
		return nil, err
	}

	callMsg := ethereum.CallMsg{
		To:   &config.PancakeRouter,
		Data: input,
	}

	output, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nil, err
	}

	var amounts []*big.Int
	err = pancakeABI.UnpackIntoInterface(&amounts, "getAmountsOut", output)
	if err != nil || len(amounts) < 2 {
		return nil, err
	}

	expectedOut := amounts[len(amounts)-1]
	minOut := new(big.Int).Mul(expectedOut, big.NewInt(int64(100-int(slippage*100))))
	minOut.Div(minOut, big.NewInt(100))
	return minOut, nil
}

// calcSecondaryMinOutForSell 卖出滑点预估
func calcSecondaryMinOutForSell(client *ethclient.Client, amountIn *big.Int, tokenAddr string, slippage float64) (*big.Int, error) {
	// 类似 calcSecondaryMinOut，但 path 反向
	//path := []common.Address{common.HexToAddress(tokenAddr), config.WBNB}
	// ... 复制逻辑，返回 min BNB out
	return nil, nil // 实现类似买入
}
