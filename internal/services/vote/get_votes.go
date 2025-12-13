package services

import (
	"context"

	"genshin-quiz/config"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	vote_repo "genshin-quiz/internal/repository/vote"
)

func GetVotes(
	ctx context.Context,
	app *config.App,
	req oapi.GetVotesRequestObject,
) (*oapi.GetVotes200JSONResponse, error) {
	// 设置默认值
	page := 1
	if req.Params.Page != nil {
		page = *req.Params.Page
	}

	limit := 25
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	voteType := "all"
	if req.Params.Type != nil {
		voteType = string(*req.Params.Type)
	}

	sortBy := "created_at"
	if req.Params.SortBy != nil {
		sortBy = *req.Params.SortBy
	}

	sortDesc := false
	if req.Params.SortDesc != nil {
		sortDesc = *req.Params.SortDesc
	}

	// 调用 repository 层获取数据
	result, err := vote_repo.GetVotes(
		ctx,
		app.DB,
		dao.VoteListParams{
			Page:     page,
			Limit:    limit,
			Type:     voteType,
			Query:    req.Params.Query,
			Language: req.Params.Language,
			SortBy:   sortBy,
			SortDesc: sortDesc,
		},
	)
	if err != nil {
		return nil, err
	}

	// 转换为 DTO
	dtos := make([]oapi.Vote, 0, len(result.Votes))
	for _, vote := range result.Votes {
		dto := transformer.ConvertSimpleVoteToDTO(vote)
		dtos = append(dtos, dto)
	}

	return &oapi.GetVotes200JSONResponse{
		Total: result.Total,
		Votes: dtos,
	}, nil
}
