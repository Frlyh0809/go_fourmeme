package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"go_fourmeme/config"
	"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/utils"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func Approve(
	tokenAddr string,
	to string,
	balance *big.Int,
	privateKey *ecdsa.PrivateKey) (string, error) {
	tokenABI := utils.GetABI("ERC20") // 标准ERC20 ABI，用于approve/allowance
	if tokenABI == nil {
		return "", fmt.Errorf("ERC20 ABI 未加载")
	}
	ethClient := manager.GetEthClient()

	publicKey := privateKey.PublicKey
	walletAddress := crypto.PubkeyToAddress(publicKey)

	tokenAddress := common.HexToAddress(tokenAddr)
	managerAddr := common.HexToAddress(to)

	allowanceInput, _ := tokenABI.Pack("allowance", walletAddress, managerAddr)
	callMsg := ethereum.CallMsg{
		From: walletAddress,
		To:   &tokenAddress,
		Data: allowanceInput,
	}
	result, err := ethClient.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return "", fmt.Errorf("查询allowance失败: %v1", err)
	}
	currentAllowance := new(big.Int)
	err = tokenABI.UnpackIntoInterface(&currentAllowance, "allowance", result)
	if err != nil {
		return "", err
	}

	if currentAllowance.Cmp(balance) < 0 {
		log.LogInfo("Allowance不足 (%s < %s)，执行Approve全额", currentAllowance.String(), balance.String())

		approveInput, err := tokenABI.Pack("approve", managerAddr, balance)
		if err != nil {
			return "", fmt.Errorf("pack approve 失败: %v1", err)
		}

		nonce, _ := ethClient.PendingNonceAt(context.Background(), walletAddress)
		gasPrice, _ := ethClient.SuggestGasPrice(context.Background())

		approveTx := types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			To:       &tokenAddress,
			Value:    big.NewInt(0),
			Gas:      100000,
			GasPrice: gasPrice,
			Data:     approveInput,
		})

		signedApproveTx, err := types.SignTx(approveTx, types.NewEIP155Signer(config.BSCChain.ChainID), privateKey)
		if err != nil {
			return "", err
		}

		err = ethClient.SendTransaction(context.Background(), signedApproveTx)
		if err != nil {
			return "", fmt.Errorf("approve交易发送失败: %v1", err)
		}
		log.LogInfo("Approve提交成功 Tx: %s", signedApproveTx.Hash().Hex())

		//// 等待approve确认
		//_, err = trade.WaitForReceipt(ethClient, signedApproveTx.Hash())
		//if err != nil {
		//	return signedApproveTx.Hash().Hex(), fmt.Errorf("等待Approve确认失败: %v1", err)
		//}
		//log.LogInfo("Approve已确认")
		// approve后nonce已+1，后续卖出会自动使用新nonce
		return signedApproveTx.Hash().Hex(), nil
	}
	return "", nil
}
