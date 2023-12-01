package keeper_test

import (
	"testing"
	"runtime/debug"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"loan/x/loan/types"
	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)




func TestRepayLoan(t *testing.T) {
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
		
		// repay loan uses burn coins so add to bank escrow helpers to expect burn
		repay, err := msgServer.RepayLoan(context, &types.MsgRepayLoan{Id: loan.Id})
		require.NoError(t, err)
		_ = repay
}

