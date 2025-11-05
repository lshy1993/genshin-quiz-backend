package services

import (
	"context"
	"errors"
	"time"

	"genshin-quiz/config"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
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
		return nil, errors.New("user not found in context")
	}

	// 调用仓库层获取问题详情
	questionID, err := question_repo.GetQuestionIDByUUID(ctx, app.DB, req.Id)
	if err != nil {
		return nil, err
	}
	// 获取问题的正确答案
	correctAnswerUUIDs, err := question_repo.GetQuestionCorrectOptions(ctx, app.DB, *questionID)
	if err != nil {
		return nil, err
	}

	// 比较提交的答案与正确答案
	optionUUIDs := req.Body.SelectedOptionIds
	correct := uuidSliceEqual(optionUUIDs, *correctAnswerUUIDs)

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 更新问题统计
	err = question_repo.UpdateQuestionSolved(ctx, tx, *questionID, correct)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 记录提交结果
	now := time.Now()
	timeTaken := int32(req.Body.TimeSpent)
	insertData := model.QuestionSubmissions{
		SubmissionUUID: uuid.New(),
		QuestionID:     *questionID,
		UserID:         userClaims.UserID,
		TimeTaken:      &timeTaken,
		IsPractice:     false,
		IsCorrect:      correct,
		CreatedAt:      now,
		// SelectedOptionIDs: optionIDs,
	}
	err = question_repo.InsertSubmission(ctx, tx, insertData)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 实时更新用户统计信息
	err = user_repo.UpdateUserSubmissionStats(ctx, tx, userClaims.UserID, correct)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &oapi.PostSubmitAnswer200JSONResponse{Correct: &correct}, nil
}

func uuidSliceEqual(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[uuid.UUID]int)
	for _, id := range a {
		m[id]++
	}
	for _, id := range b {
		if m[id] == 0 {
			return false
		}
		m[id]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}
