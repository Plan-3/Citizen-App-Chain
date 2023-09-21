package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"loan/x/loan/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		bankKeeper types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,

	bankKeeper types.BankKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,

		bankKeeper: bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) MintTokens(ctx sdk.Context, receiver sdk.AccAddress, tokens sdk.Coin) error {
	// mint new tokens if the source of the transfer is the same chain
	if err := k.bankKeeper.MintCoins(
		ctx, types.ModuleName, sdk.NewCoins(tokens),
	); err != nil {
		return err
	}

	// send to receiver
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx, types.ModuleName, receiver, sdk.NewCoins(tokens),
	); err != nil {
		panic(fmt.Sprintf("unable to send coins from module to account despite previously minting coins to module account: %v", err))
	}

	return nil
}

func (k Keeper) BurnTokens(ctx sdk.Context, receiver sdk.AccAddress, tokens sdk.Coin) error {

	// send to receiver
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, receiver, types.ModuleName, sdk.NewCoins(tokens),
	); err != nil {
		panic(fmt.Sprintf("unable to send coins from account to module: %v", err))
	}

	// mint new tokens if the source of the transfer is the same chain
	if err := k.bankKeeper.BurnCoins(
		ctx, types.ModuleName, sdk.NewCoins(tokens),
	); err != nil {
		return err
	}

	return nil
}

func (k Keeper) TypedLoan(ctx sdk.Context, token sdk.Coins) *types.TokenPrice {
	// set up pointer to TokenPrice collateral price
	collateralPrice := &types.TokenPrice{}

	// switch on denom string to set parsed coin t a type TokenPrice{sdk.Coin, int}
	switch token[0].Denom {
	case "ctz":
		collateralPrice.Denom = token[0]
		collateralPrice.Price = 1800
		break
	case "cqt":
		collateralPrice.Denom = token[0]
		collateralPrice.Price = 100
		break
	default:
		break
	}
	return collateralPrice
}

func (k Keeper) ModuleStakingAmounts(ctx sdk.Context) (sdk.Int, sdk.Int, sdk.Int, sdk.Coins) {

	ModuleAccountToAddress, _ := sdk.AccAddressFromBech32("cosmos1gu4m79yj8ch8em7c22vzt3qparg69ymm75qf6l")

	/*
	* type the module account to type Balance
	* Balance has getAddress and getCoins methods
	*/ 

	moduleBalances := k.bankKeeper.GetAccountsBalances(ctx)
	var loanModule banktypes.Balance
	for _, accounts := range moduleBalances {
		if accounts.GetAddress().Equals(ModuleAccountToAddress) {
			loanModule = accounts
		}
	}

	moduleCoins := loanModule.GetCoins()

	// set up variables to hold collateral prices and total zusd in bank vault at time of deposit
	ctzPrice := sdk.NewInt(0)
	cqtPrice := sdk.NewInt(0)
	zusdTotalAtTimeOfDeposit := sdk.NewInt(0)

	/*
	* loop through all coins in module account i.e. bank vault
	* get price of collateral coins
	* get total zusd in bank vault at time of deposit
	* will need to add new cases as new collaterals are accepted
	*/

	for _, coin := range moduleCoins {
		switch coin.Denom {
		case "ctz":
			ctzPrice = coin.Amount
			break
		case "cqt":
			cqtPrice = coin.Amount
			break
		case "zusd":
			zusdTotalAtTimeOfDeposit = coin.Amount
			break
		}
	}

	return ctzPrice, cqtPrice, zusdTotalAtTimeOfDeposit, moduleCoins
}

func (k Keeper) GetLoanContent(ctx sdk.Context, loan types.Loan) (sdk.Coins, sdk.Coins, sdk.AccAddress){

	borrower, _ := sdk.AccAddressFromBech32(loan.Borrower)
	collateral, _ := sdk.ParseCoinsNormalized(loan.Collateral)
	amount, _ := sdk.ParseCoinsNormalized(loan.Amount)

	return collateral, amount, borrower
}

/*
* add more keeper methods here
*/
