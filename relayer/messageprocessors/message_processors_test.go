package messageprocessors

import (
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/types"
	"math/big"
	"testing"
)

// TestRouter tests relayers router
func TestAdjustDecimalsForERC20AmountMessageProcessor(t *testing.T) {
	a, _ := big.NewInt(0).SetString("145556700000000000000", 10) // 145.5567 tokens
	msg := &message.Message{
		Destination: 2,
		Source:      1,
		Payload: []interface{}{
			a.Bytes(), // 145.5567 tokens
		},
	}
	err := AdjustDecimalsForERC20AmountMessageProcessor(map[uint8]uint8{1: 18, 2: 2})(msg)
	if err != nil {
		t.Fatal()
	}
	amount := new(big.Int).SetBytes(msg.Payload[0].([]byte))
	if amount.Cmp(big.NewInt(14555)) != 0 {
		t.Fatal(amount.String())
	}
	msg2 := &message.Message{
		Destination: 1,
		Source:      2,
		Payload: []interface{}{
			big.NewInt(14555).Bytes(), // 145.55 tokens from 2nd chain
		},
	}
	err = AdjustDecimalsForERC20AmountMessageProcessor(map[uint8]uint8{1: 18, 2: 2})(msg2)
	if err != nil {
		t.Fatal()
	}
	a2, _ := big.NewInt(0).SetString("145550000000000000000", 10)
	amount2 := new(big.Int).SetBytes(msg2.Payload[0].([]byte))
	if amount2.Cmp(a2) != 0 {
		t.Fatal()
	}
}

// TestRouter tests relayers router
func TestAdjustDecimalsForIndividualERC20AmountMessageProcessor(t *testing.T) {
	a, _ := big.NewInt(0).SetString("145556700000000000000", 10) // 145.5567 tokens
	msg := &message.Message{
		Destination: 2,
		Source:      1,
		ResourceId:  types.ResourceID{1},
		Payload: []interface{}{
			a.Bytes(), // 145.5567 tokens
		},
	}
	decimalMap := map[uint8]map[types.ResourceID]uint8{1: {types.ResourceID{1}: 18}, 2: {types.ResourceID{1}: 2}}
	err := AdjustDecimalsForIndividualERC20AmountMessageProcessor(decimalMap)(msg)
	if err != nil {
		t.Fatal()
	}
	amount := new(big.Int).SetBytes(msg.Payload[0].([]byte))
	if amount.Cmp(big.NewInt(14555)) != 0 {
		t.Fatal(amount.String())
	}
	msg2 := &message.Message{
		Destination: 1,
		Source:      2,
		ResourceId:  types.ResourceID{1},
		Payload: []interface{}{
			big.NewInt(14555).Bytes(), // 145.55 tokens from 2nd chain
		},
	}
	err = AdjustDecimalsForIndividualERC20AmountMessageProcessor(decimalMap)(msg2)
	if err != nil {
		t.Fatal()
	}
	a2, _ := big.NewInt(0).SetString("145550000000000000000", 10)
	amount2 := new(big.Int).SetBytes(msg2.Payload[0].([]byte))
	if amount2.Cmp(a2) != 0 {
		t.Fatal()
	}
}
