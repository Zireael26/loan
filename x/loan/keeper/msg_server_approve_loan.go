package keeper

import (
	"context"
	"strings"

	"loan/x/loan/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) ApproveLoan(goCtx context.Context, msg *types.MsgApproveLoan) (*types.MsgApproveLoanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get loan from store by id
	loan, isFound := k.GetLoan(ctx, msg.Id)
	// throw an error if loan is not found
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "Loan with specified ID does not exist")
	}

	// else, check loan state to see if it is equal to requested
	if strings.Compare(loan.State, "requested") != 0 {
		return nil, sdkerrors.Wrap(types.ErrWrongLoanState, "Loan is not in the correct state for an approval")
	}

	// if everything is correct, parse bordrower, lender addresses as sdk.AccAddress and amount as sdk.Coins
	borrower, _ := sdk.AccAddressFromBech32(loan.Borrower)
	lender, _ := sdk.AccAddressFromBech32(msg.Creator)
	amount, parseError := sdk.ParseCoinsNormalized(loan.Amount)
	if parseError != nil {
		return nil, sdkerrors.Wrap(types.ErrWrongLoanState, "Unable to parse tokens")
	}

	// if everything went well, transfer tokens to bowrrower account from sender account and set lender in loan in store
	transferError := k.bankKeeper.SendCoins(ctx, lender, borrower, amount)
	if transferError != nil {
		return nil, sdkerrors.Wrap(transferError, "Unable to transfer funds from lender to borrower accounts")
	}
	loan.Lender = msg.Creator
	loan.State = "approved"

	k.SetLoan(ctx, loan)

	return &types.MsgApproveLoanResponse{}, nil
}
