package services

import (
	"context"
	"log"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	dao "genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	question_repo "genshin-quiz/internal/repository/question"
	"genshin-quiz/internal/webserver/middleware"
)

func GetQuestions(
	ctx context.Context,
	app *config.App,
	req oapi.GetQuestionsRequestObject,
) (*oapi.GetQuestions200JSONResponse, error) {
	var page int
	if req.Params.Page != nil {
		page = *req.Params.Page
	} else {
		page = 1 // 修正：页码从1开始，不是0
	}
	var limit int
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	} else {
		limit = 25
	}

	// 添加参数验证
	if page <= 0 {
		page = 1
	}
	if limit <= 1 || limit > 100 {
		limit = 25
	}

	sortDesc := false
	if req.Params.SortDesc != nil {
		sortDesc = *req.Params.SortDesc
	}

	param := dao.QuestionListParams{
		Page:       page,
		NumPerPage: limit,
		Category:   req.Params.Category,
		Difficulty: req.Params.Difficulty,
		Query:      req.Params.Query,
		SortBy:     req.Params.SortBy,
		SortDesc:   sortDesc,
		Language:   req.Params.Language,
	}
	result, err := question_repo.GetQuestions(ctx, app.DB, param)
	if err != nil {
		return nil, err
	}

	// 检查用户是否已解答这些题目（如果用户已登录）
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	var solvedMap map[int64]bool
	if ok {
		questionIDs := make([]int64, 0, len(result.Questions))
		for _, q := range result.Questions {
			questionIDs = append(questionIDs, q.Question.ID)
		}

		solvedMap, err = question_repo.CheckMultipleQuestionsSolved(
			ctx,
			app.DB,
			userClaims.UserID,
			questionIDs,
		)
		if err != nil {
			return nil, err
		}

		log.Println("Solved Map:", solvedMap)
	}

	dtos := make([]oapi.Question, 0, len(result.Questions))
	for _, q := range result.Questions {
		solved := false
		if val, exists := solvedMap[q.Question.ID]; exists {
			solved = val
		}
		dtos = append(dtos, transformer.ConvertSimpleToQuestion(q, solved, 0))
	}

	return &oapi.GetQuestions200JSONResponse{
		Questions: dtos,
		Total:     result.Total,
	}, nil
}
