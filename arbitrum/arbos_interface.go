package arbitrum

import (
	"context"

	"github.com/palomachain/arbitrum-geth/arbitrum_types"
	"github.com/palomachain/arbitrum-geth/core"
	"github.com/palomachain/arbitrum-geth/core/types"
)

type ArbInterface interface {
	PublishTransaction(ctx context.Context, tx *types.Transaction, options *arbitrum_types.ConditionalOptions) error
	BlockChain() *core.BlockChain
	ArbNode() interface{}
}
