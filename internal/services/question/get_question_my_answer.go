package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	question_repo "genshin-quiz/internal/repository/question"
)

func GetQuestionMySubmissions(
	ctx context.Context,
	app *config.App,
	req oapi.GetQuestionMySubmissionsRequestObject,
) (*[]oapi.MySubmission, error) {
	submissions, err := question_repo.GetQuestionSubmissions(ctx, app.DB, req.Id)
	if err != nil {
		return nil, err
	}
	if len(*submissions) == 0 {
		return &[]oapi.MySubmission{}, nil
	}

	submissionIDs := make([]int64, 0, len(*submissions))
	for _, submission := range *submissions {
		submissionIDs = append(submissionIDs, submission.ID)
	}
	// 使用单一查询获取提交列表（包含选项信息）
	submissionMap, err := question_repo.GetQuestionSubmissionsWithOptions(
		ctx,
		app.DB,
		submissionIDs,
	)
	if err != nil {
		return nil, err
	}

	dtos := make([]oapi.MySubmission, 0, len(*submissions))
	for _, submission := range *submissions {
		timeSpent := 0
		if submission.TimeTaken != nil {
			timeSpent = int(*submission.TimeTaken)
		}
		dto := oapi.MySubmission{
			IsCorrect:         submission.IsCorrect,
			SelectedOptionIds: (*submissionMap)[submission.ID],
			SubmittedAt:       submission.CreatedAt,
			TimeSpent:         timeSpent,
		}
		dtos = append(dtos, dto)
	}

	return &dtos, nil
}
