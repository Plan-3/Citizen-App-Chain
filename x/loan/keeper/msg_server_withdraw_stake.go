package keeper

import (
	"context"
	//"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	//sdkmath "cosmossdk.io/math"
	"loan/x/loan/types"
)

func (k msgServer) WithdrawStake(goCtx context.Context, msg *types.MsgWithdrawStake) (*types.MsgWithdrawStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: Handling the message

	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	amount := k.bankKeeper.GetAllBalances(ctx, creator)

	// check keeper.go for the function
	ctzPrice, cqtPrice, zusdTotalAtTimeOfWithdrawal, moduleCoins := k.ModuleStakingAmounts(ctx)

	// set up variables to hold token amounts
	cphToken := sdk.NewInt(0)
	zphToken := sdk.NewInt(0)
	posiTokenAmount := sdk.NewInt(0)

	// loop through all coins in personal account
	for _, coin := range amount {
		switch coin.Denom {
		case "cPh":
			cphToken = coin.Amount
			break
		case "zPh":
			zphToken = coin.Amount
			break
		case "posi":
			posiTokenAmount = coin.Amount
			break
		}
	}

	/*
	 create coins to be burnt
	 cToken is the collateral token
	 zToken is the zusd token
	 positionCoin is the position token
	*/ 

	cToken := sdk.NewCoin("cPh", cphToken)
	zToken := sdk.NewCoin("zPh", zphToken)
	positionCoin := sdk.NewCoin("posi", posiTokenAmount)

	// loop through all coins in personal account
	for _, coin := range amount {
		switch coin.Denom {
		case "cPh":
			k.BurnTokens(ctx, creator, cToken)
			break
		case "zPh":
			k.BurnTokens(ctx, creator, zToken)
			break
		case "posi":
			k.BurnTokens(ctx, creator, positionCoin)
			break
		}
	}
	/*
		calculate the lp % based on three conditions
		if zusd in pool is greater than the zToken amount run position/(current zusd in module + position)
		if zusd in pool is less than the zToken amount and collateral is greater than cToken amount run position/(zusd at time of deposit represented as zToken + position)
		if zusd in pool is less than the zToken amont and collateral is less than cToken amount run position/(current zusd in module + position)
	*/

	lpPercent := sdk.NewInt(0)
	if zusdTotalAtTimeOfWithdrawal.Equal(zphToken) {
		lpPercent = zusdTotalAtTimeOfWithdrawal.Quo(posiTokenAmount)
	}
	if zusdTotalAtTimeOfWithdrawal.GT(zphToken) {
		lpPercent = zusdTotalAtTimeOfWithdrawal.Add(posiTokenAmount).Quo(posiTokenAmount)
	}
	if zusdTotalAtTimeOfWithdrawal.LT(zphToken) && ctzPrice.Add(cqtPrice).GT(cphToken) {
		lpPercent = zphToken.Add(posiTokenAmount).Quo(posiTokenAmount)
	}
	if zusdTotalAtTimeOfWithdrawal.LT(zphToken) && ctzPrice.Add(cqtPrice).LT(cphToken) {
		lpPercent = zusdTotalAtTimeOfWithdrawal.Add(posiTokenAmount).Quo(posiTokenAmount)
	}


	// set up variable holders
	moduleCtz := sdk.NewInt(0)
	moduleCqt := sdk.NewInt(0)
	moduleZusd := sdk.NewInt(0)

	for _, coin := range moduleCoins {
		switch coin.Denom {
		case "ctz":
			moduleCtz = coin.Amount
			break
		case "cqt":
			moduleCqt = coin.Amount
			break
		case "zusd":
			moduleZusd = coin.Amount
			break
		}
	}

	/*
		because I will forget basic math
		x.Mul(.5) = total is the same as x/2
		instead of using decimals get a whole number from if checks on lpPercent
		then just divide the coins in the bank by that to keep the math simple
	*/

	// calculate the amount of each token to send to creator
	ctzToSend := moduleCtz.Quo(lpPercent)
	cqtToSend := moduleCqt.Quo(lpPercent)
	zusdToSend := moduleZusd.Quo(lpPercent)

	// create coins from Int
	ctzCoin := sdk.NewCoin("ctz", ctzToSend)
	cqtCoin := sdk.NewCoin("cqt", cqtToSend)
	zusdCoin := sdk.NewCoin("zusd", zusdToSend)
	withdrawableCoins := sdk.NewCoins(ctzCoin, cqtCoin, zusdCoin)

	withdraw := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creator, withdrawableCoins)
	if withdraw != nil {
		return nil, withdraw
	}
	// fmt.Println(lpPercent, zusdToSendInt, withdrawableCoins)

	return &types.MsgWithdrawStakeResponse{}, nil
}
