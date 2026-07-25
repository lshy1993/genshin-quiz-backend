package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	question_repo "genshin-quiz/internal/repository/question"
	vote_repo "genshin-quiz/internal/repository/vote"
	"genshin-quiz/internal/webserver/middleware"
)

func GetHome(
	ctx context.Context,
	app *config.App,
	req oapi.GetHomeRequestObject,
) (*oapi.GetHome200JSONResponse, error) {
	var language *[]string
	if req.Params.Language != nil {
		lang := []string{*req.Params.Language}
		language = &lang
	}

	questionSortBy := "PublishDate"
	questionResult, err := question_repo.GetQuestions(ctx, app.DB, dao.QuestionListParams{
		Page:       1,
		NumPerPage: 5,
		SortBy:     &questionSortBy,
		SortDesc:   true,
		Language:   language,
	})
	if err != nil {
		return nil, err
	}

	latestQuestions := make([]oapi.Question, 0, len(questionResult.Questions))
	for _, q := range questionResult.Questions {
		latestQuestions = append(latestQuestions, transformer.ConvertSimpleToQuestion(q, false, 0))
	}

	voteSortBy := "created_at"
	voteResult, err := vote_repo.GetVotes(ctx, app.DB, dao.VoteListParams{
		Page:     1,
		Limit:    5,
		Type:     "all",
		SortBy:   voteSortBy,
		SortDesc: true,
		Language: language,
	})
	if err != nil {
		return nil, err
	}

	var userClaims *middleware.UserClaims
	if claims, ok := middleware.GetUserFromContextOnly(ctx); ok {
		userClaims = claims
	}

	latestVotes := make([]oapi.Poll, 0, len(voteResult.Votes))
	popularVotes := make([]oapi.Poll, 0, len(voteResult.Votes))
	for _, vote := range voteResult.Votes {
		voted := false
		likeStatus := int16(0)
		if userClaims != nil {
			userVotes, err := vote_repo.GetUserVotedOptions(
				ctx,
				app.DB,
				userClaims.UserID,
				vote.Vote.ID,
			)
			if err == nil && userVotes != nil && len(*userVotes) > 0 {
				voted = true
			}

			userLikeStatus, err := vote_repo.GetVoteLikeStatus(
				ctx,
				app.DB,
				userClaims.UserID,
				vote.Vote.ID,
			)
			if err == nil && userLikeStatus != nil {
				likeStatus = *userLikeStatus
			}
		}

		dto := transformer.ConvertSimpleVoteToDTO(vote, voted, likeStatus)
		latestVotes = append(latestVotes, dto)
		popularVotes = append(popularVotes, dto)
	}

	return &oapi.GetHome200JSONResponse{
		PopularExams:    []oapi.Exam{},
		LatestQuestions: latestQuestions,
		LatestPolls:     latestVotes,
		PopularPolls:    popularVotes,
	}, nil
}
