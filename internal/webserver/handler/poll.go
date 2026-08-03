package handler

import (
	"context"
	"genshin-quiz/generated/oapi"

	services "genshin-quiz/internal/services/poll"
)

func (h *Handler) GetPolls(
	ctx context.Context,
	req oapi.GetPollsRequestObject,
) (oapi.GetPollsResponseObject, error) {
	res, err := services.GetPolls(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *Handler) GetPoll(
	ctx context.Context,
	req oapi.GetPollRequestObject,
) (oapi.GetPollResponseObject, error) {
	res, err := services.GetPoll(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *Handler) PostCreatePoll(
	ctx context.Context,
	req oapi.PostCreatePollRequestObject,
) (oapi.PostCreatePollResponseObject, error) {
	res, err := services.PostCreatePoll(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *Handler) PostVotePoll(
	ctx context.Context,
	req oapi.PostVotePollRequestObject,
) (oapi.PostVotePollResponseObject, error) {
	res, err := services.PostVotePoll(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *Handler) PostLikePoll(
	ctx context.Context,
	req oapi.PostLikePollRequestObject,
) (oapi.PostLikePollResponseObject, error) {
	err := services.PostLikePoll(ctx, h.app, req)
	if err != nil {
		return nil, err
	}
	return oapi.PostLikePoll201Response{}, nil
}
