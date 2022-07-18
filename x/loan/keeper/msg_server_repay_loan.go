package keeper

import (
	"context"
	"strings"

	"loan/x/loan/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) RepayLoan(goCtx context.Context, msg *types.MsgRepayLoan) (*types.MsgRepayLoanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Fetch the loan from store
	loan, isFound := k.GetLoan(ctx, msg.Id)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "Loan with specified ID does not exist.")
	}

	// Check if loan is approved, i.e. in correct state to be repaid
	if strings.Compare(loan.State, "approved") != 0 {
		return nil, sdkerrors.Wrap(types.ErrWrongLoanState, "Loan is not in appropriate state to be repaid.")
	}

	// Parse the msg creator, loan lender and borrower addresses
	repayingAccount, _ := sdk.AccAddressFromBech32(msg.Creator)
	borrower, _ := sdk.AccAddressFromBech32(loan.Borrower)
	lender, _ := sdk.AccAddressFromBech32(loan.Lender)

	// Check if msg.creator is authorized to repay the loan
	if !repayingAccount.Equals(borrower) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "You are not authorized to repay this loan.")
	}

	// Parse Loan Amount, Fee and Collateral to sdk.Coins
	principal, _ := sdk.ParseCoinsNormalized(loan.Amount)
	fee, _ := sdk.ParseCoinsNormalized(loan.Fee)
	collateral, _ := sdk.ParseCoinsNormalized(loan.Collateral)

	// Repay principal + fee to lender
	repayError := k.bankKeeper.SendCoins(ctx, repayingAccount, lender, principal.Add(fee...))
	if repayError != nil {
		return nil, sdkerrors.Wrap(repayError, "Error occurred while repaying.")
	}

	// Return collateral to borrower from module escrow account
	k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrower, collateral)

	loan.State = "repaid"
	k.SetLoan(ctx, loan)

	return &types.MsgRepayLoanResponse{}, nil
}
