package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
)

func GetExams(
	ctx context.Context,
	app *config.App,
	req oapi.GetExamsRequestObject,
) error {
	return nil
}
