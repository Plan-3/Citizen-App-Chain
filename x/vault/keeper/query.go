package keeper

import (
	"loan/x/vault/types"
)

var _ types.QueryServer = Keeper{}
