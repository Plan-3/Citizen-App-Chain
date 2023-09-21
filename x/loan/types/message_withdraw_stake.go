package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgWithdrawStake = "withdraw_stake"

var _ sdk.Msg = &MsgWithdrawStake{}

func NewMsgWithdrawStake(creator string) *MsgWithdrawStake {
	return &MsgWithdrawStake{
		Creator: creator,
	}
}

func (msg *MsgWithdrawStake) Route() string {
	return RouterKey
}

func (msg *MsgWithdrawStake) Type() string {
	return TypeMsgWithdrawStake
}

func (msg *MsgWithdrawStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgWithdrawStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgWithdrawStake) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
