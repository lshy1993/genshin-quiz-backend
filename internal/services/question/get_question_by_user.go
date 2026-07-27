package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	dao "genshin-quiz/internal/dao"
	question_repo "genshin-quiz/internal/repository/question"
	"genshin-quiz/internal/webserver/middleware"
)

func GetQuestionsByUser(
	ctx context.Context,
	app *config.App,
	req oapi.GetUserQuestionsRequestObject,
) (*oapi.GetUserQuestions200JSONResponse, error) {
	// 从 context 中获取用户信息
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, common.ErrUserNotInContext
	}

	page := 1 // 修正：页码从1开始，不是0
	if req.Params.Page != nil {
		page = *req.Params.Page
	}

	limit := 25
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	// 调用 repository 层获取数据
	param := dao.QuestionListParams{
		Page:       page,
		NumPerPage: limit,
		Author:     &userClaims.UserID,
	}
	result, err := question_repo.GetQuestions(ctx, app.DB, param)
	if err != nil {
		return nil, err
	}

	dtos, err := buildQuestionsWith(ctx, app.DB, result)
	if err != nil {
		return nil, err
	}

	return &oapi.GetUserQuestions200JSONResponse{
		Questions: dtos,
		Total:     result.Total,
	}, nil
}
