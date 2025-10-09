package main

import (
	"genshin-quiz/config"
	"genshin-quiz/internal/cronjob"
)

func main() {
	app := config.NewApp()
	cronjob.NewCronjob(app)
}
