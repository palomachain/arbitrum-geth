package eth

import (
	"context"

	"github.com/palomachain/arbitrum-geth/core"
	"github.com/palomachain/arbitrum-geth/core/state"
	"github.com/palomachain/arbitrum-geth/core/types"
	"github.com/palomachain/arbitrum-geth/core/vm"
	"github.com/palomachain/arbitrum-geth/eth/tracers"
	"github.com/palomachain/arbitrum-geth/ethdb"
)

func NewArbEthereum(
	blockchain *core.BlockChain,
	chainDb ethdb.Database,
) *Ethereum {
	return &Ethereum{
		blockchain: blockchain,
		chainDb:    chainDb,
	}
}

func (eth *Ethereum) StateAtTransaction(ctx context.Context, block *types.Block, txIndex int, reexec uint64) (*types.Transaction, vm.BlockContext, *state.StateDB, tracers.StateReleaseFunc, error) {
	return eth.stateAtTransaction(ctx, block, txIndex, reexec)
}
