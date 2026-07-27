package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	dao "genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	question_repo "genshin-quiz/internal/repository/question"
	"genshin-quiz/internal/webserver/middleware"

	"github.com/go-jet/jet/v2/qrm"
)

func GetQuestions(
	ctx context.Context,
	app *config.App,
	req oapi.GetQuestionsRequestObject,
) (*oapi.GetQuestions200JSONResponse, error) {
	page := 1 // 修正：页码从1开始，不是0
	if req.Params.Page != nil {
		page = *req.Params.Page
	}

	limit := 25
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	sortBy := "PublishDate"
	if req.Params.SortBy != nil {
		sortBy = *req.Params.SortBy
	}
	sortDesc := false
	if req.Params.SortDesc != nil {
		sortDesc = *req.Params.SortDesc
	}

	// 调用 repository 层获取数据
	param := dao.QuestionListParams{
		Page:       page,
		NumPerPage: limit,
		Category:   req.Params.Category,
		Difficulty: req.Params.Difficulty,
		Query:      req.Params.Query,
		Language:   req.Params.Language,
		SortBy:     sortBy,
		SortDesc:   sortDesc,
	}
	result, err := question_repo.GetQuestions(ctx, app.DB, param)
	if err != nil {
		return nil, err
	}

	dtos, err := buildQuestionsWith(ctx, app.DB, result)
	if err != nil {
		return nil, err
	}

	return &oapi.GetQuestions200JSONResponse{
		Questions: dtos,
		Total:     result.Total,
	}, nil
}

func buildQuestionsWith(
	ctx context.Context,
	db qrm.DB,
	result *dao.QuestionListResult,
) ([]oapi.Question, error) {
	// 检查用户是否已解答这些题目（如果用户已登录）
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)

	var solvedMap map[int64]bool
	var err error
	if ok {
		questionIDs := make([]int64, 0, len(result.Questions))
		for _, q := range result.Questions {
			questionIDs = append(questionIDs, q.Question.ID)
		}

		solvedMap, err = question_repo.CheckMultipleQuestionsSolved(
			ctx,
			db,
			userClaims.UserID,
			questionIDs,
		)
		if err != nil {
			return nil, err
		}

		// log.Println("Solved Map:", solvedMap)
	}

	dtos := make([]oapi.Question, 0, len(result.Questions))
	for _, q := range result.Questions {
		dtos = append(dtos, transformer.ConvertSimpleToQuestion(q, solvedMap[q.Question.ID], 0))
	}

	return dtos, nil
}
