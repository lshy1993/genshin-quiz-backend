package vote_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"
	"genshin-quiz/internal/dao"

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

	offset := (params.Page - 1) * params.Limit
	if offset < 0 {
		offset = 0
	}

	// 基础查询 - 使用语言过滤的 JOIN
	defaultLang := "zh"
	if params.Language != nil && len(*params.Language) > 0 {
		defaultLang = (*params.Language)[0]
	}

	// 构建 WHERE 条件
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
	}

	// 关键字搜索（在标题中搜索）
	if params.Query != nil && *params.Query != "" {
		searchTerm := "%" + *params.Query + "%"
		condition = condition.AND(pg.LOWER(transTbl.Title).LIKE(pg.String(searchTerm)))
	}

	// 主查询
	userTbl := table.Users
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
		condition,
	).ORDER_BY(
		buildVoteOrderBy(params.SortBy, params.SortDesc),
	).LIMIT(int64(params.Limit)).OFFSET(int64(offset))

	// 获取总数
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(
			tbl.LEFT_JOIN(
				transTbl,
				tbl.ID.EQ(transTbl.VoteID).AND(transTbl.Language.EQ(pg.String(defaultLang))),
			),
		).
		WHERE(condition)

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

func buildVoteOrderBy(sortBy string, sortDesc bool) pg.OrderByClause {
	var orderExpr pg.Expression

	switch sortBy {
	case "start_at":
		orderExpr = table.Votes.StartAt
	case "expires_at":
		orderExpr = table.Votes.ExpiresAt
	case "participants":
		orderExpr = table.Votes.ParticipantsCount
	case "votes":
		orderExpr = table.Votes.TotalVotesCount
	case "likes":
		orderExpr = table.Votes.LikesCount
	default: // created_at
		orderExpr = table.Votes.CreatedAt
	}

	if sortDesc {
		return orderExpr.DESC()
	}
	return orderExpr.ASC()
}

func GetVoteByUUID(ctx context.Context, db qrm.DB, uuid uuid.UUID) (*model.Votes, error) {
	tbl := table.Votes
	stmt := pg.SELECT(tbl.AllColumns).FROM(
		tbl,
	).WHERE(
		tbl.VoteUUID.EQ(pg.UUID(uuid)),
	)

	var vote model.Votes
	err := stmt.QueryContext(ctx, db, &vote)
	if err != nil {
		return nil, err
	}

	return &vote, nil
}
