package vote_repo

import (
	"context"
	"strings"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"
	"genshin-quiz/internal/common"
	"genshin-quiz/internal/dao"
	"genshin-quiz/internal/util"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

func GetVotes(
	ctx context.Context,
	db qrm.DB,
	params dao.VoteListParams,
) (*dao.VoteListResult, error) {
	tbl := table.Votes
	transTbl := table.VoteTranslations
	userTbl := table.Users

	offset := (params.Page - 1) * params.Limit
	if offset < 0 {
		offset = 0
	}

	// 获取默认语言
	defaultLang := util.GetDefaultLanguage(params.Language)

	// 主查询
	stmt := pg.SELECT(
		tbl.AllColumns,
		transTbl.AllColumns,
		userTbl.AllColumns,
	).FROM(
		tbl.LEFT_JOIN(
			transTbl,
			tbl.ID.EQ(transTbl.VoteID).AND(transTbl.Language.EQ(pg.String(defaultLang))),
		).
			LEFT_JOIN(userTbl, tbl.CreatedBy.EQ(userTbl.ID)),
	).
		WHERE(
			buildVoteCondition(params),
		).
		ORDER_BY(buildVoteOrderBy(params)).
		LIMIT(int64(params.Limit)).
		OFFSET(int64(offset))

	// 获取总数
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(tbl).
		WHERE(buildVoteCondition(params))

	var countResult struct {
		Count int64 `alias:"count"`
	}
	err := countStmt.QueryContext(ctx, db, &countResult)
	if err != nil {
		return nil, err
	}

	var votes []dao.SimpleVote
	err = stmt.QueryContext(ctx, db, &votes)
	if err != nil {
		return nil, err
	}

	return &dao.VoteListResult{
		Votes: votes,
		Total: int(countResult.Count),
	}, nil
}

func buildVoteCondition(params dao.VoteListParams) pg.BoolExpression {
	tbl := table.Votes
	transTbl := table.VoteTranslations
	condition := tbl.Public.IS_TRUE()

	// 类型筛选：all, available, expired
	now := pg.CURRENT_TIMESTAMP()
	switch params.Type {
	case "available":
		condition = condition.AND(
			tbl.StartAt.LT_EQ(now).AND(
				tbl.ExpiresAt.IS_NULL().OR(tbl.ExpiresAt.GT(now)),
			),
		)
	case "expired":
		condition = condition.AND(
			tbl.ExpiresAt.IS_NOT_NULL().AND(tbl.ExpiresAt.LT_EQ(now)),
		)
		// case "all" 或默认不添加时间过滤
	}

	// 关键字搜索（在标题中搜索）
	if params.Query != nil && *params.Query != "" {
		searchTerm := "%" + strings.ToLower(*params.Query) + "%"
		condition = condition.AND(pg.LOWER(transTbl.Title).LIKE(pg.String(searchTerm)))
	}

	return condition
}

func buildVoteOrderBy(params dao.VoteListParams) pg.OrderByClause {
	tbl := table.Votes
	var orderExpr pg.Expression

	switch params.SortBy {
	case "start_at":
		orderExpr = tbl.StartAt
	case "expires_at":
		orderExpr = tbl.ExpiresAt
	case "participants":
		orderExpr = tbl.ParticipantsCount
	case "votes":
		orderExpr = tbl.TotalVotesCount
	case "likes":
		orderExpr = tbl.LikesCount
	default: // created_at
		orderExpr = tbl.CreatedAt
	}

	if params.SortDesc {
		return orderExpr.DESC()
	}
	return orderExpr.ASC()
}

func GetVoteByUUID(
	ctx context.Context,
	db qrm.DB,
	voteUUID uuid.UUID,
	language *[]string,
) (*dao.DetailedVote, error) {
	tbl := table.Votes
	transTbl := table.VoteTranslations
	userTbl := table.Users

	// 获取默认语言
	defaultLang := util.GetDefaultLanguage(language)

	stmt := pg.SELECT(
		tbl.AllColumns,
		transTbl.AllColumns,
		userTbl.AllColumns,
	).FROM(
		tbl.LEFT_JOIN(
			transTbl,
			tbl.ID.EQ(transTbl.VoteID).AND(transTbl.Language.EQ(pg.String(defaultLang))),
		).LEFT_JOIN(userTbl, tbl.CreatedBy.EQ(userTbl.ID)),
	).WHERE(
		tbl.VoteUUID.EQ(pg.UUID(voteUUID)),
	)

	var result []struct {
		model.Votes
		model.VoteTranslations
		model.Users
	}

	err := stmt.QueryContext(ctx, db, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, common.ErrVoteNotFound
	}

	detailedVote := &dao.DetailedVote{
		Vote:        result[0].Votes,
		User:        result[0].Users,
		Translation: result[0].VoteTranslations,
	}

	// 如果没有获取到指定语言的翻译，则 fallback 到任意语言
	if detailedVote.Translation.Language == "" {
		fallbackTrans, err := getVoteTranslationFallback(ctx, db, detailedVote.Vote.ID)
		if err != nil {
			return nil, err
		}
		if fallbackTrans != nil {
			detailedVote.Translation = *fallbackTrans
		}
	}

	return detailedVote, nil
}

// getVoteTranslationFallback 获取投票的任意语言翻译（用于 fallback）.
func getVoteTranslationFallback(
	ctx context.Context,
	db qrm.DB,
	voteID int64,
) (*model.VoteTranslations, error) {
	tbl := table.VoteTranslations

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.VoteID.EQ(pg.Int64(voteID))).
		LIMIT(1)

	var trans []model.VoteTranslations
	err := stmt.QueryContext(ctx, db, &trans)
	if err != nil {
		return nil, err
	}

	if len(trans) == 0 {
		return nil, nil
	}

	return &trans[0], nil
}

func GetVoteOptions(
	ctx context.Context,
	db qrm.DB,
	voteID int64,
) (*[]model.VoteOptions, error) {
	tbl := table.VoteOptions

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.VoteID.EQ(pg.Int64(voteID))).
		ORDER_BY(tbl.OptionOrder.ASC())

	var options []model.VoteOptions
	err := stmt.QueryContext(ctx, db, &options)
	if err != nil {
		return nil, err
	}

	return &options, nil
}

func GetVoteOptionTranslations(
	ctx context.Context,
	db qrm.DB,
	optionIDs []int64,
	language *[]string,
) (*[]model.VoteOptionTranslations, error) {
	if len(optionIDs) == 0 {
		return &[]model.VoteOptionTranslations{}, nil
	}

	tbl := table.VoteOptionTranslations

	// 构建选项 ID 表达式列表
	optionIDExpressions := util.BuildInt64Expressions(optionIDs)

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.OptionID.IN(optionIDExpressions...))

	// 如果指定了语言，则过滤
	if language != nil && len(*language) > 0 {
		langExpressions := util.BuildStringExpressions(*language)
		stmt = stmt.WHERE(tbl.Language.IN(langExpressions...))
	}

	var translations []model.VoteOptionTranslations
	err := stmt.QueryContext(ctx, db, &translations)
	if err != nil {
		return nil, err
	}

	return &translations, nil
}

func GetUserVotedOptions(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	voteID int64,
) (*[]model.UserVotes, error) {
	tbl := table.UserVotes
	optionTbl := table.VoteOptions

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(
			tbl.INNER_JOIN(optionTbl, tbl.OptionID.EQ(optionTbl.ID)),
		).
		WHERE(
			tbl.UserID.EQ(pg.Int64(userID)).AND(
				optionTbl.VoteID.EQ(pg.Int64(voteID)),
			),
		)

	var userVotes []model.UserVotes
	err := stmt.QueryContext(ctx, db, &userVotes)
	if err != nil {
		return nil, err
	}

	return &userVotes, nil
}
