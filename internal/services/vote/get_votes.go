package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
)

func GetVotes(
	ctx context.Context,
	app *config.App,
	req oapi.GetVotesRequestObject,
) error {
	return nil
}
