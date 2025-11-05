package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	question_repo "genshin-quiz/internal/repository/question"

	"github.com/google/uuid"
)

func GetQuestionSubmissions(
	ctx context.Context,
	app *config.App,
	req oapi.GetQuestionSubmissionsRequestObject,
) (*[]oapi.Submission, error) {
	// 获取提交列表
	submissions, err := question_repo.GetQuestionSubmissions(ctx, app.DB, req.Id, nil)
	if err != nil {
		return nil, err
	}

	dtos := make([]oapi.Submission, 0, len(*submissions))
	for _, submission := range *submissions {
		timeSpent := 0
		if submission.TimeTaken != nil {
			timeSpent = int(*submission.TimeTaken)
		}
		dto := oapi.Submission{
			Id:                submission.SubmissionUUID,
			IsCorrect:         submission.IsCorrect,
			SelectedOptionIds: []uuid.UUID{}, // todo: 填充选项ID
			SubmittedAt:       submission.CreatedAt,
			TimeSpent:         timeSpent,
		}
		dtos = append(dtos, dto)
	}

	return &dtos, nil
}
