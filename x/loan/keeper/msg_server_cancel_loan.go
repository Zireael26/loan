package keeper

import (
	"context"
	"strings"

	"loan/x/loan/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CancelLoan(goCtx context.Context, msg *types.MsgCancelLoan) (*types.MsgCancelLoanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Try and find loan in store using ID
	loan, isFound := k.GetLoan(ctx, msg.Id)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "Loan with specified ID not found in store")
	}

	// Check if the actor is authorised, i.e. The borrower is the same as msg.Creator
	if strings.Compare(msg.Creator, loan.Borrower) != 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "You are not authorized to cancel this loan")
	}

	// Check if loan state is appropriate, i.e. The loan is in the requested state/not yet approved
	if strings.Compare(loan.State, "requested") != 0 {
		return nil, sdkerrors.Wrapf(types.ErrWrongLoanState, "%v", loan.State)
	}

	// Parse required entities into appropriate types
	borrower, _ := sdk.AccAddressFromBech32(loan.Borrower)
	collateral, _ := sdk.ParseCoinsNormalized(loan.Collateral)

	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrower, collateral)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "Failed to return collateral to borrower")
	}

	loan.State = "cancelled"
	k.SetLoan(ctx, loan)

	return &types.MsgCancelLoanResponse{}, nil
}
