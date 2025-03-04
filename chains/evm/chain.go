// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package evm

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/contracts/bridge"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/evmclient"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/evmgaspricer"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor/signAndSend"

	"github.com/ChainSafe/chainbridge-core/chains/evm/calls"
	"github.com/ChainSafe/chainbridge-core/chains/evm/listener"
	"github.com/ChainSafe/chainbridge-core/chains/evm/voter"
	"github.com/ChainSafe/chainbridge-core/config/chain"
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/store"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
)

type EventListener interface {
	ListenToEvents(startBlock, blockConfirmations *big.Int, blockRetryInterval time.Duration, domainID uint8, blockstore *store.BlockStore, stopChn <-chan struct{}, errChn chan<- error) <-chan *message.Message
}

type ProposalVoter interface {
	VoteProposal(message *message.Message, chainConfig *chain.EVMConfig) error
}

// EVMChain is struct that aggregates all data required for
type EVMChain struct {
	listener   EventListener
	writer     ProposalVoter
	blockstore *store.BlockStore
	config     *chain.EVMConfig
}

// SetupDefaultEVMChain sets up an EVMChain with all supported handlers configured
func SetupDefaultEVMChain(rawConfig map[string]interface{}, txFabric calls.TxFabric, blockstore *store.BlockStore) (*EVMChain, error) {
	config, err := chain.NewEVMConfig(rawConfig)
	if err != nil {
		return nil, err
	}

	client, err := evmclient.NewEVMClient(config)
	if err != nil {
		return nil, err
	}

	gasPricer := evmgaspricer.NewLondonGasPriceClient(client, nil)
	t := signAndSend.NewSignAndSendTransactor(txFabric, gasPricer, client)
	bridgeContract := bridge.NewBridgeContract(client, common.HexToAddress(config.Bridge), t)

	eventHandler := listener.NewETHEventHandler(*bridgeContract)
	mh := voter.NewEVMMessageHandler(*bridgeContract)

	for _, erc20HandlerContract := range config.Erc20Handlers {
		eventHandler.RegisterEventHandler(erc20HandlerContract, listener.Erc20EventHandler)
		mh.RegisterMessageHandler(erc20HandlerContract, voter.ERC20MessageHandler)
	}
	for _, erc721HandlerContract := range config.Erc721Handlers {
		eventHandler.RegisterEventHandler(erc721HandlerContract, listener.Erc721EventHandler)
		mh.RegisterMessageHandler(erc721HandlerContract, voter.ERC721MessageHandler)
	}
	for _, genericHandlerContract := range config.GenericHandlers {
		eventHandler.RegisterEventHandler(genericHandlerContract, listener.GenericEventHandler)
		mh.RegisterMessageHandler(genericHandlerContract, voter.GenericMessageHandler)
	}

	evmListener := listener.NewEVMListener(client, eventHandler, common.HexToAddress(config.Bridge))
	var evmVoter *voter.EVMVoter
	evmVoter, err = voter.NewVoterWithSubscription(mh, client, bridgeContract)
	if err != nil {
		log.Error().Msgf("failed creating voter with subscription: %s. Falling back to default voter.", err.Error())
		evmVoter = voter.NewVoter(mh, client, bridgeContract)
	}

	return NewEVMChain(evmListener, evmVoter, blockstore, config), nil
}

func NewEVMChain(listener EventListener, writer ProposalVoter, blockstore *store.BlockStore, config *chain.EVMConfig) *EVMChain {
	return &EVMChain{listener: listener, writer: writer, blockstore: blockstore, config: config}
}

// PollEvents is the goroutine that polls blocks and searches Deposit events in them.
// Events are then sent to eventsChan.
func (c *EVMChain) PollEvents(stop <-chan struct{}, sysErr chan<- error, eventsChan chan *message.Message) {
	log.Info().Msg("Polling Blocks...")

	startBlock, err := c.blockstore.GetStartBlock(
		*c.config.GeneralChainConfig.Id,
		c.config.StartBlock,
		c.config.GeneralChainConfig.LatestBlock,
		c.config.GeneralChainConfig.FreshStart,
	)
	if err != nil {
		sysErr <- fmt.Errorf("error %w on getting last stored block", err)
		return
	}

	ech := c.listener.ListenToEvents(startBlock, c.config.BlockConfirmations, c.config.BlockRetryInterval, *c.config.GeneralChainConfig.Id, c.blockstore, stop, sysErr)
	for {
		select {
		case <-stop:
			return
		case newEvent := <-ech:
			// Here we can place middlewares for custom logic?
			eventsChan <- newEvent
			continue
		}
	}
}

func (c *EVMChain) Write(msg *message.Message) error {
	return c.writer.VoteProposal(msg, c.config)
}

func (c *EVMChain) DomainID() uint8 {
	return *c.config.GeneralChainConfig.Id
}
