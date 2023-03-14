package tendermint_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	cometbytes "github.com/cometbft/cometbft/libs/bytes"
	cometproto "github.com/cometbft/cometbft/proto/tendermint/types"
	comettypes "github.com/cometbft/cometbft/types"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	ibctestingmock "github.com/cosmos/ibc-go/v7/testing/mock"
	"github.com/cosmos/ibc-go/v7/testing/simapp"
)

const (
	chainID                        = "gaia"
	chainIDRevision0               = "gaia-revision-0"
	chainIDRevision1               = "gaia-revision-1"
	clientID                       = "gaiamainnet"
	trustingPeriod   time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod        time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift    time.Duration = time.Second * 10
)

var (
	height          = clienttypes.NewHeight(0, 4)
	newClientHeight = clienttypes.NewHeight(1, 1)
	upgradePath     = []string{"upgrade", "upgradedIBCState"}
)

type TendermintTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	// TODO: deprecate usage in favor of testing package
	ctx        sdk.Context
	cdc        codec.Codec
	privVal    comettypes.PrivValidator
	valSet     *comettypes.ValidatorSet
	signers    map[string]comettypes.PrivValidator
	valsHash   cometbytes.HexBytes
	header     *ibctm.Header
	now        time.Time
	headerTime time.Time
	clientTime time.Time
}

func (suite *TendermintTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))
	// commit some blocks so that QueryProof returns valid proof (cannot return valid query if height <= 1)
	suite.coordinator.CommitNBlocks(suite.chainA, 2)
	suite.coordinator.CommitNBlocks(suite.chainB, 2)

	// TODO: deprecate usage in favor of testing package
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.cdc = app.AppCodec()

	// now is the time of the current chain, must be after the updating header
	// mocks ctx.BlockTime()
	suite.now = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	suite.clientTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	// Header time is intended to be time for any new header used for updates
	suite.headerTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	suite.privVal = ibctestingmock.NewPV()

	pubKey, err := suite.privVal.GetPubKey()
	suite.Require().NoError(err)

	heightMinus1 := clienttypes.NewHeight(0, height.RevisionHeight-1)

	val := comettypes.NewValidator(pubKey, 10)
	suite.signers = make(map[string]comettypes.PrivValidator)
	suite.signers[val.Address.String()] = suite.privVal
	suite.valSet = comettypes.NewValidatorSet([]*comettypes.Validator{val})
	suite.valsHash = suite.valSet.Hash()
	suite.header = suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight), heightMinus1, suite.now, suite.valSet, suite.valSet, suite.valSet, suite.signers)
	suite.ctx = app.BaseApp.NewContext(checkTx, cometproto.Header{Height: 1, Time: suite.now})
}

func getAltSigners(altVal *comettypes.Validator, altPrivVal comettypes.PrivValidator) map[string]comettypes.PrivValidator {
	return map[string]comettypes.PrivValidator{altVal.Address.String(): altPrivVal}
}

func getBothSigners(suite *TendermintTestSuite, altVal *comettypes.Validator, altPrivVal comettypes.PrivValidator) (*comettypes.ValidatorSet, map[string]comettypes.PrivValidator) {
	// Create bothValSet with both suite validator and altVal. Would be valid update
	bothValSet := comettypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create signer array and ensure it is in same order as bothValSet
	_, suiteVal := suite.valSet.GetByIndex(0)
	bothSigners := map[string]comettypes.PrivValidator{
		suiteVal.Address.String(): suite.privVal,
		altVal.Address.String():   altPrivVal,
	}
	return bothValSet, bothSigners
}

func TestTendermintTestSuite(t *testing.T) {
	suite.Run(t, new(TendermintTestSuite))
}
