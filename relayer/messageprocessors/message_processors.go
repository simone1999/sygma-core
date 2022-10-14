package messageprocessors

import (
	"errors"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/contracts/bridge"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/contracts/erc20"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor"
	"github.com/ChainSafe/chainbridge-core/config/chain"
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"

	"github.com/rs/zerolog/log"
)

type MessageProcessor func(message *message.Message) error

// AdjustDecimalsForERC20AmountMessageProcessor is a function, that accepts message and map[domainID uint8]{decimal uint}
// using this  params processor converts amount for one chain to another for provided decimals with floor rounding
func AdjustDecimalsForERC20AmountMessageProcessor(args ...interface{}) MessageProcessor {
	return func(m *message.Message) error {
		if len(args) == 0 {
			return errors.New("processor requires 1 argument")
		}
		decimalsMap, ok := args[0].(map[uint8]uint8)
		if !ok {
			return errors.New("no decimals map found in args")
		}
		sourceDecimal, ok := decimalsMap[m.Source]
		if !ok {
			return errors.New("no source decimals found at decimalsMap")
		}
		destDecimal, ok := decimalsMap[m.Destination]
		if !ok {
			return errors.New("no destination decimals found at decimalsMap")
		}
		amountByte, ok := m.Payload[0].([]byte)
		if !ok {
			return errors.New("could not cast interface to byte slice")
		}
		amount := new(big.Int).SetBytes(amountByte)
		if sourceDecimal > destDecimal {
			diff := sourceDecimal - destDecimal
			roundedAmount := big.NewInt(0)
			roundedAmount.Div(amount, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(0).SetUint64(uint64(diff)), nil))
			log.Info().Msgf("amount %s rounded to %s from chain %v to chain %v", amount.String(), roundedAmount.String(), m.Source, m.Destination)
			m.Payload[0] = roundedAmount.Bytes()
			return nil
		}
		if sourceDecimal < destDecimal {
			diff := destDecimal - sourceDecimal
			roundedAmount := big.NewInt(0)
			roundedAmount.Mul(amount, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(0).SetUint64(uint64(diff)), nil))
			m.Payload[0] = roundedAmount.Bytes()
			log.Info().Msgf("amount %s rounded to %s from chain %v to chain %v", amount.String(), roundedAmount.String(), m.Source, m.Destination)
		}
		return nil
	}
}

// AdjustDecimalsForIndividualERC20AmountMessageProcessor is a function, that accepts message and map[domainID uint8]{decimal uint}
// using this  params processor converts amount for one chain to another for provided decimals with floor rounding
func AdjustDecimalsForIndividualERC20AmountMessageProcessor(args ...interface{}) MessageProcessor {
	return func(m *message.Message) error {
		if len(args) == 0 {
			return errors.New("processor requires 1 argument")
		}
		decimalsTokenMap, ok := args[0].(map[uint8]map[types.ResourceID]uint8)
		if !ok {
			return errors.New("no decimals map found in args")
		}
		sourceDecimals, ok := decimalsTokenMap[m.Source]
		if !ok {
			return errors.New("no source chain decimals found at decimalsMap")
		}
		sourceDecimal, ok := sourceDecimals[m.ResourceId]
		if !ok {
			return errors.New("no source token decimals found at decimalsMap")
		}
		destDecimals, ok := decimalsTokenMap[m.Destination]
		if !ok {
			return errors.New("no destination chain decimals found at decimalsMap")
		}
		destDecimal, ok := destDecimals[m.ResourceId]
		if !ok {
			return errors.New("no destination token decimals found at decimalsMap")
		}
		amountByte, ok := m.Payload[0].([]byte)
		if !ok {
			return errors.New("could not cast interface to byte slice")
		}
		amount := new(big.Int).SetBytes(amountByte)
		if sourceDecimal > destDecimal {
			diff := sourceDecimal - destDecimal
			roundedAmount := big.NewInt(0)
			roundedAmount.Div(amount, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(0).SetUint64(uint64(diff)), nil))
			log.Info().Msgf("amount %s rounded to %s from chain %v to chain %v", amount.String(), roundedAmount.String(), m.Source, m.Destination)
			m.Payload[0] = roundedAmount.Bytes()
			return nil
		}
		if sourceDecimal < destDecimal {
			diff := destDecimal - sourceDecimal
			roundedAmount := big.NewInt(0)
			roundedAmount.Mul(amount, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(0).SetUint64(uint64(diff)), nil))
			m.Payload[0] = roundedAmount.Bytes()
			log.Info().Msgf("amount %s rounded to %s from chain %v to chain %v", amount.String(), roundedAmount.String(), m.Source, m.Destination)
		}
		return nil
	}
}

// AdjustDecimalsForERC20AmountMessageAutoProcessor is a function, that requires clients, transactors and evmConfigs as maps for each chain
// using this  params processor automatically converts amount for one token's decimals to another with floor rounding
// by looking up the tokens decimals on chain
func AdjustDecimalsForERC20AmountMessageAutoProcessor(args ...interface{}) MessageProcessor {
	return func(m *message.Message) error {
		if m.Type != message.FungibleTransfer {
			return nil
		}
		if len(args) != 3 {
			return errors.New("processor requires 3 arguments")
		}

		clients, ok := args[0].(map[uint8]calls.ContractCallerDispatcher)
		if !ok {
			return errors.New("no client found in args")
		}
		transactors, ok := args[1].(map[uint8]transactor.Transactor)
		if !ok {
			return errors.New("no transactor found in args")
		}
		evmConfigs, ok := args[2].(map[uint8]*chain.EVMConfig)
		if !ok {
			return errors.New("no evmConfig found in args")
		}
		sourceClient := clients[m.Source]
		destClient := clients[m.Destination]
		sourceTransactor := transactors[m.Source]
		destTransactor := transactors[m.Destination]

		sourceBridge := bridge.NewBridgeContract(sourceClient, common.HexToAddress(evmConfigs[m.Source].Bridge), sourceTransactor)
		destBridge := bridge.NewBridgeContract(destClient, common.HexToAddress(evmConfigs[m.Destination].Bridge), destTransactor)

		sourceHandlerAddress, err := sourceBridge.GetHandlerAddressForResourceID(m.ResourceId)
		if err != nil {
			return err
		}
		destHandlerAddress, err := destBridge.GetHandlerAddressForResourceID(m.ResourceId)
		if err != nil {
			return err
		}
		sourceHandler := erc20.NewERC20HandlerContract(sourceClient, sourceHandlerAddress, sourceTransactor)
		destHandler := erc20.NewERC20HandlerContract(destClient, destHandlerAddress, destTransactor)

		sourceTokenAddress, err := sourceHandler.ResourceIdToTokenContractAddress(m.ResourceId)
		if err != nil {
			return err
		}
		destTokenAddress, err := destHandler.ResourceIdToTokenContractAddress(m.ResourceId)
		if err != nil {
			return err
		}

		sourceToken := erc20.NewERC20Contract(sourceClient, sourceTokenAddress, sourceTransactor)
		destToken := erc20.NewERC20Contract(destClient, destTokenAddress, destTransactor)

		sourceDecimalPointer, err := sourceToken.GetDecimals()
		if err != nil {
			return err
		}
		destDecimalPointer, err := destToken.GetDecimals()
		if err != nil {
			return err
		}
		sourceDecimal := *sourceDecimalPointer
		destDecimal := *destDecimalPointer

		amountByte, ok := m.Payload[0].([]byte)
		if !ok {
			return errors.New("could not cast interface to byte slice")
		}
		amount := new(big.Int).SetBytes(amountByte)
		if sourceDecimal > destDecimal {
			diff := sourceDecimal - destDecimal
			roundedAmount := big.NewInt(0)
			roundedAmount.Div(amount, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(0).SetUint64(uint64(diff)), nil))
			log.Info().Msgf("amount %s rounded to %s from chain %v to chain %v", amount.String(), roundedAmount.String(), m.Source, m.Destination)
			m.Payload[0] = roundedAmount.Bytes()
		} else if sourceDecimal < destDecimal {
			diff := destDecimal - sourceDecimal
			roundedAmount := big.NewInt(0)
			roundedAmount.Mul(amount, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(0).SetUint64(uint64(diff)), nil))
			log.Info().Msgf("amount %s rounded to %s from chain %v to chain %v", amount.String(), roundedAmount.String(), m.Source, m.Destination)
			m.Payload[0] = roundedAmount.Bytes()
		}
		return nil
	}
}
