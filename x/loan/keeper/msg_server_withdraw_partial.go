
package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"loan/x/loan/types"
)

func (k msgServer) WithdrawPartial(goCtx context.Context, msg *types.MsgWithdrawPartial) (*types.MsgWithdrawPartialResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: Handling the message
	// retrieve loan to approve k is the msgServer object getLoan is a method of a keeper
	loan, found := k.GetLoan(ctx, msg.Id)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrKeyNotFound, "loan %d doesn't exist", msg.Id)
	}

	if loan.State != "approved" {
		return nil, sdkerrors.Wrapf(types.ErrWrongLoanState, "loan %d is not in requested state", msg.Id)
	}
	
	collateral, loanAmount, borrower := k.GetLoanContent(ctx, loan)
	
	// get loan amount
	amount, err := sdk.ParseCoinsNormalized(msg.Amount)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidRequest, "Can't parse amount")
	}

	
	// collateral and amount should be in cwei units already 
	// add to var for easier use
	dollarAmount := amount[0].Amount
	collateralPrice := collateral[0].Amount
	
	// get first part of fraction
	// eg 1800000000000 / 900000000000 = 2
	fraction := collateralPrice.Quo(dollarAmount)
	// get the percentage 
	// eg 100 / 2 = 50
	percentage := sdk.NewInt(100).Quo(fraction)
	
	// get the amount of zusd to send back
	// eg 1000 * 50 = 50000 / 100 = 500
	// dividing by zero issue
	zusdToSendBack := loanAmount[0].Amount.Mul(percentage).QuoRaw(100)
	
	
	// type to a coin 
	ztsbCoin := sdk.NewCoin("zusd", zusdToSendBack)
	
	// burn the zusd
	err2 := k.BurnTokens(ctx, borrower, ztsbCoin)
	if err2 != nil {
		return nil, err2
	}
	
	// send coins back to account from collateral holder module account
	sdkError := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.Nbtp, borrower, amount)
	if sdkError != nil {
		return nil, sdkError
	}
	
	// update loan values
	newLoanAmount := loanAmount[0].Amount.Sub(zusdToSendBack)
	newLoanAmountCoin := sdk.NewCoin("zusd", newLoanAmount)
	loan.Amount = newLoanAmountCoin.String()
	newCollateralAmount := collateralPrice.Sub(dollarAmount)
	newCollateralAmountCoin := sdk.NewCoin(collateral[0].Denom, newCollateralAmount)
	loan.Collateral = newCollateralAmountCoin.String()
	
	// store updated loan values
	k.SetLoan(ctx, loan)
	
	return &types.MsgWithdrawPartialResponse{}, nil
}