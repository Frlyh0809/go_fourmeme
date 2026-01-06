// parser/fourmeme/parser_interface.go
package fourmeme

import (
	"fmt"
	"go_fourmeme/config"
	"math/big"
	"strconv"
	"time"

	"go_fourmeme/client"
	"go_fourmeme/database"
	"go_fourmeme/entity/po"
	"go_fourmeme/log"
	"go_fourmeme/utils"

	"github.com/ethereum/go-ethereum/common"
)

func saveTradeRecord(
	blockNum uint64,
	txHash string,
	txIndex uint,
	send, receive, protocol, protocolAddress, tokenAddr, paymentSymbol string,
	tokenAmount, bnbAmount *big.Int,
	recordType int,
) {
	bnbPrice := client.GetBNBPriceUSDT()

	tokenAmountFloat := utils.Div10Pow(tokenAmount, big.NewInt(18))
	bnbAmountFloat := utils.Div10Pow(bnbAmount, big.NewInt(18))

	price := new(big.Float).Quo(bnbAmountFloat, tokenAmountFloat)
	volume := new(big.Float).Set(bnbAmountFloat)

	if paymentSymbol == "USD1" || paymentSymbol == "BUSDT" {
		price = new(big.Float).Mul(price, big.NewFloat(bnbPrice))
		volume = new(big.Float).Mul(volume, big.NewFloat(bnbPrice))
	}

	txRecord := &po.Transaction{
		TxUniqueSeq:     utils.CalcTxUniqueSeq(blockNum, txIndex, 0), // logIndex 暂用 0，实际可传
		BlockNumber:     strconv.FormatUint(blockNum, 10),
		TxIndex:         strconv.FormatUint(uint64(txIndex), 10),
		TxHash:          txHash,
		Send:            send,
		Receive:         receive,
		Type:            recordType,
		Protocol:        protocol,
		ProtocolAddress: protocolAddress,
		TokenAddress:    tokenAddr,
		PaymentToken:    paymentSymbol,
		TokenAmount:     utils.BigFloatToString(tokenAmountFloat),
		PaymentAmount:   utils.BigFloatToString(bnbAmountFloat),
		Price:           utils.BigFloatToString(price),
		Volume:          utils.BigFloatToString(volume),
		CreatedAt:       time.Now(),
	}

	if err := database.SaveTransaction(txRecord); err != nil {
		log.LogError("保存交易记录失败: %v", err)
	}
}

func getPaymentToken(tokenAddress string) (string, string, error) {
	tokenStatus, err := client.GetTokenStatus(tokenAddress)
	if err != nil {
		return "", "", fmt.Errorf("GetTokenStatus err for token %s err:%v", tokenAddress, err)
	}

	paymentToken := ""
	paymentTokenSymbol := ""
	if tokenStatus.Quote != config.AddrZero {
		paymentToken = tokenStatus.Quote.String()
		if tokenStatus.Quote == config.AddrUSDT {
			paymentTokenSymbol = "BUSDT"
		} else if tokenStatus.Quote == config.AddrUSD1 {
			paymentTokenSymbol = "USD1"
		} else if tokenStatus.Quote == config.WBNB {
			paymentTokenSymbol = "WBNB"
		}
	} else {
		paymentToken = config.ZeroAddress
		paymentTokenSymbol = "BNB"

	}
	return paymentToken, paymentTokenSymbol, nil
}

func isFourmemeManager(addr common.Address) bool {
	return addr == config.AddrTokenManager2 || addr == config.AddrTokenManagerHelper3 || addr == config.AddrTokenManager1
}
