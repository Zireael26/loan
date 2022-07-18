package keeper

import (
	"context"
	"strconv"
	"strings"

	"loan/x/loan/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) LiquidateLoan(goCtx context.Context, msg *types.MsgLiquidateLoan) (*types.MsgLiquidateLoanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Try and find loan in store using ID
	loan, isFound := k.GetLoan(ctx, msg.Id)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "Loan with specified ID not found in store")
	}

	// Check if the actor is authorised, i.e. The lender is the same as msg.Creator
	if strings.Compare(msg.Creator, loan.Lender) != 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "You are not authorized to liquidate this loan")
	}

	// Check if loan state is appropriate, i.e. The loan is approved
	if strings.Compare(loan.State, "approved") != 0 {
		return nil, sdkerrors.Wrapf(types.ErrWrongLoanState, "%v", loan.State)
	}

	lender, _ := sdk.AccAddressFromBech32(loan.Lender)
	collateral, _ := sdk.ParseCoinsNormalized(loan.Collateral)

	// Parse the deadline
	deadline, err := strconv.ParseInt(loan.Deadline, 10, 64)
	if err != nil {
		panic(err)
	}

	// Throw an error if lender is trying to liquidate the loan before the deadline
	if ctx.BlockHeight() < deadline {
		return nil, sdkerrors.Wrap(types.ErrDeadline, "Cannot liquidate loan before deadline")
	}

	// If all conditions are passed, liquidate the loan
	k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lender, collateral)
	loan.State = "liquidated"
	k.SetLoan(ctx, loan)

	return &types.MsgLiquidateLoanResponse{}, nil
}
