package listener_test

import (
	"math/big"
	"testing"

	"github.com/ChainSafe/chainbridge-core/chains/evm/listener"
	mock_listener "github.com/ChainSafe/chainbridge-core/chains/evm/listener/mock"
	"github.com/ChainSafe/chainbridge-core/relayer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type EVMListenerTestSuite struct {
	suite.Suite
	chainReaderMock  *mock_listener.MockChainClient
	eventHandlerMock *mock_listener.MockEventHandler
	gomockController *gomock.Controller
}

func TestRunTestSuite(t *testing.T) {
	suite.Run(t, new(EVMListenerTestSuite))
}

func (s *EVMListenerTestSuite) SetupSuite()    {}
func (s *EVMListenerTestSuite) TearDownSuite() {}
func (s *EVMListenerTestSuite) SetupTest() {
	s.gomockController = gomock.NewController(s.T())
	s.chainReaderMock = mock_listener.NewMockChainClient(s.gomockController)
	s.eventHandlerMock = mock_listener.NewMockEventHandler(s.gomockController)
}
func (s *EVMListenerTestSuite) TearDownTest() {}

func (s *EVMListenerTestSuite) Test() {

	s.chainReaderMock.EXPECT().LatestBlock().DoAndReturn(
		func() (*big.Int, error) {
			return big.NewInt(0), nil
		},
	)

	s.chainReaderMock.EXPECT().FetchDepositLogs(
		gomock.Any, gomock.Eq(), gomock.Eq(), gomock.Eq(),
	).DoAndReturn(
		func() (*[]listener.DepositLogs, error) {
			return nil, nil
		},
	)

	s.eventHandlerMock.EXPECT().HandleEvent(
		gomock.Eq(), gomock.Eq(), gomock.Eq(), gomock.Eq(),
	).DoAndReturn(
		func() (*relayer.Message, error) {
			return nil, nil
		},
	)

	l := listener.NewEVMListener(
		s.chainReaderMock, s.eventHandlerMock, common.Address{},
	)

	l.ListenToEvents()
}
