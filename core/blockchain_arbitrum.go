// Copyright 2021 The go-ethereum Authors
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

// Package core implements the Ethereum consensus protocol.
package core

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/palomachain/arbitrum-geth/common"
	"github.com/palomachain/arbitrum-geth/core/rawdb"
	"github.com/palomachain/arbitrum-geth/core/state"
	"github.com/palomachain/arbitrum-geth/core/types"
	"github.com/palomachain/arbitrum-geth/log"
	"github.com/palomachain/arbitrum-geth/rpc"
)

func (bc *BlockChain) generateGcprocRandOffset() time.Duration {
	if bc.cacheConfig.TrieTimeLimitRandomOffset > 0 {
		return time.Duration(rand.Int64N(int64(bc.cacheConfig.TrieTimeLimitRandomOffset)))
	}
	return 0
}

func (bc *BlockChain) FlushTrieDB(capLimit common.StorageSize) error {
	if bc.triedb.Scheme() == rawdb.PathScheme {
		return nil
	}

	if !bc.chainmu.TryLock() {
		return errChainStopped
	}
	defer bc.chainmu.Unlock()

	if !bc.triegc.Empty() {
		_, triegcBlockNumber := bc.triegc.Peek()
		blockNumber := uint64(-triegcBlockNumber)

		header := bc.GetHeaderByNumber(blockNumber)
		if header == nil {
			log.Warn("Reorg in progress, trie commit postponed")
		} else {
			err := bc.triedb.Commit(header.Root, true)
			if err != nil {
				return err
			}

			bc.gcproc = 0
			bc.gcprocRandOffset = bc.generateGcprocRandOffset()
			bc.lastWrite = blockNumber
		}
	}

	err := bc.triedb.Cap(capLimit)
	if err != nil {
		return err
	}

	return nil
}

// WriteBlockAndSetHeadWithTime also counts processTime, which will cause intermittent TrieDirty cache writes
func (bc *BlockChain) WriteBlockAndSetHeadWithTime(block *types.Block, receipts []*types.Receipt, logs []*types.Log, state *state.StateDB, emitHeadEvent bool, processTime time.Duration) (status WriteStatus, err error) {
	if !bc.chainmu.TryLock() {
		return NonStatTy, errChainStopped
	}
	defer bc.chainmu.Unlock()
	bc.gcproc += processTime
	return bc.writeBlockAndSetHead(block, receipts, logs, state, emitHeadEvent)
}

func (bc *BlockChain) ReorgToOldBlock(newHead *types.Block) error {
	if _, err := bc.SetCanonical(newHead); err != nil {
		return fmt.Errorf("error reorging to old block: %w", err)
	}
	return nil
}

func (bc *BlockChain) ClipToPostNitroGenesis(blockNum rpc.BlockNumber) (rpc.BlockNumber, rpc.BlockNumber) {
	currentBlock := rpc.BlockNumber(bc.CurrentBlock().Number.Uint64())
	nitroGenesis := rpc.BlockNumber(bc.Config().ArbitrumChainParams.GenesisBlockNum)
	if blockNum == rpc.LatestBlockNumber || blockNum == rpc.PendingBlockNumber {
		blockNum = currentBlock
	}
	if blockNum > currentBlock {
		blockNum = currentBlock
	}
	if blockNum < nitroGenesis {
		blockNum = nitroGenesis
	}
	return blockNum, currentBlock
}

func (bc *BlockChain) RecoverState(block *types.Block) error {
	if bc.HasState(block.Root()) {
		return nil
	}
	log.Warn("recovering block state", "num", block.Number(), "hash", block.Hash(), "root", block.Root())
	_, err := bc.recoverAncestors(block, false)
	return err
}
