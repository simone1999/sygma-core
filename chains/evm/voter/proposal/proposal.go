package proposal

import (
	"github.com/ChainSafe/chainbridge-core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func NewProposal(source uint8, depositNonce uint64, resourceId types.ResourceID, data []byte, handlerAddress, bridgeAddress common.Address, depositTxHash common.Hash, depositBlock uint64) *Proposal {
	return &Proposal{
		DepositTxHash:  depositTxHash,
		DepositBlock:   depositBlock,
		Source:         source,
		DepositNonce:   depositNonce,
		ResourceId:     resourceId,
		Data:           data,
		HandlerAddress: handlerAddress,
		BridgeAddress:  bridgeAddress,
	}
}

type Proposal struct {
	DepositTxHash  common.Hash // transaction hash of the deposit transaction
	DepositBlock   uint64      // Block the Deposit transaction were made
	Source         uint8       // Source domainID where message was initiated
	DepositNonce   uint64      // Nonce for the deposit
	ResourceId     types.ResourceID
	Payload        []interface{} // data associated with event sequence
	Data           []byte
	HandlerAddress common.Address
	BridgeAddress  common.Address
}

// GetDataHash constructs and returns proposal data hash
func (p *Proposal) GetDataHash() common.Hash {
	return crypto.Keccak256Hash(append(p.HandlerAddress.Bytes(), p.Data...))
}

// GetID constructs proposal unique identifier
func (p *Proposal) GetID() common.Hash {
	return crypto.Keccak256Hash(append([]byte{p.Source}, byte(p.DepositNonce)))
}
