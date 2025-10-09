package services

import (
	"context"
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
)

func CreateUser(
	ctx context.Context,
	app *config.App,
	req oapi.CreateUserRequestObject,
) error {
	return nil
}
