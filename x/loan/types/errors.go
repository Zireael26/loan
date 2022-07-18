package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/loan module sentinel errors
var (
	ErrSample         = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrWrongLoanState = sdkerrors.Register(ModuleName, 1, "wrongn loan state")
)
