// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"context"

	"github.com/palomachain/arbitrum-geth/common"
	"github.com/palomachain/arbitrum-geth/core/state"
	"github.com/palomachain/arbitrum-geth/core/types"
	"github.com/palomachain/arbitrum-geth/core/vm"
	"github.com/palomachain/arbitrum-geth/params"
	"github.com/palomachain/arbitrum-geth/rpc"
)

// Installs an Arbitrum TxProcessor, enabling ArbOS for this state transition (see vm/evm_arbitrum.go)
var ReadyEVMForL2 func(evm *vm.EVM, msg *Message)

// Allows ArbOS to swap out or return early from an RPC message to support the NodeInterface virtual contract
var InterceptRPCMessage = func(
	msg *Message,
	ctx context.Context,
	statedb *state.StateDB,
	header *types.Header,
	backend NodeInterfaceBackendAPI,
	blockCtx *vm.BlockContext,
) (*Message, *ExecutionResult, error) {
	return msg, nil, nil
}

// Gets ArbOS's maximum intended gas per second
var GetArbOSSpeedLimitPerSecond func(statedb *state.StateDB) (uint64, error)

// While processing RPC only - Ask ArbOS what are the poster costs for this message.
var RPCPostingGasHook = func(msg *Message, header *types.Header, statedb *state.StateDB) (uint64, error) { return 0, nil }

// Renders a solidity error in human-readable form
var RenderRPCError func(data []byte) error

type NodeInterfaceBackendAPI interface {
	ChainConfig() *params.ChainConfig
	CurrentBlock() *types.Header
	BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error)
	HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error)
	GetLogs(ctx context.Context, blockHash common.Hash, number uint64) ([][]*types.Log, error)
	GetEVM(ctx context.Context, state *state.StateDB, header *types.Header, vmConfig *vm.Config, blockCtx *vm.BlockContext) *vm.EVM
}
