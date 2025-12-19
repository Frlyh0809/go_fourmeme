// trade/trade_utils.go
package trade

import (
	"context"
	"fmt"
	configentity "go_fourmeme/entity/config"
	"math/big"
	"strings"
	"time"

	"go_fourmeme/config"
	"go_fourmeme/entity"

	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	//"go_fourmeme/entity/po"
	//"go_fourmeme/log"
	"go_fourmeme/manager"
	"go_fourmeme/utils"
)

// ToWei BNB Float → Wei Int
func ToWei(amountBNB *big.Float) *big.Int {
	if amountBNB == nil {
		return big.NewInt(0)
	}
	weiPerBNB := new(big.Float).SetInt64(1e18)
	amountWeiFloat := new(big.Float).Mul(amountBNB, weiPerBNB)
	amountWei := new(big.Int)
	amountWeiFloat.Int(amountWei)
	return amountWei
}

// GetPrivateKey 从 config 加载 ECDSA 私钥
func GetPrivateKey() (*ecdsa.PrivateKey, error) {
	pkHex := strings.TrimPrefix(config.BSCChain.PrivateKey, "0x")
	pkBytes, err := hexutil.Decode("0x" + pkHex)
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSA(pkBytes)
}

// WaitForReceipt 等待交易收据
func WaitForReceipt(client *ethclient.Client, hash common.Hash) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for {
		receipt, err := client.TransactionReceipt(ctx, hash)
		if err == nil {
			return receipt, nil
		}
		if err != ethereum.NotFound {
			return nil, err
		}
		time.Sleep(3 * time.Second)
	}
}

// TODO bnb也提取出来
// ExtractTokenOutFromReceipt 从收据提取实际 token out (Manager → Wallet Transfer)
func ExtractTokenOutFromReceipt(receipt *types.Receipt, tokenAddr string) *big.Int {
	transferTopic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	managerAddr := common.HexToAddress(config.TokenManager2)
	walletAddr := common.HexToAddress(config.BSCChain.WalletAddress)

	for _, l := range receipt.Logs {
		if len(l.Topics) >= 3 && l.Topics[0] == transferTopic && l.Address.Hex() == tokenAddr {
			from := common.BytesToAddress(l.Topics[1].Bytes())
			to := common.BytesToAddress(l.Topics[2].Bytes())
			if from == managerAddr && to == walletAddr {
				return new(big.Int).SetBytes(l.Data)
			}
		}
	}
	return nil
}

// AddPositionFromReceipt 添加持仓 (从收据)
func AddPositionFromReceipt(tokenAddr, txHash string, target *configentity.MonitorTarget, tokenOut *big.Int) {
	buyBNBFloat := target.BuyAmountBNB
	buyPriceAvg := new(big.Float).Quo(buyBNBFloat, new(big.Float).SetInt(tokenOut))

	pos := &entity.Position{
		TokenAddr:        tokenAddr,
		BuyTxHash:        txHash,
		BuyAmountBNB:     buyBNBFloat,
		BuyTokenAmount:   tokenOut,
		BuyPriceAvg:      buyPriceAvg,
		BuyTime:          time.Now(),
		TargetProfitMult: target.TakeProfitMultiple,
		TargetLossMult:   target.StopLossMultiple,
		Sold:             false,
	}
	manager.AddPosition(pos)
}

// AddPositionFromEstimate 添加持仓 (估算 fallback)
func AddPositionFromEstimate(tokenAddr, txHash string, target *configentity.MonitorTarget, estimatedOut *big.Int) {
	AddPositionFromReceipt(tokenAddr, txHash, target, estimatedOut)
}

// calcSlippageMinOut 滑点计算 minOut (使用 Helper3 calcTokenOut 预估)
func calcSlippageMinOut(client *ethclient.Client, tokenAddr string, amountInWei *big.Int, slippage float64) (*big.Int, error) {
	helperABI := utils.GetABI("TokenManagerHelper")
	if helperABI == nil {
		return new(big.Int).Mul(amountInWei, big.NewInt(int64(100-int(slippage*100)))), fmt.Errorf("Helper ABI 未加载")
	}

	input, err := helperABI.Pack("calcTokenOut", common.HexToAddress(tokenAddr), amountInWei)
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
	err = helperABI.UnpackIntoInterface(&expectedOut, "calcTokenOut", output)
	if err != nil {
		return nil, err
	}

	minOut := new(big.Int).Mul(expectedOut, big.NewInt(int64(100-int(slippage*100))))
	minOut.Div(minOut, big.NewInt(100))
	return minOut, nil
}
