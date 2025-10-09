package cronjob

import (
	"genshin-quiz/config"
)

type Cronjob struct {
	app *config.App
}

func NewCronjob(app *config.App) *Cronjob {
	return &Cronjob{
		app: app,
	}
}
