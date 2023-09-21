package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRedeem = "redeem"

var _ sdk.Msg = &MsgRedeem{}

func NewMsgRedeem(creator string, id uint64) *MsgRedeem {
	return &MsgRedeem{
		Creator: creator,
		Id:      id,
	}
}

func (msg *MsgRedeem) Route() string {
	return RouterKey
}

func (msg *MsgRedeem) Type() string {
	return TypeMsgRedeem
}

func (msg *MsgRedeem) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRedeem) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRedeem) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
