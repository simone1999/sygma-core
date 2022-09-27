package erc20

import (
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/consts"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/contracts"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor"
	"github.com/ChainSafe/chainbridge-core/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rs/zerolog/log"
	"strings"
)

type ERC20HandlerContract struct {
	contracts.Contract
}

func NewERC20HandlerContract(
	client calls.ContractCallerDispatcher,
	erc20HandlerContractAddress common.Address,
	t transactor.Transactor,
) *ERC20HandlerContract {
	a, _ := abi.JSON(strings.NewReader(consts.ERC20HandlerABI))
	b := common.FromHex(consts.ERC20HandlerBin)
	return &ERC20HandlerContract{contracts.NewContract(erc20HandlerContractAddress, a, b, client, t)}
}

func (c *ERC20HandlerContract) ResourceIdToTokenContractAddress(resourceID types.ResourceID) (common.Address, error) {
	log.Debug().Msgf("Getting token address from resourceID %s", hexutil.Encode(resourceID[:]))
	res, err := c.CallContract("_resourceIDToTokenContractAddress", resourceID)
	if err != nil {
		return common.Address{}, err
	}
	out := *abi.ConvertType(res[0], new(common.Address)).(*common.Address)
	return out, nil
}
