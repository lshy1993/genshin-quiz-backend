package services

import (
	"context"
	"log"
	"time"

	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
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
		return nil, common.ErrUserNotInContext
	}

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	log.Println("Creating question by user ID:", userClaims.UserID)
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
	createdQuestion, err := question_repo.InsertQuestion(ctx, tx, insertModel)
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
	err = question_repo.InsertQuestionTranslations(ctx, tx, transModels)
	if err != nil {
		return nil, err
	}

	// 选项生成数据
	optionModels := make([]model.QuestionOptions, 0, len(req.Body.Options))
	for _, option := range req.Body.Options {
		isAnswer := option.IsAnswer

		optionModel := model.QuestionOptions{
			QuestionID: createdQuestion.ID,
			OptionUUID: uuid.New(),
			OptionType: model.QuestionOptionType(option.OptionType),
			ImgURL:     option.MediaUrl,
			IsAnswer:   isAnswer,
			CreatedAt:  now,
		}
		optionModels = append(optionModels, optionModel)
	}
	// 插入选项
	insertedOptions, err := question_repo.InsertQuestionOptions(ctx, tx, optionModels)
	if err != nil {
		return nil, err
	}

	// 翻译生成数据
	optionTransModels := make(
		[]model.QuestionOptionTranslations,
		0,
		len(req.Body.Options)*len(req.Body.QuestionText),
	)
	for i, option := range *insertedOptions {
		source := req.Body.Options[i]
		// 为每个选项创建翻译记录
		for lang, text := range *source.Text {
			optionTransModel := model.QuestionOptionTranslations{
				OptionID:   option.ID,
				Language:   lang,
				OptionText: text,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			optionTransModels = append(optionTransModels, optionTransModel)
		}
	}
	// 插入选项翻译
	err = question_repo.InsertOptionTranslations(ctx, tx, optionTransModels)
	if err != nil {
		return nil, err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Convert to API response format
	response := oapi.PostCreateQuestion201JSONResponse{
		Category:     oapi.Category(createdQuestion.Category),
		Difficulty:   oapi.Difficulty(createdQuestion.Difficulty),
		QuestionType: oapi.QuestionType(createdQuestion.QuestionType),
		QuestionText: req.Body.QuestionText,                // 直接使用请求中的翻译数据
		Explanation:  req.Body.Explanation,                 // 直接使用请求中的解释数据
		Options:      []oapi.CreateQuestionOptionRequest{}, // Add options
		Public:       createdQuestion.Public,
	}

	return response, nil
}
