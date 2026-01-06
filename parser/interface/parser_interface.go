// parser/parser_interface.go
package _interface

import (
	configentity "go_fourmeme/entity/config"

	"github.com/ethereum/go-ethereum/core/types"
)

type Parser interface {
	ParseLog(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget)
	ParseLogMulti(allLogs []types.Log, receipt *types.Receipt, target *configentity.MonitorTarget)
}
