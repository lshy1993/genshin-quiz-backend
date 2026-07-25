package services

import (
	"context"
	"time"

	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	vote_repo "genshin-quiz/internal/repository/vote"
	"genshin-quiz/internal/webserver/middleware"

	"github.com/google/uuid"
)

func PostCreatePoll(
	ctx context.Context,
	app *config.App,
	req oapi.PostCreatePollRequestObject,
) (oapi.PostCreatePollResponseObject, error) {
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
	insertModel := genInsertModel(req, userClaims.UserID, now)
	createdVote, err := vote_repo.InsertVote(ctx, tx, insertModel)
	if err != nil {
		return nil, err
	}

	// 批量插入翻译数据
	transModels := genTranslationModels(req, createdVote.ID, now)
	err = vote_repo.InsertVoteTranslations(ctx, tx, transModels)
	if err != nil {
		return nil, err
	}

	// 插入选项
	optionModels := genOptionModels(req, createdVote.ID, now)
	insertedOptions, err := vote_repo.InsertVoteOptions(ctx, tx, optionModels)
	if err != nil {
		return nil, err
	}

	// 插入选项翻译
	optionTransModels := genOptionTranslationModels(*insertedOptions, req, now)
	err = vote_repo.InsertOptionTranslations(ctx, tx, optionTransModels)
	if err != nil {
		return nil, err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Convert to API response format
	response := oapi.PostCreatePoll201JSONResponse{
		Category:    oapi.Category(createdVote.Category),
		Public:      createdVote.Public,
		Title:       req.Body.Title,       // 直接使用请求中的翻译数据
		Description: req.Body.Description, // 直接使用请求中的解释数据
		Options:     []oapi.PollOption{},  // Add options
	}

	return response, nil
}

func genInsertModel(
	req oapi.PostCreatePollRequestObject,
	userID int64,
	now time.Time,
) model.Votes {
	var expiredTime *time.Time
	if req.Body.ExpireAt != nil {
		expiredTime = req.Body.ExpireAt
	}
	votesPerOption := int32(req.Body.VotesPerOption)

	// 投票主体
	insertModel := model.Votes{
		VoteUUID:          uuid.New(),
		Public:            req.Body.Public,
		Category:          model.Category(req.Body.Category),
		StartAt:           req.Body.StartAt,
		ExpiresAt:         expiredTime,
		VotesPerUser:      int32(req.Body.VotesPerUser),
		VotesPerOption:    &votesPerOption,
		CreatedBy:         userID, // 使用从 JWT 获取的用户 ID
		CreatedAt:         now,
		LikesCount:        0, // 初始化点赞数为 0
		ParticipantsCount: 0, // 初始化参与者数为 0
		TotalVotesCount:   0, // 初始化总投票数为 0
	}
	return insertModel
}

func genTranslationModels(
	req oapi.PostCreatePollRequestObject,
	voteID int64,
	now time.Time,
) []model.VoteTranslations {
	// 投票主体的翻译 - 预先合并 title 和 description
	type TranslationData struct {
		Title       string
		Description *string
	}
	translationMap := make(map[string]TranslationData)
	for lang, title := range req.Body.Title {
		data := TranslationData{Title: title}
		if req.Body.Description != nil {
			if desc, exists := (*req.Body.Description)[lang]; exists {
				data.Description = &desc
			}
		}
		translationMap[lang] = data
	}

	transModels := make([]model.VoteTranslations, 0, len(translationMap))
	for lang, data := range translationMap {
		transModel := model.VoteTranslations{
			VoteID:      voteID,
			Language:    lang,
			Title:       data.Title,
			Description: data.Description,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		transModels = append(transModels, transModel)
	}
	return transModels
}

func genOptionModels(
	req oapi.PostCreatePollRequestObject,
	voteID int64,
	now time.Time,
) []model.VoteOptions {
	// 生成投票项数据
	optionModels := make([]model.VoteOptions, 0, len(req.Body.Options))
	for index := range req.Body.Options {
		optionModel := model.VoteOptions{
			VoteID:      voteID,
			OptionUUID:  uuid.New(),
			OptionOrder: int32(index),
			CreatedAt:   now,
		}
		optionModels = append(optionModels, optionModel)
	}
	return optionModels
}

func genOptionTranslationModels(
	insertedOptions []model.VoteOptions,
	req oapi.PostCreatePollRequestObject,
	now time.Time,
) []model.VoteOptionTranslations {
	// 生成选项的翻译数据
	optionTransModels := make(
		[]model.VoteOptionTranslations,
		0,
		len(req.Body.Options)*len(req.Body.Title),
	)
	for i, option := range insertedOptions {
		source := req.Body.Options[i]
		// 为每个选项创建翻译记录
		for lang, text := range source.Text {
			optionTransModel := model.VoteOptionTranslations{
				OptionID:   option.ID,
				Language:   lang,
				OptionText: text,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			optionTransModels = append(optionTransModels, optionTransModel)
		}
	}
	return optionTransModels
}
