package v2

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"go_fourmeme/config"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// BuyTokenViaManagerV2 只发起V2买入交易 (无计算/后处理)
func BuyTokenViaManagerV2(
	tokenAddr string,
	amountInWei *big.Int,
	minAmountOut *big.Int,
	privateKey *ecdsa.PrivateKey) (string, error) {
	ethClient := manager.GetEthClient()

	managerABI := utils.GetABI("TokenManager2")
	if managerABI == nil {
		return "", fmt.Errorf("TokenManager2 ABI 未加载")
	}

	input, err := managerABI.Pack("buyTokenAMAP",
		big.NewInt(0),
		common.HexToAddress(tokenAddr),
		common.HexToAddress(config.BSCChain.WalletAddress),
		amountInWei,
		minAmountOut,
	)
	if err != nil {
		return "", fmt.Errorf("Pack buyTokenAMAP 失败: %v1", err)
	}

	nonce, nonceErr := ethClient.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	if nonceErr != nil {
		return "", fmt.Errorf("获取 nonce 失败: %v1", nonceErr)
	}
	log.LogInfo("当前钱包 nonce: %d", nonce)

	gasPrice, gasErr := ethClient.SuggestGasPrice(context.Background())
	if gasErr != nil {
		gasPrice = big.NewInt(1000000000)
	} else {
		gasPrice.Mul(gasPrice, big.NewInt(11)).Div(gasPrice, big.NewInt(10))
	}

	managerAddr := common.HexToAddress(config.TokenManager2)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &managerAddr,
		Value:    amountInWei,
		Gas:      1200000,
		GasPrice: gasPrice,
		Data:     input,
	})

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(config.BSCChain.ChainID), privateKey)
	if err != nil {
		return "", err
	}

	err = ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	txHash := signedTx.Hash().Hex()
	log.LogInfo("V2 一级买入提交成功 Tx: %s Token: %s", txHash, tokenAddr[:10])

	return txHash, nil
}

// SellTokenViaManagerV2 只发起V2卖出交易 (无计算/后处理)
func SellTokenViaManagerV2(
	tokenAddr string,
	sellAmount *big.Int,
	minFunds *big.Int,
	privateKey *ecdsa.PrivateKey) (string, error) {
	ethClient := manager.GetEthClient()

	managerABI := utils.GetABI("TokenManager2")
	if managerABI == nil {
		return "", fmt.Errorf("TokenManager2 ABI 未加载")
	}

	input, err := managerABI.Pack("sellToken",
		big.NewInt(0),
		common.HexToAddress(tokenAddr),
		sellAmount,
		minFunds,
		big.NewInt(0),
		common.HexToAddress(config.BSCChain.WalletAddress),
	)
	if err != nil {
		return "", fmt.Errorf("Pack sellToken 失败: %v1", err)
	}

	nonce, nonceErr := ethClient.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	if nonceErr != nil {
		return "", fmt.Errorf("获取 nonce 失败: %v1", nonceErr)
	}
	log.LogInfo("当前钱包 nonce: %d", nonce)

	gasPrice, gasErr := ethClient.SuggestGasPrice(context.Background())
	if gasErr != nil {
		gasPrice = big.NewInt(1000000000)
	} else {
		gasPrice.Mul(gasPrice, big.NewInt(11)).Div(gasPrice, big.NewInt(10))
	}

	managerAddr := common.HexToAddress(config.TokenManager2)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &managerAddr,
		Value:    big.NewInt(0),
		Gas:      1200000,
		GasPrice: gasPrice,
		Data:     input,
	})

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(config.BSCChain.ChainID), privateKey)
	if err != nil {
		return "", err
	}

	err = ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	txHash := signedTx.Hash().Hex()
	log.LogInfo("V2 一级卖出提交成功 Tx: %s Token: %s", txHash, tokenAddr[:10])

	return txHash, nil
}
