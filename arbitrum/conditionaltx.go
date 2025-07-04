package arbitrum

import (
	"context"
	"errors"

	"github.com/palomachain/arbitrum-geth/arbitrum_types"
	"github.com/palomachain/arbitrum-geth/common"
	"github.com/palomachain/arbitrum-geth/common/hexutil"
	"github.com/palomachain/arbitrum-geth/core/types"
	"github.com/palomachain/arbitrum-geth/crypto"
	"github.com/palomachain/arbitrum-geth/internal/ethapi"
	"github.com/palomachain/arbitrum-geth/log"
	"github.com/palomachain/arbitrum-geth/rpc"
)

type ArbTransactionAPI struct {
	b *APIBackend
}

func NewArbTransactionAPI(b *APIBackend) *ArbTransactionAPI {
	return &ArbTransactionAPI{b}
}

func (s *ArbTransactionAPI) SendRawTransactionConditional(ctx context.Context, input hexutil.Bytes, options *arbitrum_types.ConditionalOptions) (common.Hash, error) {
	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(input); err != nil {
		return common.Hash{}, err
	}
	return SubmitConditionalTransaction(ctx, s.b, tx, options)
}

func SubmitConditionalTransaction(ctx context.Context, b *APIBackend, tx *types.Transaction, options *arbitrum_types.ConditionalOptions) (common.Hash, error) {
	// If the transaction fee cap is already specified, ensure the
	// fee of the given transaction is _reasonable_.
	if err := ethapi.CheckTxFee(tx.GasPrice(), tx.Gas(), b.RPCTxFeeCap()); err != nil {
		return common.Hash{}, err
	}
	if !b.UnprotectedAllowed() && !tx.Protected() {
		// Ensure only eip155 signed transactions are submitted if EIP155Required is set.
		return common.Hash{}, errors.New("only replay-protected (EIP-155) transactions allowed over RPC")
	}
	if err := b.SendConditionalTx(ctx, tx, options); err != nil {
		return common.Hash{}, err
	}
	// Print a log with full tx details for manual investigations and interventions
	arbosVersion := types.DeserializeHeaderExtraInformation(b.CurrentBlock()).ArbOSFormatVersion
	signer := types.MakeSigner(b.ChainConfig(), b.CurrentBlock().Number, b.CurrentBlock().Time, arbosVersion)
	from, err := types.Sender(signer, tx)
	if err != nil {
		return common.Hash{}, err
	}

	if tx.To() == nil {
		addr := crypto.CreateAddress(from, tx.Nonce())
		log.Info("Submitted contract creation", "hash", tx.Hash().Hex(), "from", from, "nonce", tx.Nonce(), "contract", addr.Hex(), "value", tx.Value())
	} else {
		log.Info("Submitted transaction", "hash", tx.Hash().Hex(), "from", from, "nonce", tx.Nonce(), "recipient", tx.To(), "value", tx.Value())
	}
	return tx.Hash(), nil
}

func SendConditionalTransactionRPC(ctx context.Context, rpc *rpc.Client, tx *types.Transaction, options *arbitrum_types.ConditionalOptions) error {
	data, err := tx.MarshalBinary()
	if err != nil {
		return err
	}
	return rpc.CallContext(ctx, nil, "eth_sendRawTransactionConditional", hexutil.Encode(data), options)
}
