package testutil

import (
	"context"

	"loan/x/loan/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	//types0 "github.com/cosmos/cosmos-sdk/x/auth/types"
	//types1 "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/golang/mock/gomock"
)

func (escrow *MockBankKeeper) ExpectAny(context context.Context) {
	/* 
	using gomock.Any() as first argument to mock any context argument 
	this allows us to test block height for deadline in both
	liquidate and repay loan
	using 2 contexts below
	*/
	escrow.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	escrow.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	escrow.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	escrow.EXPECT().GetBalance(sdk.UnwrapSDKContext(context), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin{
			return sdk.Coin{
				Denom:  denom,
				Amount: sdk.NewInt(1000),
			}
		}).AnyTimes()
	escrow.EXPECT().MintCoins(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	/*.DoAndReturn(func(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
			if amt[0].Denom == "" || amt[0].IsLTE(sdk.NewCoin(amt[0].Denom, sdk.NewInt(0))) {
				return errors.New("invalid amount")
			}
			if moduleName == "" || moduleName[0:6] != "cosmos" {
				return errors.New("invalid module name")
			}
			return nil
			}).AnyTimes()*/
	escrow.EXPECT().BurnCoins(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	/*
	set two different block heights to test deadline
	for liquidate and repay loan
	*/
	ctx1 := sdk.UnwrapSDKContext(context)
	ctx1 = ctx1.WithBlockHeight(0)
	ctx2 := sdk.UnwrapSDKContext(context)
	ctx2 = ctx2.WithBlockHeight(1000000000000000000)
}

func coinsOf(amount uint64) sdk.Coins {
	return sdk.Coins{
			sdk.Coin{
					Denom:  sdk.DefaultBondDenom,
					Amount: sdk.NewInt(int64(amount)),
			},
	}
}

func (escrow *MockBankKeeper) ExpectPay(context context.Context, who string, amount uint64) *gomock.Call {
	whoAddr, err := sdk.AccAddressFromBech32(who)
	if err != nil {
		panic(err)
	}
	return escrow.EXPECT().SendCoinsFromAccountToModule(sdk.UnwrapSDKContext(context), whoAddr, types.ModuleName, coinsOf(amount))
}

func (escrow *MockBankKeeper) ExpectRefund(context context.Context, who string, amount uint64) *gomock.Call {
	whoAddr, err := sdk.AccAddressFromBech32(who)
	if err != nil {
		panic(err)
	}
	return escrow.EXPECT().SendCoinsFromModuleToAccount(sdk.UnwrapSDKContext(context), types.ModuleName, whoAddr, coinsOf(amount))
}
