package services

import (
	"context"
	"fmt"
	"time"

	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	question_repo "genshin-quiz/internal/repository/question"
	"genshin-quiz/internal/webserver/middleware"

	"github.com/google/uuid"
)

func PostCreateQuestion(
	ctx context.Context,
	app *config.App,
	req oapi.PostCreateQuestionRequestObject,
) (oapi.PostCreateQuestionResponseObject, error) {
	// 从 context 中获取用户信息
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}

	now := time.Now()
	// 提问表主体
	insertModel := model.Questions{
		QuestionUUID: uuid.New(),
		Public:       true,
		QuestionType: model.QuestionType(req.Body.QuestionType),
		Category:     model.Category(req.Body.Category),
		Difficulty:   model.Difficulty(req.Body.Difficulty),
		IsPublished:  true,
		PublishedAt:  &now,
		CreatedBy:    userClaims.UserID, // 使用从 JWT 获取的用户 ID
		CreatedAt:    now,
	}
	createdQuestion, err := question_repo.InsertQuestion(ctx, app.DB, insertModel)
	if err != nil {
		return nil, err
	}
	// 提问翻译
	transModels := make([]model.QuestionTranslations, 0, len(req.Body.QuestionText))
	for lang, text := range req.Body.QuestionText {
		transModel := model.QuestionTranslations{
			QuestionID:   createdQuestion.ID,
			Language:     lang,
			QuestionText: text,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		// 如果有对应语言的解释，添加到翻译记录中
		if req.Body.Explanation != nil {
			if explanation, exists := (*req.Body.Explanation)[lang]; exists {
				transModel.Explanation = &explanation
			}
		}

		transModels = append(transModels, transModel)
	}

	// 批量插入翻译数据
	err = question_repo.InsertQuestionTranslations(ctx, app.DB, transModels)
	if err != nil {
		return nil, err
	}

	// 插入选项
	optionModels := make([]model.QuestionOptions, 0, len(req.Body.Options))
	// 插入翻译
	optionTransModels := make(
		[]model.OptionTranslations,
		0,
		len(req.Body.Options)*len(req.Body.QuestionText),
	)
	for _, option := range req.Body.Options {
		optionModel := model.QuestionOptions{
			QuestionID: createdQuestion.ID,
			OptionUUID: uuid.New(),
			IsAnswer:   *option.IsAnswer,
			CreatedAt:  now,
		}
		optionModels = append(optionModels, optionModel)

		// 为每个选项创建翻译记录
		for lang, text := range *option.Text {
			optionTransModel := model.OptionTranslations{
				OptionID:   optionModel.ID,
				Language:   lang,
				OptionText: text,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			optionTransModels = append(optionTransModels, optionTransModel)
		}
	}

	err = question_repo.InsertQuestionOptions(ctx, app.DB, optionModels)
	if err != nil {
		return nil, err
	}
	err = question_repo.InsertOptionTranslations(ctx, app.DB, optionTransModels)
	if err != nil {
		return nil, err
	}

	// Convert to API response format
	response := oapi.PostCreateQuestion201JSONResponse{
		Category:     oapi.QuestionCategory(createdQuestion.Category),
		Difficulty:   oapi.QuestionDifficulty(createdQuestion.Difficulty),
		QuestionType: oapi.QuestionType(createdQuestion.QuestionType),
		QuestionText: req.Body.QuestionText,   // 直接使用请求中的翻译数据
		Explanation:  req.Body.Explanation,    // 直接使用请求中的解释数据
		Options:      []oapi.QuestionOption{}, // Add options
		Public:       createdQuestion.Public,
	}

	return response, nil
}
