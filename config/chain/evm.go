package chain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/consts"
	"github.com/mitchellh/mapstructure"
)

type EVMConfig struct {
	GeneralChainConfig GeneralChainConfig
	Bridge             string
	Erc20Handlers      []string
	Erc721Handlers     []string
	GenericHandlers    []string
	MaxGasPrice        *big.Int
	GasMultiplier      *big.Float
	GasLimit           *big.Int
	StartBlock         *big.Int
	BlockConfirmations *big.Int
	BlockRetryInterval time.Duration
}

type RawEVMConfig struct {
	GeneralChainConfig `mapstructure:",squash"`
	Bridge             string   `mapstructure:"bridge"`
	Erc20Handlers      []string `mapstructure:"erc20Handlers"`
	Erc721Handlers     []string `mapstructure:"erc721Handlers"`
	GenericHandlers    []string `mapstructure:"genericHandlers"`
	MaxGasPrice        int64    `mapstructure:"maxGasPrice"`
	GasMultiplier      float64  `mapstructure:"gasMultiplier"`
	GasLimit           int64    `mapstructure:"gasLimit"`
	StartBlock         int64    `mapstructure:"startBlock"`
	BlockConfirmations int64    `mapstructure:"blockConfirmations"`
	BlockRetryInterval uint64   `mapstructure:"blockRetryInterval"`
}

func (c *RawEVMConfig) Validate() error {
	if err := c.GeneralChainConfig.Validate(); err != nil {
		return err
	}
	if c.Bridge == "" {
		return fmt.Errorf("required field chain.Bridge empty for chain %v", *c.Id)
	}
	if c.BlockConfirmations != 0 && c.BlockConfirmations < 1 {
		return fmt.Errorf("blockConfirmations has to be >=1")
	}
	return nil
}

// NewEVMConfig decodes and validates an instance of an EVMConfig from
// raw chain config
func NewEVMConfig(chainConfig map[string]interface{}) (*EVMConfig, error) {
	var c RawEVMConfig
	err := mapstructure.Decode(chainConfig, &c)
	if err != nil {
		return nil, err
	}

	err = c.Validate()
	if err != nil {
		return nil, err
	}

	c.GeneralChainConfig.ParseFlags()
	config := &EVMConfig{
		GeneralChainConfig: c.GeneralChainConfig,
		Erc20Handlers:      c.Erc20Handlers,
		Erc721Handlers:     c.Erc721Handlers,
		GenericHandlers:    c.GenericHandlers,
		Bridge:             c.Bridge,
		BlockRetryInterval: consts.DefaultBlockRetryInterval,
		GasLimit:           big.NewInt(consts.DefaultGasLimit),
		MaxGasPrice:        big.NewInt(consts.DefaultGasPrice),
		GasMultiplier:      big.NewFloat(consts.DefaultGasMultiplier),
		StartBlock:         big.NewInt(c.StartBlock),
		BlockConfirmations: big.NewInt(consts.DefaultBlockConfirmations),
	}

	if c.GasLimit != 0 {
		config.GasLimit = big.NewInt(c.GasLimit)
	}

	if c.MaxGasPrice != 0 {
		config.MaxGasPrice = big.NewInt(c.MaxGasPrice)
	}

	if c.GasMultiplier != 0 {
		config.GasMultiplier = big.NewFloat(c.GasMultiplier)
	}

	if c.BlockConfirmations != 0 {
		config.BlockConfirmations = big.NewInt(c.BlockConfirmations)
	}

	if c.BlockRetryInterval != 0 {
		config.BlockRetryInterval = time.Duration(c.BlockRetryInterval) * time.Second
	}

	return config, nil
}
