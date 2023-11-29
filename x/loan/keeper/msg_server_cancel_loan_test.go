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



func TestCancelLoan(t *testing.T) {
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
		
		// change loan state to requested
		loan.State = "requested"
		k.SetLoan(ctx, loan)
		cancelLoan, err := msgServer.CancelLoan(ctx, &types.MsgCancelLoan{Id: loan.Id, Creator: loan.Borrower})
		require.NoError(t, err)
		_ = cancelLoan
		fmt.Printf("cancelLoan passed\n")
}