package handler

import (
	"context"
	"genshin-quiz/generated/oapi"
)

func (*Handler) GetVotes(ctx context.Context, req oapi.GetVotesRequestObject) (oapi.GetVotesResponseObject, error) {
	return (oapi.GetVotes200JSONResponse{}), nil
}
