// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package listener

import (
	"context"
	"math/big"
	"time"

	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/evmclient"

	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/store"
	"github.com/ChainSafe/chainbridge-core/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rs/zerolog/log"
)

type EventHandler interface {
	HandleEvent(sourceID, destID uint8, nonce uint64, resourceID types.ResourceID, calldata, handlerResponse []byte, depositTxHash common.Hash, depositBlock uint64) (*message.Message, error)
}
type ChainClient interface {
	LatestBlock() (*big.Int, error)
	FetchDepositLogs(ctx context.Context, address common.Address, startBlock *big.Int, endBlock *big.Int) ([]*evmclient.DepositLogsEnriched, error)
	CallContract(ctx context.Context, callArgs map[string]interface{}, blockNumber *big.Int) ([]byte, error)
}

type EVMListener struct {
	chainReader   ChainClient
	eventHandler  EventHandler
	bridgeAddress common.Address
}

// NewEVMListener creates an EVMListener that listens to deposit events on chain
// and calls event handler when one occurs
func NewEVMListener(chainReader ChainClient, handler EventHandler, bridgeAddress common.Address) *EVMListener {
	return &EVMListener{chainReader: chainReader, eventHandler: handler, bridgeAddress: bridgeAddress}
}

func (l *EVMListener) ListenToEvents(
	startBlock, blockDelay *big.Int,
	blockRetryInterval time.Duration,
	domainID uint8,
	blockstore *store.BlockStore,
	stopChn <-chan struct{},
	errChn chan<- error,
) <-chan *message.Message {
	ch := make(chan *message.Message)
	go func() {
		for {
			select {
			case <-stopChn:
				return
			default:
				head, err := l.chainReader.LatestBlock()
				if err != nil {
					log.Error().Err(err).Msg("Unable to get latest block")
					time.Sleep(blockRetryInterval)
					continue
				}

				if startBlock == nil {
					startBlock = head
				}

				// Sleep if the difference is less than blockDelay; (latest - current) < BlockDelay
				if big.NewInt(0).Sub(head, startBlock).Cmp(blockDelay) == -1 {
					time.Sleep(blockRetryInterval)
					continue
				}

				// send endBlock to min(latest - BlockDelay, current + 99)
				endBlock := big.NewInt(0).Sub(head, blockDelay)
				if big.NewInt(0).Sub(endBlock, startBlock).Cmp(big.NewInt(99)) == 1 {
					endBlock.Add(startBlock, big.NewInt(99))
				}

				logs, err := l.chainReader.FetchDepositLogs(context.Background(), l.bridgeAddress, startBlock, endBlock)
				if err != nil {
					// Filtering logs error really can appear only on wrong configuration or temporary network problem
					// so i do no see any reason to break execution
					log.Error().Err(err).Str("DomainID", string(domainID)).Msgf("Unable to filter logs")
					continue
				}
				for _, eventLog := range logs {
					log.Debug().Msgf("Deposit log found from sender: %s in block: %v with  destinationDomainId: %v, resourceID: %X, depositNonce: %v", eventLog.SenderAddress, eventLog.DepositBlock, eventLog.DestinationDomainID, eventLog.ResourceID[:], eventLog.DepositNonce)
					m, err := l.eventHandler.HandleEvent(domainID, eventLog.DestinationDomainID, eventLog.DepositNonce, eventLog.ResourceID, eventLog.Data, eventLog.HandlerResponse, eventLog.DepositTxHash, eventLog.DepositBlock)
					if err != nil {
						log.Error().Str("startBlock", startBlock.String()).Str("endBlock", endBlock.String()).Uint8("domainID", domainID).Msgf("%v", err)
					} else {
						log.Debug().Msgf("Resolved message %v in block %v", m.String(), eventLog.DepositBlock)
						ch <- m
					}
				}
				if startBlock.Int64()%20 == 0 || startBlock.Int64()/20 != endBlock.Int64()/20 {
					// Logging process every 20 bocks to exclude spam
					log.Debug().Int64("block", endBlock.Int64()/20*20).Uint8("domainID", domainID).Msg("Queried block for deposit events")
				}
				// TODO: We can store blocks to DB inside listener or make listener send something to channel each block to save it.
				//Write to block store. Not a critical operation, no need to retry
				err = blockstore.StoreBlock(endBlock, domainID)
				if err != nil {
					log.Error().Str("block", endBlock.String()).Err(err).Msg("Failed to write latest block to blockstore")
				}
				// Goto next blocks
				startBlock.Add(endBlock, big.NewInt(1))
			}
		}
	}()
	return ch
}
