package live

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/palomachain/arbitrum-geth/common"
	"github.com/palomachain/arbitrum-geth/core/types"

	"github.com/palomachain/arbitrum-geth/core/tracing"
	"github.com/palomachain/arbitrum-geth/eth/tracers"
	"github.com/palomachain/arbitrum-geth/eth/tracers/native"
	"github.com/palomachain/arbitrum-geth/log"
)

type txGasDimensionLiveTraceLogger struct {
	Path               string `json:"path"`
	GasDimensionTracer *tracers.Tracer
}

func init() {
	tracers.LiveDirectory.Register("txGasDimensionLogger", newTxGasDimensionLiveTraceLogger)
}

type txGasDimensionLiveTraceLoggerConfig struct {
	Path string `json:"path"` // Path to directory for output
}

func newTxGasDimensionLiveTraceLogger(cfg json.RawMessage) (*tracing.Hooks, error) {
	var config txGasDimensionLiveTraceLoggerConfig
	if err := json.Unmarshal(cfg, &config); err != nil {
		return nil, err
	}

	if config.Path == "" {
		return nil, fmt.Errorf("gas dimension live tracer path for output is required: %v", config)
	}

	// be sure path exists
	os.MkdirAll(config.Path, 0755)

	gasDimensionTracer, err := native.NewTxGasDimensionLogger(nil, nil, nil)
	if err != nil {
		return nil, err
	}

	t := &txGasDimensionLiveTraceLogger{
		Path:               config.Path,
		GasDimensionTracer: gasDimensionTracer,
	}

	return &tracing.Hooks{
		OnOpcode:     t.OnOpcode,
		OnTxStart:    t.OnTxStart,
		OnFault:      t.OnFault,
		OnTxEnd:      t.OnTxEnd,
		OnBlockStart: t.OnBlockStart,
		OnBlockEnd:   t.OnBlockEnd,
	}, nil
}

func (t *txGasDimensionLiveTraceLogger) OnTxStart(
	vm *tracing.VMContext,
	tx *types.Transaction,
	from common.Address,
) {
	t.GasDimensionTracer.OnTxStart(vm, tx, from)
}

func (t *txGasDimensionLiveTraceLogger) OnOpcode(
	pc uint64,
	op byte,
	gas, cost uint64,
	scope tracing.OpContext,
	rData []byte,
	depth int,
	err error,
) {
	t.GasDimensionTracer.OnOpcode(pc, op, gas, cost, scope, rData, depth, err)
}

func (t *txGasDimensionLiveTraceLogger) OnFault(
	pc uint64,
	op byte,
	gas, cost uint64,
	scope tracing.OpContext,
	depth int,
	err error,
) {
	t.GasDimensionTracer.OnFault(pc, op, gas, cost, scope, depth, err)
}

func (t *txGasDimensionLiveTraceLogger) OnTxEnd(
	receipt *types.Receipt,
	err error,
) {
	// first call the native tracer's OnTxEnd
	t.GasDimensionTracer.OnTxEnd(receipt, err)

	// then get the json from the native tracer
	executionResultJsonBytes, errGettingResult := t.GasDimensionTracer.GetResult()
	if errGettingResult != nil {
		log.Error("Failed to get result", "error", errGettingResult)
		return
	}

	blockNumber := receipt.BlockNumber.String()
	txHashString := receipt.TxHash.Hex()

	// Create the filename
	filename := fmt.Sprintf("%s_%s.json", blockNumber, txHashString)
	filepath := filepath.Join(t.Path, filename)

	// Ensure the directory exists
	if err := os.MkdirAll(t.Path, 0755); err != nil {
		log.Error("Failed to create directory", "path", t.Path, "error", err)
		return
	}

	// Write the file
	if err := os.WriteFile(filepath, executionResultJsonBytes, 0644); err != nil {
		log.Error("Failed to write file", "path", filepath, "error", err)
		return
	}
}

func (t *txGasDimensionLiveTraceLogger) OnBlockStart(ev tracing.BlockEvent) {
}

func (t *txGasDimensionLiveTraceLogger) OnBlockEnd(err error) {
}
