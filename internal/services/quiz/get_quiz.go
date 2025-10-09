package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
)

func GetQuizzes(
	ctx context.Context,
	app *config.App,
	req oapi.GetQuizzesRequestObject,
) error {
	return nil
}
