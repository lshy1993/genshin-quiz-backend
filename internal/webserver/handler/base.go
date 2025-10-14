package handler

import (
	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
)

type Handler struct {
	oapi.StrictServerInterface
	app *config.App
}

func NewHandler(app *config.App) *Handler {
	return &Handler{
		app: app,
	}
}

func ptr[T any](v T) *T {
	return &v
}
