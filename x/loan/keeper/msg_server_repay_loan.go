package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"loan/x/loan/types"
)

func (k msgServer) RepayLoan(goCtx context.Context, msg *types.MsgRepayLoan) (*types.MsgRepayLoanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get loan
	loan, found := k.GetLoan(ctx, msg.Id)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrKeyNotFound, "key %d not found", msg.Id)
	}
	// check if loan is in the correct state, approved is only state allowing repayment
	if loan.State != "approved" {
		return nil, sdkerrors.Wrapf(types.ErrWrongLoanState, "loan is not in the correct state for this action, loan is in %s state. Needs to be in approved", loan.State)
	}

	collateral, amount, borrower := k.GetLoanContent(ctx, loan)

	// !!add balance checks to make sure borrower has enough to repay!!

	// calculate interest 1% of loan amount a year: (block current - block start) * (1/blocks in a year)
	// until we have a better way to calculate block time set to standard
	// need a big Int to do math on coins amount
	divisor := sdkmath.NewInt(100)
	multiplier := sdkmath.NewInt(1)
	interest := amount[0].Amount.Mul(multiplier).Quo(divisor)
	interestPayment := sdk.NewCoin("usdc", interest.Mul(types.Cwei))


	errR := k.bankKeeper.SendCoinsFromAccountToModule(ctx, borrower, types.ModuleName, sdk.NewCoins(interestPayment))
	if errR != nil {
		return nil, errR
	}
	// send coins out to and from appropriate accounts
	// keeper burn function takes back zusd to burn keep interest out of the burn function
	errK := k.BurnTokens(ctx, borrower, sdk.NewCoin("zusd", amount[0].Amount))
	if errK != nil {
		return nil, errK
	}

	// collateral is sent back to the borrower from the module not the lender
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.Nbtp, borrower, collateral)
	if err != nil {
		return nil, err
	}
	loan.State = "repayed"
	k.SetLoan(ctx, loan)
	return &types.MsgRepayLoanResponse{}, nil
}
