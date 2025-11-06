package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	question_repo "genshin-quiz/internal/repository/question"
)

func GetQuestionRecentSubmissions(
	ctx context.Context,
	app *config.App,
	req oapi.GetQuestionRecentSubmissionsRequestObject,
) (*[]oapi.RecentSubmission, error) {
	submissions, err := question_repo.GetQuestionSubmissions(ctx, app.DB, req.Id)
	if err != nil {
		return nil, err
	}
	if len(*submissions) == 0 {
		return &[]oapi.RecentSubmission{}, nil
	}

	submissionIDs := make([]int64, 0, len(*submissions))
	for _, submission := range *submissions {
		submissionIDs = append(submissionIDs, submission.ID)
	}

	dtos := make([]oapi.RecentSubmission, 0, len(*submissions))
	for _, submission := range *submissions {
		timeSpent := 0
		if submission.TimeTaken != nil {
			timeSpent = int(*submission.TimeTaken)
		}
		dto := oapi.RecentSubmission{
			IsCorrect:   submission.IsCorrect,
			SubmittedAt: submission.CreatedAt,
			TimeSpent:   timeSpent,
		}
		dtos = append(dtos, dto)
	}

	return &dtos, nil
}
