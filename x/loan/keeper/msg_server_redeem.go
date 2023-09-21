package keeper

import (
	"context"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"loan/x/loan/types"
)

type loanSorter struct {
	amount     float64
	collateral float64
	ltv        float64
}

func (k msgServer) Redeem(goCtx context.Context, msg *types.MsgRedeem) (*types.MsgRedeemResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message

	loans := k.GetAllLoan(ctx)
	loanSorted := make([]loanSorter, len(loans))
	// loop over loans to parse amount and collateral to float64 and add to loanSorter struct
	for i, loan := range loans {
		amount, _ := sdk.ParseCoinsNormalized(loan.Amount)
		collateral, _ := sdk.ParseCoinsNormalized(loan.Collateral)
		amountFloat, _ := sdkmath.LegacyDec(amount[0].Amount).Float64()
		collateralFloat, _ := sdkmath.LegacyDec(collateral[0].Amount).Float64()
		loanSorted[i] = loanSorter{
			amount:     amountFloat,
			collateral: collateralFloat,
			ltv:        amountFloat / collateralFloat,
		}
	}
	// sort loans by loan to value ratio
	sort.Slice(loanSorted, func(i, j int) bool {
		return loanSorted[i].ltv < loanSorted[j].ltv
	})

	/*  v2
	take amount of msg.redeem against first loan in loanSorted if redemption is > than loan amount
	move on to next loan in loanSorted and take the difference between the redemption amount and the loan amount
	each loan needs to call repay if closed out
	subtract redemption dollar amount from collateral dollar amount *redemptionAmount *borrowerAmount
	sende redemptionAmount to msg.Creator send *borrowerAmount to borrower
	close loan
	update loan with new amount and collateral if not closed
	*/

	loan, found := k.GetLoan(ctx, msg.Id)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrKeyNotFound, "key %d not found", msg.Id)
	}
	// check if loan is in the correct state, approved is only state allowing repayment
	if loan.State != "approved" {
		return nil, sdkerrors.Wrapf(types.ErrWrongLoanState, "loan is not in the correct state for this action, loan is in %s state. Needs to be in approved", loan.State)
	}

	// parse account from msg.Creator
	redeemer, _ := sdk.AccAddressFromBech32(msg.Creator)

	collateral, amount, borrower := k.GetLoanContent(ctx, loan)

	// the dollar amount of the collateral is handled in request loan 
	// as well as all the nano amounts
	redeemerAmount := collateral[0].Amount.MulRaw(95).QuoRaw(100)
	toCoin := sdk.NewCoin(collateral[0].Denom, redeemerAmount)
	remainderAmount := collateral[0].Amount.Sub(redeemerAmount)
	toCoinRemainder := sdk.NewCoin(collateral[0].Denom, remainderAmount)
	// burn amount from module
	errB := k.BurnTokens(ctx, redeemer, sdk.NewCoin("zusd", amount[0].Amount))
	if errB != nil {
		return nil, errB
	}

	// send collateral from module to msg.creator && loan.borrower
	errK := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.Nbtp, redeemer, sdk.NewCoins(toCoin))
	if errK != nil {
		return nil, errK
	}
	errR := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.Nbtp, borrower, sdk.NewCoins(toCoinRemainder))
	if errR != nil {
		return nil, errK
	}

	// update loan
	// TODO add a loan state in the enum for redeemed
	loan.State = "repayed"
	k.SetLoan(ctx, loan)

	return &types.MsgRedeemResponse{}, nil
}
