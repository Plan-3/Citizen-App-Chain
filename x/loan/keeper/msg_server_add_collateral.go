package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"loan/x/loan/types"
)

func (k msgServer) AddCollateral(goCtx context.Context, msg *types.MsgAddCollateral) (*types.MsgAddCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// retrieve loan to approve k is the msgServer object getLoan is a method of a keeper
	loan, found := k.GetLoan(ctx, msg.Id)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "loan %d doesn't exist", msg.Id)
	}

	if loan.State != "approved" {
		return nil, sdkerrors.Wrapf(types.ErrWrongLoanState, "loan %d is not in requested state", msg.Id)
	}

	collateral, amount, borrower := k.GetLoanContent(ctx, loan)

	getCwei := amount[0].Amount.Mul(types.Cwei)

	addition := collateral[0].Amount.Add(getCwei)

	// updated collateral
	newCollateral := sdk.NewCoin(collateral[0].Denom, addition)


	cCoin := sdk.NewCoin(amount[0].Denom, getCwei)

	// send coins to the module account that holds collateral
	sdkError := k.bankKeeper.SendCoinsFromAccountToModule(ctx, borrower, types.Nbtp, sdk.NewCoins(cCoin))
	if sdkError != nil {
		return nil, sdkError
	}

	// update values
	loan.Collateral = newCollateral.String()
	// store updated loan values
	k.SetLoan(ctx, loan)

	return &types.MsgAddCollateralResponse{}, nil
}
