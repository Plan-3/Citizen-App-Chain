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
	alice         = "cosmos1cpkpv6v0rdfjk6hfsaq4qjrt7deaa46cp84483"
	bob           = "cosmos1gxrdcutv2plpdqcm8ldg4frafy7tms0qk9lcn6"
	blackList     = map[string]bool{
		"cosmos1gxrdcutv2plpdqcm8ldg4frafy7tms0qk9lcn6": true,
	}
)

// sets up the msg server
// returns a msg server, keeper, context, controller, and bank mock
// takes in a testing.TB
// TB is interface common to T B and F testing types from go testing package
func setUpMsgServer(t testing.TB) (types.MsgServer, keeper.Keeper, context.Context, *gomock.Controller, *testutil.MockBankKeeper) {
	ctrl := gomock.NewController(t)
	bankMock := testutil.NewMockBankKeeper(ctrl)
	k, ctx := keepertest.LoanKeeperWithMocks(t, bankMock)
	loan.InitGenesis(ctx, *k, *types.DefaultGenesis())
	server := keeper.NewMsgServerImpl(*k)
	return server, *k, sdk.WrapSDKContext(ctx), ctrl, bankMock
}

// TestRequestLoan tests the request loan functionality
// Creates a mock loan and checks that no error occured
// also checks correct mulplication of amount by 10**9
// also checks that the coins are transferred from the borrower to the module account
// also checks that the coins are transferred from the module account to the lender
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
		_, _ = borrower, lender
		
		bank.ExpectAny(context)
		defer ctrl.Finish()
		// create loan
		loan := createLoan(ctx, t, msgServer, k)
		_ = loan

		// check mint not returning any error
		minted := bank.MintCoins(ctx, alice, sdk.NewCoins(sdk.NewCoin("col", sdk.NewInt(100))))
		require.NoError(t, minted)
		
		// test amount is multiplied by 10**9 correctly
		loanCoinAmount, _ := sdk.ParseCoinsNormalized(loan.Amount)
		baseAmount := types.Cwei.Mul(loanCoinAmount[0].Amount)
		expectedAmount := sdk.NewInt(1000000000).Mul(loanCoinAmount[0].Amount)
		require.Equal(t, expectedAmount, baseAmount, "amount is not equivalent to 10**9")
		
		// test coins are sent from borrower to module account
		passed, err := transferredCoinsToModule(ctx, t, bank, borrower, lender)
		require.Equal(t, "ok", passed)
		require.NoError(t, err)
		fmt.Println("Request loan test passed")
		
}

// creates a mock loan 
// takes in a context, testing.T type, msg server, and keeper
// context info about block
// testing.T type manages test state and logs test output
// msg server handles any messages sent to the module
// keeper holds info about the module
func createLoan(ctx sdk.Context, t *testing.T, msgServer types.MsgServer, k keeper.Keeper) types.Loan{
	// all test values for coin price are 1 set in keeper.go typedloan default case
	// failed risk check 100 amount collateral < 110
	// passed risk check 100 amount collateral > 110
	loan := types.Loan{
		Amount:     "100amount",
		Fee:        "50stake",
		Collateral: "200col",
		Deadline:   "10000",
		State:      "requested",
		Borrower:   alice,
		Lender:     moduleaccount,
		Timestamp:  ctx.BlockHeight(),
	}
	
	
	require.False(t, blackList[loan.Borrower], "borrower is blacklisted")

	
	createdLoan, err := msgServer.RequestLoan(ctx, &types.MsgRequestLoan{
		Amount: loan.Amount, 
		Fee: loan.Fee, 
		Collateral: loan.Collateral, 
		Deadline: loan.Deadline, 
		Creator: loan.Borrower,
	})
	// using require.NoError to check if error is nil
	require.NoError(t, err)

	loan.State = "approved"
	k.AppendLoan(ctx, loan)
	sysinfo, found := k.GetLoan(ctx, 1)
	// test loan is created
	// test loan is found with id of 0
	require.True(t, found)
	// fmt.Printf("%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n", sysinfo.Id, sysinfo.Amount, sysinfo.Fee, sysinfo.Collateral, sysinfo.Deadline, sysinfo.State, sysinfo.Borrower, sysinfo.Lender, sysinfo.Timestamp)
	// testing loan values are correctly stored
	require.Equal(t, loan.Amount, sysinfo.Amount)
	require.Equal(t, loan.Fee, sysinfo.Fee)
	require.Equal(t, loan.Collateral, sysinfo.Collateral)
	require.Equal(t, "approved", sysinfo.State)
	require.Equal(t, loan.Borrower, sysinfo.Borrower)
	require.Equal(t, loan.Lender, sysinfo.Lender)
	

	_ = createdLoan
	return loan
}

// tests that the coins are transferred from an account to the module account
func transferredCoinsToModule(ctx sdk.Context, t *testing.T, bank *testutil.MockBankKeeper, borrower sdk.AccAddress, lender sdk.AccAddress) (string, error){
	// we are not testing any of the cosmos functions here we must assume they are working correctly
	bank.SendCoinsFromAccountToModule(ctx, borrower, moduleaccount, sdk.NewCoins(sdk.NewCoin("col", sdk.NewInt(50))))
	expectedBalanceBorrower := sdk.NewCoins(sdk.NewCoin("col", sdk.NewInt(950))) // 1000 (initial) - 50 (loan fee)
	actualBalanceBorrower := bank.GetBalance(ctx, borrower, "col")
	// coins returns [] we only test index 0 for now
	// use .Amount to get the amount of the coin in form of sdk.Int
	// balance is set to 1000 in the testutil/bank_escrow_helpers.go
	// can't use dynamic amount so need to manually set it using math func
	// subtract the collateral
	// this require equates to 950 = 1000 - 50
	require.Equal(t, expectedBalanceBorrower[0].Amount, actualBalanceBorrower.Amount.Sub(sdk.NewInt(50)), "Expected balance")

	expectedBalanceLender := sdk.NewCoins(sdk.NewCoin("col", sdk.NewInt(50))) // 50 (initial)
	actualBalanceLender := bank.GetBalance(ctx, lender, "col")
	// same as balance borrower but need to subtract the 
	// expected balance of borrower from the initial balance of lender 
	// to get the expected balance of lender
	require.Equal(t, expectedBalanceLender[0].Amount, actualBalanceLender.Amount.Sub(expectedBalanceBorrower[0].Amount))
	return "ok", nil
}

// tests that the coins are transferred from the module account to an account (borrower here)
func transferredCoinsToBorrower(ctx sdk.Context, t *testing.T, bank *testutil.MockBankKeeper, borrower sdk.AccAddress, lender sdk.AccAddress) (string, error){
	// we are not testing any of the cosmos functions here we must assume they are working correctly
	bank.SendCoinsFromModuleToAccount(ctx, moduleaccount, borrower, sdk.NewCoins(sdk.NewCoin("zusd", sdk.NewInt(50))))
	expectedBalanceModule := sdk.NewCoins(sdk.NewCoin("zusd", sdk.NewInt(950))) // 1000 (initial) - 50 (loan fee)
	actualBalanceModule := bank.GetBalance(ctx, borrower, "zusd")
	// coins returns [] we only test index 0 for now
	// use .Amount to get the amount of the coin in form of sdk.Int
	// balance is set to 1000 in the testutil/bank_escrow_helpers.go
	// can't use dynamic amount so need to manually set it using math func
	// subtract the collateral
	// this require equates to 950 = 1000 - 50
	require.Equal(t, expectedBalanceModule[0].Amount, actualBalanceModule.Amount.Sub(sdk.NewInt(50)), "Expected balance")

	expectedBalanceBorrower := sdk.NewCoins(sdk.NewCoin("zusd", sdk.NewInt(50))) // 50 (initial)
	actualBalanceBorrower := bank.GetBalance(ctx, lender, "zusd")
	// same as balance borrower but need to subtract the 
	// expected balance of borrower from the initial balance of lender 
	// to get the expected balance of lender
	require.Equal(t, expectedBalanceBorrower[0].Amount, actualBalanceBorrower.Amount.Sub(expectedBalanceBorrower[0].Amount))
	return "ok", nil
}
