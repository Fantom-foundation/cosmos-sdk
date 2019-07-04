package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// AddVote Adds a vote on a specific proposal
func (keeper Keeper) AddVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress, option types.VoteOption) sdk.Error {
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return types.ErrUnknownProposal(keeper.codespace, proposalID)
	}
	if proposal.Status != types.StatusVotingPeriod {
		return types.ErrInactiveProposal(keeper.codespace, proposalID)
	}

	if !types.ValidVoteOption(option) {
		return types.ErrInvalidVote(keeper.codespace, option)
	}

	vote := types.NewVote(proposalID, voterAddr, option)
	keeper.SetVote(ctx, vote)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalVote,
			sdk.NewAttribute(types.AttributeKeyOption, option.String()),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	return nil
}

// GetAllVotes returns all the votes from the store
func (keeper Keeper) GetAllVotes(ctx sdk.Context) (votes types.Votes) {
	keeper.IterateAllVotes(ctx, func(vote types.Vote) bool {
		votes = append(votes, vote)
		return false
	})
	return
}

// GetVotes returns all the votes from a proposal
func (keeper Keeper) GetVotes(ctx sdk.Context, proposalID uint64) (votes types.Votes) {
	keeper.IterateVotes(ctx, proposalID, func(vote types.Vote) bool {
		votes = append(votes, vote)
		return false
	})
	return
}

// GetVote gets the vote from an address on a specific proposal
func (keeper Keeper) GetVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) (vote types.Vote, found bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.VoteKey(proposalID, voterAddr))
	if bz == nil {
		return vote, false
	}

	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &vote)
	return vote, true
}

// SetVote sets a Vote to the gov store
func (keeper Keeper) SetVote(ctx sdk.Context, vote types.Vote) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(vote)
	store.Set(types.VoteKey(vote.ProposalID, vote.Voter), bz)
}

// GetVotesIterator gets all the votes on a specific proposal as an sdk.Iterator
func (keeper Keeper) GetVotesIterator(ctx sdk.Context, proposalID uint64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, types.VotesKey(proposalID))
}

func (keeper Keeper) deleteVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(types.VoteKey(proposalID, voterAddr))
}
