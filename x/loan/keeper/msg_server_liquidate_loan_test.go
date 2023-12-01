package keeper_test

import (
	"fmt"
	"testing"
	"runtime/debug"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"loan/x/loan/types"
	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func manipulateBlockHeight(ctx sdk.Context, height int64) sdk.Context {
	ctx = ctx.WithBlockHeight(height)
	return ctx
}

// TestLiquidateLoan tests the liquidate loan functionality
// Creates a mock loan and uses manipulateBlockHeight to set the blockheight to 1000000000000000000
// Then calls the liquidate loan function and checks that no error occured
func TestLiquidateLoan(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			t.FailNow()
		}
		}()
		msgServer, k, context, ctrl, bank := setUpMsgServer(t)
		borrower, _ := sdk.AccAddressFromBech32(alice)
		lender, _ := sdk.AccAddressFromBech32(moduleaccount)
		ctx := sdk.UnwrapSDKContext(context)
		_, _ = borrower, lender
		
		bank.ExpectAny(context)
		defer ctrl.Finish()
		// create loan
		loanRequest := createLoan(ctx, t, msgServer, k)
		// test loan is created and found
		loan, found := k.GetLoan(ctx, loanRequest.Id)
		require.True(t, found)
		
		loanRequest.Deadline = "0"
		ctx = manipulateBlockHeight(ctx, 1000000000000000000)
		liquidate, err := msgServer.LiquidateLoan(ctx, &types.MsgLiquidateLoan{Id: loan.Id})
		require.NoError(t, err)
		_ = liquidate
		fmt.Printf("liquidate passed\n")
}