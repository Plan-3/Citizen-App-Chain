package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"loan/x/loan/types"
)

// in future maybe add liquidator is not borrower also make a storage for accounts that get liquidated then create a check for bad actors in request loan
func (k msgServer) LiquidateLoan(goCtx context.Context, msg *types.MsgLiquidateLoan) (*types.MsgLiquidateLoanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	loan, found := k.GetLoan(ctx, msg.Id)

	// convert deadline to int to compare to block height ParseInt(string, base, bitSize)
	deadline, err := strconv.ParseInt(loan.Deadline, 10, 64)
	if err != nil {
		panic(err)
	}
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrKeyNotFound, "key %d doesn't exist", msg.Id)
	}
	if (loan.Timestamp + deadline) > ctx.BlockHeight() {
		return nil, sdkerrors.Wrap(types.ErrUnauthorized, "Cannot liquidate: not past deadline")
	}
	if loan.State != "approved" {
		return nil, sdkerrors.Wrapf(types.ErrWrongLoanState, "%v", loan.State)
	}

	// returns are in order: collateral coins, amount coins, borrower address
	collateral, amount, borrower := k.GetLoanContent(ctx, loan)

	// burn 99% of collateral
	collateralBurn := collateral[0].Amount.MulRaw(99).QuoRaw(100)
	collateralLiquidatedToPool := collateral[0].Amount.Sub(collateralBurn)
	// convert to sdk.Coin to send to burn coins
	burnCoin := sdk.NewCoin(collateral[0].Denom, collateralBurn)
	liquidatedCollateral := sdk.NewCoin(collateral[0].Denom, collateralLiquidatedToPool)
	burnZusd := sdk.NewCoin("zusd", amount[0].Amount)
	errB := k.bankKeeper.BurnCoins(ctx, types.Nbtp, sdk.NewCoins(burnCoin))
	if errB != nil {
		return nil, errB
	}
	// for time being force zusd back from borrower to loan module to burn
	errB1 := k.BurnTokens(ctx, borrower, burnZusd)
	if errB1 != nil {
		return nil, errB1
	}

	// send collateral from holding account to pool
	errS := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.Nbtp, types.ModuleName, sdk.NewCoins(liquidatedCollateral))
	if errS != nil {
		return nil, errS
	}

	loan.State = "liquidated"
	k.SetLoan(ctx, loan)

	return &types.MsgLiquidateLoanResponse{}, nil
}
