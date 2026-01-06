// parser/handler.go
package parser

import (
	configentity "go_fourmeme/entity/config"
	_interface "go_fourmeme/parser/interface"
	"go_fourmeme/parser/native"

	"github.com/ethereum/go-ethereum/core/types"
)

func HandleEventV3(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget) {
	handlers := []_interface.Parser{
		//fourmeme.NewCreateHandler(),
		//fourmeme.NewBuyHandler(),
		//fourmeme.NewSellHandler(),
	}

	for _, handler := range handlers {
		handler.ParseLog(allLogs, receipt, target)

		native.NewTransferHandler().ParseLogMulti(allLogs, receipt, target)
	}
}
