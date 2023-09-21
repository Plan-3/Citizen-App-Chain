package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRequestLoan{}, "loan/RequestLoan", nil)
	cdc.RegisterConcrete(&MsgRepayLoan{}, "loan/RepayLoan", nil)
	cdc.RegisterConcrete(&MsgLiquidateLoan{}, "loan/LiquidateLoan", nil)
	cdc.RegisterConcrete(&MsgCancelLoan{}, "loan/CancelLoan", nil)
	cdc.RegisterConcrete(&MsgRedeem{}, "loan/Redeem", nil)
	cdc.RegisterConcrete(&MsgStake{}, "loan/Stake", nil)
	cdc.RegisterConcrete(&MsgWithdrawStake{}, "loan/WithdrawStake", nil)
	cdc.RegisterConcrete(&MsgAddCollateral{}, "loan/AddCollateral", nil)
	cdc.RegisterConcrete(&MsgWithdrawPartial{}, "loan/WithdrawPartial", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRequestLoan{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRepayLoan{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgLiquidateLoan{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCancelLoan{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRedeem{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgStake{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgWithdrawStake{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddCollateral{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgWithdrawPartial{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
