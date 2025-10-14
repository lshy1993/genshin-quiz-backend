package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
)

func GetQuestions(
	ctx context.Context,
	app *config.App,
	req oapi.GetQuestionsRequestObject,
) error {
	return nil
}
