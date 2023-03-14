package ibctesting

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/cometbft/cometbft/abci/types"
	comettypes "github.com/cometbft/cometbft/types"
)

// ApplyValSetChanges takes in tmtypes.ValidatorSet and []abci.ValidatorUpdate and will return a new tmtypes.ValidatorSet which has the
// provided validator updates applied to the provided validator set.
func ApplyValSetChanges(t *testing.T, valSet *comettypes.ValidatorSet, valUpdates []abci.ValidatorUpdate) *comettypes.ValidatorSet {
	updates, err := comettypes.PB2TM.ValidatorUpdates(valUpdates)
	require.NoError(t, err)

	// must copy since validator set will mutate with UpdateWithChangeSet
	newVals := valSet.Copy()
	err = newVals.UpdateWithChangeSet(updates)
	require.NoError(t, err)

	return newVals
}
