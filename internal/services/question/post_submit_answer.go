package services

import (
	"context"
	"time"

	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/common"
	question_repo "genshin-quiz/internal/repository/question"
	user_repo "genshin-quiz/internal/repository/user"
	"genshin-quiz/internal/webserver/middleware"

	"github.com/google/uuid"
)

func PostSubmitAnswer(
	ctx context.Context,
	app *config.App,
	req oapi.PostSubmitAnswerRequestObject,
) (*oapi.PostSubmitAnswer200JSONResponse, error) {
	userClaims, ok := middleware.GetUserFromContextOnly(ctx)
	if !ok {
		return nil, common.ErrUserNotInContext
	}

	// 调用仓库层获取问题详情
	questionID, err := question_repo.GetQuestionIDByUUID(ctx, app.DB, req.Id)
	if err != nil {
		return nil, err
	}
	// 获取问题的正确答案
	correctAnswerIDs, err := question_repo.GetQuestionCorrectOptions(ctx, app.DB, *questionID)
	if err != nil {
		return nil, err
	}
	// 获取用户选择的选项ID
	optionIDs, err := question_repo.GetOptionIDsByUUIDs(ctx, app.DB, req.Body.SelectedOptionIds)
	if err != nil {
		return nil, err
	}
	// 比较提交的答案与正确答案
	correct := sliceEqual(optionIDs, *correctAnswerIDs)

	// 检查用户是否已经解决过这道题
	alreadySolved, err := question_repo.CheckQuestionSolved(
		ctx,
		app.DB,
		userClaims.UserID,
		*questionID,
	)
	if err != nil {
		return nil, err
	}

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 记录提交结果
	now := time.Now()
	timeTaken := int32(req.Body.TimeSpent)
	insertData := model.QuestionSubmissions{
		SubmissionUUID: uuid.New(),
		QuestionID:     *questionID,
		UserID:         userClaims.UserID,
		TimeTaken:      &timeTaken,
		IsPractice:     alreadySolved,
		IsCorrect:      correct,
		CreatedAt:      now,
	}
	submissionResult, err := question_repo.InsertSubmission(ctx, tx, insertData)
	if err != nil {
		return nil, err
	}

	// 记录用户选择的选项
	err = question_repo.InsertSubmissionOptions(ctx, tx, submissionResult.ID, optionIDs)
	if err != nil {
		return nil, err
	}

	if !alreadySolved {
		// 更新问题统计
		err = question_repo.UpdateQuestionSolved(ctx, tx, *questionID, correct)
		if err != nil {
			return nil, err
		}
		// 更新选项统计
		err = question_repo.UpdateOptionSelected(ctx, tx, optionIDs)
		if err != nil {
			return nil, err
		}

		// 实时更新用户统计信息
		err = user_repo.UpdateUserSubmissionStats(ctx, tx, userClaims.UserID, correct)
		if err != nil {
			return nil, err
		}
	}
	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &oapi.PostSubmitAnswer200JSONResponse{Correct: correct}, nil
}

func sliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[T]int)
	for _, item := range a {
		m[item]++
	}
	for _, item := range b {
		if m[item] == 0 {
			return false
		}
		m[item]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}
