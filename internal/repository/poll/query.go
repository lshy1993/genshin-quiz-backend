package poll_repo

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

func GetPolls(
	ctx context.Context,
	db qrm.DB,
	params dao.PollListParams,
) (*dao.PollListResult, error) {
	tbl := table.Polls
	transTbl := table.PollTranslations
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
			tbl.ID.EQ(transTbl.PollID).AND(transTbl.Language.EQ(pg.String(defaultLang))),
		).
			LEFT_JOIN(userTbl, tbl.CreatedBy.EQ(userTbl.ID)),
	).
		WHERE(
			buildPollCondition(params),
		).
		ORDER_BY(buildPollOrderBy(params)).
		LIMIT(int64(params.Limit)).
		OFFSET(int64(offset))

	// 获取总数
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(tbl).
		WHERE(buildPollCondition(params))

	var countResult struct {
		Count int64 `alias:"count"`
	}
	err := countStmt.QueryContext(ctx, db, &countResult)
	if err != nil {
		return nil, err
	}

	var votes []dao.SimplePoll
	err = stmt.QueryContext(ctx, db, &votes)
	if err != nil {
		return nil, err
	}

	return &dao.PollListResult{
		Polls: votes,
		Total: int(countResult.Count),
	}, nil
}

func buildPollCondition(params dao.PollListParams) pg.BoolExpression {
	tbl := table.Polls
	transTbl := table.PollTranslations
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

func buildPollOrderBy(params dao.PollListParams) pg.OrderByClause {
	tbl := table.Polls
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

func GetPollByUUID(
	ctx context.Context,
	db qrm.DB,
	voteUUID uuid.UUID,
	language *[]string,
) (*dao.DetailedPoll, error) {
	tbl := table.Polls
	transTbl := table.PollTranslations
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
			tbl.ID.EQ(transTbl.PollID).AND(transTbl.Language.EQ(pg.String(defaultLang))),
		).LEFT_JOIN(userTbl, tbl.CreatedBy.EQ(userTbl.ID)),
	).WHERE(
		tbl.PollUUID.EQ(pg.UUID(voteUUID)),
	)

	var result []struct {
		model.Polls
		model.PollTranslations
		model.Users
	}

	err := stmt.QueryContext(ctx, db, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, common.ErrPollNotFound
	}

	detailedPoll := &dao.DetailedPoll{
		Poll:        result[0].Polls,
		User:        result[0].Users,
		Translation: result[0].PollTranslations,
	}

	// 如果没有获取到指定语言的翻译，则 fallback 到任意语言
	if detailedPoll.Translation.Language == "" {
		fallbackTrans, err := getPollTranslationFallback(ctx, db, detailedPoll.Poll.ID)
		if err != nil {
			return nil, err
		}
		if fallbackTrans != nil {
			detailedPoll.Translation = *fallbackTrans
		}
	}

	return detailedPoll, nil
}

// getPollTranslationFallback 获取投票的任意语言翻译（用于 fallback）.
func getPollTranslationFallback(
	ctx context.Context,
	db qrm.DB,
	voteID int64,
) (*model.PollTranslations, error) {
	tbl := table.PollTranslations

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.PollID.EQ(pg.Int64(voteID))).
		LIMIT(1)

	var trans []model.PollTranslations
	err := stmt.QueryContext(ctx, db, &trans)
	if err != nil {
		return nil, err
	}

	if len(trans) == 0 {
		return nil, common.ErrNotFound
	}

	return &trans[0], nil
}

func GetPollOptions(
	ctx context.Context,
	db qrm.DB,
	voteID int64,
) (*[]model.PollOptions, error) {
	tbl := table.PollOptions

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.PollID.EQ(pg.Int64(voteID))).
		ORDER_BY(tbl.OptionOrder.ASC())

	var options []model.PollOptions
	err := stmt.QueryContext(ctx, db, &options)
	if err != nil {
		return nil, err
	}

	return &options, nil
}

func GetPollOptionTranslations(
	ctx context.Context,
	db qrm.DB,
	optionIDs []int64,
	language *[]string,
) (*[]model.PollOptionTranslations, error) {
	if len(optionIDs) == 0 {
		return &[]model.PollOptionTranslations{}, nil
	}

	tbl := table.PollOptionTranslations

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

	var translations []model.PollOptionTranslations
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
	optionTbl := table.PollOptions

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(
			tbl.INNER_JOIN(optionTbl, tbl.OptionID.EQ(optionTbl.ID)),
		).
		WHERE(
			tbl.UserID.EQ(pg.Int64(userID)).AND(
				optionTbl.PollID.EQ(pg.Int64(voteID)),
			),
		)

	var userVotes []model.UserVotes
	err := stmt.QueryContext(ctx, db, &userVotes)
	if err != nil {
		return nil, err
	}

	return &userVotes, nil
}
