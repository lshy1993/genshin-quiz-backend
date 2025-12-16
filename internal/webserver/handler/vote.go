package handler

import (
	"context"
	"genshin-quiz/generated/oapi"

	services "genshin-quiz/internal/services/vote"
)

func (h *Handler) GetVotes(
	ctx context.Context,
	req oapi.GetVotesRequestObject,
) (oapi.GetVotesResponseObject, error) {
	res, err := services.GetVotes(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *Handler) GetVote(
	ctx context.Context,
	req oapi.GetVoteRequestObject,
) (oapi.GetVoteResponseObject, error) {
	res, err := services.GetVote(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *Handler) PostCreateVote(
	ctx context.Context,
	req oapi.PostCreateVoteRequestObject,
) (oapi.PostCreateVoteResponseObject, error) {
	res, err := services.PostCreateVote(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
