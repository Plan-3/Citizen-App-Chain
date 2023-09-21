package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgWithdrawPartial = "withdraw_partial"

var _ sdk.Msg = &MsgWithdrawPartial{}

func NewMsgWithdrawPartial(creator string) *MsgWithdrawPartial {
	return &MsgWithdrawPartial{
		Creator: creator,
	}
}

func (msg *MsgWithdrawPartial) Route() string {
	return RouterKey
}

func (msg *MsgWithdrawPartial) Type() string {
	return TypeMsgWithdrawPartial
}

func (msg *MsgWithdrawPartial) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgWithdrawPartial) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgWithdrawPartial) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
