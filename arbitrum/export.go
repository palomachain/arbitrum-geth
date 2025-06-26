package arbitrum

import (
	"context"

	"github.com/palomachain/arbitrum-geth/common/hexutil"
	"github.com/palomachain/arbitrum-geth/core"
	"github.com/palomachain/arbitrum-geth/internal/ethapi"
	"github.com/palomachain/arbitrum-geth/internal/ethapi/override"
	"github.com/palomachain/arbitrum-geth/rpc"
)

type TransactionArgs = ethapi.TransactionArgs

func EstimateGas(ctx context.Context, b ethapi.Backend, args TransactionArgs, blockNrOrHash rpc.BlockNumberOrHash, overrides *override.StateOverride, blockOverrides *override.BlockOverrides, gasCap uint64) (hexutil.Uint64, error) {
	return ethapi.DoEstimateGas(ctx, b, args, blockNrOrHash, overrides, blockOverrides, gasCap)
}

func NewRevertReason(result *core.ExecutionResult) error {
	return ethapi.NewRevertError(result)
}
