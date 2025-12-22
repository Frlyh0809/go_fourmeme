package v3

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
)

// client/fourmeme_client.go (顶部添加)
type TryBuyResult struct {
	TokenManager    common.Address
	Quote           common.Address
	EstimatedAmount *big.Int // 预计得到的 token 数量
	EstimatedCost   *big.Int // 预计花费 BNB
	EstimatedFee    *big.Int // 预计手续费
	AmountMsgValue  *big.Int // 交易 value
	AmountApproval  *big.Int // 需要 approve 的数量 (USD1 时)
	AmountFunds     *big.Int // 实际投入资金
}

type TrySellResult struct {
	TokenManager common.Address
	Quote        common.Address
	Funds        *big.Int // 预计得到的 BNB
	Fee          *big.Int // 手续费
}

// BuyTokenViaManagerV3 只发起 V3 (Helper3) 买入交易 (无计算/后处理)
// 使用 buyWithEth(origin=0, token, to, funds, minAmount)
func BuyTokenViaManagerV3(
	tokenAddr string,
	amountInWei *big.Int, // funds，通过 Value 传递
	minAmountOut *big.Int, // 滑点保护
	privateKey *ecdsa.PrivateKey,
) (string, error) {
	ethClient := manager.GetEthClient()

	helperABI := utils.GetABI("TokenManagerHelper") // 确保 utils 中加载了 Helper3 的 ABI
	if helperABI == nil {
		return "", fmt.Errorf("TokenManagerHelper ABI 未加载")
	}

	// Pack buyWithEth: origin=0, token, to, funds, minAmount
	input, err := helperABI.Pack("buyWithEth",
		big.NewInt(0),                  // origin (requestId, 一般为0)
		common.HexToAddress(tokenAddr), // token
		common.HexToAddress(config.BSCChain.WalletAddress), // to (接收者)
		amountInWei,  // funds
		minAmountOut, // minAmount
	)
	if err != nil {
		return "", fmt.Errorf("Pack buyWithEth 失败: %v", err)
	}

	nonce, nonceErr := ethClient.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	if nonceErr != nil {
		return "", fmt.Errorf("获取 nonce 失败: %v", nonceErr)
	}
	log.LogInfo("V3 买入 当前钱包 nonce: %d", nonce)

	gasPrice, gasErr := ethClient.SuggestGasPrice(context.Background())
	if gasErr != nil {
		gasPrice = big.NewInt(1000000000) // 1 gwei 默认
	} else {
		gasPrice.Mul(gasPrice, big.NewInt(11)).Div(gasPrice, big.NewInt(10)) // +10%
	}

	helperAddr := common.HexToAddress(config.TokenManagerHelper3) // 确保 config 中有这个字段

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &helperAddr,
		Value:    amountInWei, // payable BNB
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
	log.LogInfo("V3 一级买入提交成功 Tx: %s Token: %s", txHash, tokenAddr[:10])

	return txHash, nil
}

// SellTokenViaManagerV3 只发起 V3 (Helper3) 卖出交易
func SellTokenViaManagerV3(
	tokenAddr string,
	sellAmount *big.Int,
	minFunds *big.Int,
	privateKey *ecdsa.PrivateKey,
) (string, error) {
	ethClient := manager.GetEthClient()

	helperABI := utils.GetABI("TokenManagerHelper")
	if helperABI == nil {
		return "", fmt.Errorf("TokenManagerHelper ABI 未加载")
	}

	// Pack sellForEth (7 参数版本)
	input, err := helperABI.Pack("sellForEth",
		big.NewInt(0),                  // origin
		common.HexToAddress(tokenAddr), // token
		common.HexToAddress(config.BSCChain.WalletAddress), // from (你的钱包)
		sellAmount,     // amount
		minFunds,       // minFunds (滑点保护)
		big.NewInt(50), // feeRate = 0 (使用默认费率)
		//common.Address{}, // feeRecipient = 零地址 (官方常用)
		common.HexToAddress("0x69aaaff7419fd2ce58079f05ff06b2ff66d446d2"), // feeRecipient = 零地址 (官方常用)
	)
	if err != nil {
		return "", fmt.Errorf("Pack sellForEth 失败: %v", err)
	}

	nonce, nonceErr := ethClient.PendingNonceAt(context.Background(), common.HexToAddress(config.BSCChain.WalletAddress))
	if nonceErr != nil {
		return "", fmt.Errorf("获取 nonce 失败: %v", nonceErr)
	}
	log.LogInfo("V3 卖出 当前钱包 nonce: %d", nonce)

	gasPrice, gasErr := ethClient.SuggestGasPrice(context.Background())
	if gasErr != nil {
		gasPrice = big.NewInt(1000000000) // 1 gwei
	} else {
		// 加 10% gas price
		gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(11))
		gasPrice.Div(gasPrice, big.NewInt(10))
	}

	helperAddr := common.HexToAddress(config.TokenManagerHelper3)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &helperAddr,
		Value:    big.NewInt(0), // sell 不需要发送 BNB
		Gas:      1200000,
		GasPrice: gasPrice,
		Data:     input,
	})

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(config.BSCChain.ChainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("签名失败: %v", err)
	}

	err = ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("发送交易失败: %v", err)
	}

	txHash := signedTx.Hash().Hex()
	log.LogInfo("V3 一级卖出提交成功 Tx: %s Token: %s Amount: %s MinBNB: %s",
		txHash, tokenAddr[:10], sellAmount.String(), minFunds.String())

	return txHash, nil
}

// TryBuy 模拟买入，返回预估信息（view 函数，安全不消耗 gas）
func TryBuy(tokenAddr string, amountInWei *big.Int, wantTokenAmounts *big.Int) (TryBuyResult, error) {
	var result TryBuyResult

	ethClient := manager.GetEthClient()
	if ethClient == nil {
		return result, fmt.Errorf("ethclient 未初始化")
	}

	helperABI := utils.GetABI("TokenManagerHelper")
	if helperABI == nil {
		return result, fmt.Errorf("TokenManagerHelper ABI 未加载")
	}

	helperAddr := common.HexToAddress(config.TokenManagerHelper3)

	// Pack tryBuy(token, amount)
	//input, err := helperABI.Pack("tryBuy", common.HexToAddress(tokenAddr), wantTokenAmounts, amountInWei)
	//input, err := helperABI.Pack("tryBuy", common.HexToAddress(tokenAddr), amountInWei, big.NewInt(0))
	input, err := helperABI.Pack("tryBuy", common.HexToAddress(tokenAddr), big.NewInt(0), amountInWei)
	if err != nil {
		return result, fmt.Errorf("pack tryBuy 失败: %w", err)
	}

	// view 调用
	output, err := ethClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &helperAddr,
		Data: input,
	}, nil)
	if err != nil {
		return result, fmt.Errorf("tryBuy 调用失败: %w", err)
	}
	if len(output) == 0 {
		log.LogWarn("tryBuy 返回空数据 (token 可能不可交易或已饱和): %s", tokenAddr[:10])
		return result, fmt.Errorf("tryBuy 返回空值，token 不可买入")
	}

	// 解包 8 个返回值
	var out []interface{}
	err = helperABI.UnpackIntoInterface(&out, "tryBuy", output)
	if err != nil || len(out) < 8 {
		return result, fmt.Errorf("tryBuy 解包失败: %w", err)
	}

	result.TokenManager = out[0].(common.Address)
	result.Quote = out[1].(common.Address)
	result.EstimatedAmount = out[2].(*big.Int)
	result.EstimatedCost = out[3].(*big.Int)
	result.EstimatedFee = out[4].(*big.Int)
	result.AmountMsgValue = out[5].(*big.Int)
	result.AmountApproval = out[6].(*big.Int)
	result.AmountFunds = out[7].(*big.Int)
	// 检查是否全零
	if result.EstimatedAmount.Sign() == 0 {
		log.LogWarn("tryBuy 预计 tokenOut 为 0，跳过买入: %s", tokenAddr[:10])
		return result, fmt.Errorf("预计输出 0 token")
	}

	log.LogInfo("TryBuy 成功 Token: %s | EstAmount: %s | Cost: %s wei | Fee: %s | Value: %s",
		tokenAddr[:10],
		utils.BigIntToString(result.EstimatedAmount),
		utils.BigIntToString(result.EstimatedCost),
		utils.BigIntToString(result.EstimatedFee),
		utils.BigIntToString(result.AmountMsgValue),
	)

	return result, nil
}

// TrySell 模拟卖出，返回预估信息
func TrySell(tokenAddr string, tokenAmount *big.Int) (TrySellResult, error) {
	var result TrySellResult

	client := manager.GetEthClient()
	if client == nil {
		return result, fmt.Errorf("ethclient 未初始化")
	}

	helperABI := utils.GetABI("TokenManagerHelper")
	if helperABI == nil {
		return result, fmt.Errorf("TokenManagerHelper ABI 未加载")
	}

	helperAddr := common.HexToAddress(config.TokenManagerHelper3)

	input, err := helperABI.Pack("trySell", common.HexToAddress(tokenAddr), tokenAmount)
	if err != nil {
		return result, fmt.Errorf("pack trySell 失败: %w", err)
	}

	output, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &helperAddr,
		Data: input,
	}, nil)
	if err != nil {
		log.LogError("trySell 调用失败: %v", err)
		return result, fmt.Errorf("trySell 调用失败: %w", err)
	}

	if len(output) == 0 {
		log.LogWarn("trySell 返回空数据 (token 可能不可卖或已毕业): %s", tokenAddr[:10])
		return result, fmt.Errorf("trySell 返回空值，token 不可卖出")
	}

	var out []interface{}
	err = helperABI.UnpackIntoInterface(&out, "trySell", output)
	if err != nil {
		log.LogError("trySell 解包失败: %v", err)
		return result, fmt.Errorf("trySell 解包失败: %w", err)
	}
	if len(out) < 4 {
		return result, fmt.Errorf("trySell 返回字段不足，got %d want 4", len(out))
	}

	result.TokenManager = out[0].(common.Address)
	result.Quote = out[1].(common.Address)
	result.Funds = out[2].(*big.Int)
	result.Fee = out[3].(*big.Int)

	if result.Funds.Sign() == 0 {
		log.LogWarn("trySell 预计输出 0 BNB，跳过卖出: %s", tokenAddr[:10])
		return result, fmt.Errorf("预计输出 0 BNB")
	}

	log.LogInfo("TrySell 成功 Token: %s | EstFunds: %s wei | Fee: %s",
		tokenAddr[:10],
		utils.BigIntToString(result.Funds),
		utils.BigIntToString(result.Fee))

	return result, nil
}
