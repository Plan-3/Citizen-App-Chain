package keeper_test

import (
	"fmt"
	"context"
	"testing"
	"runtime/debug"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"loan/x/loan"
	"loan/x/loan/keeper"
	"loan/x/loan/types"
	"loan/x/loan/testutil"
	keepertest "loan/testutil/keeper"
	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/golang/mock/gomock"

)

var (
	moduleaccount = "cosmos1gu4m79yj8ch8em7c22vzt3qparg69ymm75qf6l"
	alice = "cosmos1cpkpv6v0rdfjk6hfsaq4qjrt7deaa46cp84483"
)

func setUpMsgServer(t testing.TB) (types.MsgServer, keeper.Keeper, context.Context, *gomock.Controller, *testutil.MockBankKeeper) {
	ctrl := gomock.NewController(t)
	bankMock := testutil.NewMockBankKeeper(ctrl)
	k, ctx := keepertest.LoanKeeperWithMocks(t, bankMock)
	loan.InitGenesis(ctx, k, *types.DefaultGenesis())
	return keeper.NewMsgServerImpl(k),k, sdk.WrapSDKContext(ctx), ctrl, bankMock
}

func TestRequestLoan(t *testing.T) {
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
	_ = k

	bank.ExpectAny(context)
	bank.MintCoins(ctx, alice, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000))))
	bank.SendCoinsFromAccountToModule(ctx, borrower, moduleaccount, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(50))))
	// create loan
	defer ctrl.Finish()
	loan := types.Loan{
		Amount:     "10amount",
		Fee:        "50stake",
		Collateral: "10collateral",
		Deadline:   "10000",
		State:      "requested",
		Borrower:   alice,
		Lender:     moduleaccount,
		Timestamp:  ctx.BlockHeight(),
	}
	
	createdLoan, _ := msgServer.RequestLoan(ctx, &types.MsgRequestLoan{
		Amount: loan.Amount, 
		Fee: loan.Fee, 
		Collateral: loan.Collateral, 
		Deadline: loan.Deadline, 
		Creator: loan.Borrower,
	})
	// test amount is multiplied by 10**9 correctly
	loanCoinAmount, _ := sdk.ParseCoinsNormalized(loan.Amount)
	baseAmount := types.Cwei.Mul(loanCoinAmount[0].Amount)
	require.Equal(t, sdk.NewInt(10000000000), baseAmount)
	fmt.Printf("baseAmount: %v\n%v\n", baseAmount, loanCoinAmount[0].Amount)

	// test coins are sent from borrower to module account
	bank.MintCoins(ctx, alice, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000))))
	bank.SendCoinsFromAccountToModule(ctx, borrower, moduleaccount, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(50))))
	expectedBalanceBorrower := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(950))) // 1000 (initial) - 50 (loan fee)
	actualBalanceBorrower := bank.GetBalance(ctx, borrower, "stake")
	// coins returns [] we only test index 0 for now
	// use .Amount to get the amount of the coin in form of sdk.Int
	// balance is set to 1000 in the testutil/bank_escrow_helpers.go
	// can't use dynamic amount so need to manually set it using math func
	// subtract the collateral
	// this require equates to 950 = 1000 - 50
	require.Equal(t, expectedBalanceBorrower[0].Amount, actualBalanceBorrower.Amount.Sub(sdk.NewInt(50)))

	expectedBalanceLender := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(50))) // 50 (initial)
	actualBalanceLender := bank.GetBalance(ctx, lender, "stake")
	// same as balance borrower but need to subtract the 
	// expected balance of borrower from the initial balance of lender 
	// to get the expected balance of lender
	require.Equal(t, expectedBalanceLender[0].Amount, actualBalanceLender.Amount.Sub(expectedBalanceBorrower[0].Amount))


	_ = createdLoan
}
