package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
	poll_repo "genshin-quiz/internal/repository/poll"
	question_repo "genshin-quiz/internal/repository/question"
)

func GetHome(
	ctx context.Context,
	app *config.App,
	req oapi.GetHomeRequestObject,
) (*oapi.GetHome200JSONResponse, error) {
	var language *[]string
	if req.Params.Language != nil {
		lang := []string{*req.Params.Language}
		language = &lang
	}

	// 获取
	latestQuestions, err := getLatestQuestions(ctx, app, language)
	if err != nil {
		return nil, err
	}
	latestPolls, err := getLatestPolls(ctx, app, language)
	if err != nil {
		return nil, err
	}
	popularPolls, err := getPopularPolls(ctx, app, language)
	if err != nil {
		return nil, err
	}

	return &oapi.GetHome200JSONResponse{
		PopularExams:    []oapi.Exam{},
		LatestQuestions: latestQuestions,
		LatestPolls:     latestPolls,
		PopularPolls:    popularPolls,
	}, nil
}

func getLatestQuestions(
	ctx context.Context,
	app *config.App,
	language *[]string,
) ([]oapi.Question, error) {
	questionSortBy := "PublishDate"
	result, err := question_repo.GetQuestions(ctx, app.DB, dao.QuestionListParams{
		Page:       1,
		NumPerPage: 5,
		SortBy:     questionSortBy,
		SortDesc:   true,
		Language:   language,
	})
	if err != nil {
		return nil, err
	}

	dtos, err := question_repo.BuildQuestionsWithTransaction(ctx, app.DB, result)
	if err != nil {
		return nil, err
	}

	return dtos, nil
}

func getPopularPolls(
	ctx context.Context,
	app *config.App,
	language *[]string,
) ([]oapi.Poll, error) {
	sortBy := "created_at"
	result, err := poll_repo.GetPolls(ctx, app.DB, dao.PollListParams{
		Page:       1,
		NumPerPage: 5,
		Type:       "all",
		SortBy:     sortBy,
		SortDesc:   true,
		Language:   language,
	})
	if err != nil {
		return nil, err
	}

	dtos, err := poll_repo.BuildPollsWithLike(ctx, app.DB, result)
	if err != nil {
		return nil, err
	}

	return dtos, nil
}

func getLatestPolls(
	ctx context.Context,
	app *config.App,
	language *[]string,
) ([]oapi.Poll, error) {
	sortBy := "created_at"
	result, err := poll_repo.GetPolls(ctx, app.DB, dao.PollListParams{
		Page:       1,
		NumPerPage: 5,
		Type:       "all",
		SortBy:     sortBy,
		SortDesc:   true,
		Language:   language,
	})
	if err != nil {
		return nil, err
	}

	dtos, err := poll_repo.BuildPollsWithLike(ctx, app.DB, result)
	if err != nil {
		return nil, err
	}
	return dtos, nil
}
