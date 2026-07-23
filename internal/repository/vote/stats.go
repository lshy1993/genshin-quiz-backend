package vote_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

/*
RecalculateAllVoteLikeStats 根据 vote_likes 表统计每个投票的点赞数，并更新 votes.likes_count 字段.
只统计 value=1 的点赞数.
*/
func RecalculateAllVoteLikeStats(
	ctx context.Context,
	db qrm.DB,
) error {
	voteTbl := table.Votes
	likesTbl := table.VoteLikes

	// 统计每个投票的点赞数
	stmt := pg.SELECT(
		likesTbl.VoteID,
		pg.COUNT(likesTbl.Value).AS("like_count"),
	).FROM(
		likesTbl,
	).WHERE(
		likesTbl.Value.EQ(pg.Int16(1)),
	).GROUP_BY(
		likesTbl.VoteID,
	)

	var results []struct {
		VoteID    int64 `alias:"vote_likes.vote_id"`
		LikeCount int64 `alias:"like_count"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return err
	}

	// 构建 voteID -> likeCount 映射
	likeCountMap := make(map[int64]int64, len(results))
	for _, r := range results {
		likeCountMap[r.VoteID] = r.LikeCount
	}

	// 对每个投票执行更新
	for voteID, likeCount := range likeCountMap {
		updateStmt := voteTbl.UPDATE().SET(
			voteTbl.LikesCount.SET(pg.Int64(likeCount)),
		).WHERE(
			voteTbl.ID.EQ(pg.Int64(voteID)),
		)
		_, err := updateStmt.ExecContext(ctx, db)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
RecalculateAllVoteOptionStats 根据 user_votes 表统计每个投票选项的票数，并更新 vote_options.vote_count 字段.
同时统计整个投票的总票数和参与人数.
*/
func RecalculateAllVoteOptionStats(
	ctx context.Context,
	db qrm.DB,
) error {
	voteOptionTbl := table.VoteOptions
	userVotesTbl := table.UserVotes

	// 统计每个选项的总票数
	stmt := pg.SELECT(
		userVotesTbl.OptionID,
		pg.SUM(pg.CAST(userVotesTbl.VoteCount).AS("BIGINT")).AS("total_votes"),
	).FROM(
		userVotesTbl,
	).GROUP_BY(
		userVotesTbl.OptionID,
	)

	var results []struct {
		OptionID   int64  `alias:"user_votes.option_id"`
		TotalVotes *int64 `alias:"total_votes"` // 使用指针处理可能的 NULL
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return err
	}

	// 对每个选项执行更新
	for _, r := range results {
		voteCount := int64(0)
		if r.TotalVotes != nil {
			voteCount = *r.TotalVotes
		}

		updateStmt := voteOptionTbl.UPDATE().SET(
			voteOptionTbl.VoteCount.SET(pg.Int64(voteCount)),
		).WHERE(
			voteOptionTbl.ID.EQ(pg.Int64(r.OptionID)),
		)
		_, err := updateStmt.ExecContext(ctx, db)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
RecalculateVoteStats 根据 user_votes 表统计指定投票的参与人数和总票数，并更新 votes 表.
参与人数 = 该投票下不同用户的数量.
总票数 = 该投票下所有用户投票的总和.
*/
func RecalculateVoteStats(
	ctx context.Context,
	db qrm.DB,
	voteID int64,
) error {
	voteTbl := table.Votes
	userVotesTbl := table.UserVotes
	voteOptionTbl := table.VoteOptions

	// 统计参与人数（不同用户的数量）
	participantsStmt := pg.SELECT(
		pg.COUNT(pg.DISTINCT(userVotesTbl.UserID)).AS("participants_count"),
	).FROM(
		userVotesTbl.INNER_JOIN(voteOptionTbl, voteOptionTbl.ID.EQ(userVotesTbl.OptionID)),
	).WHERE(
		voteOptionTbl.VoteID.EQ(pg.Int64(voteID)),
	)

	var participantsResult struct {
		ParticipantsCount int64 `alias:"participants_count"`
	}
	err := participantsStmt.QueryContext(ctx, db, &participantsResult)
	if err != nil {
		return err
	}

	// 统计总票数（所有投票的总和）
	totalVotesStmt := pg.SELECT(
		pg.SUM(pg.CAST(userVotesTbl.VoteCount).AS("BIGINT")).AS("total_votes"),
	).FROM(
		userVotesTbl.INNER_JOIN(voteOptionTbl, voteOptionTbl.ID.EQ(userVotesTbl.OptionID)),
	).WHERE(
		voteOptionTbl.VoteID.EQ(pg.Int64(voteID)),
	)

	var totalVotesResult struct {
		TotalVotes *int64 `alias:"total_votes"`
	}
	err = totalVotesStmt.QueryContext(ctx, db, &totalVotesResult)
	if err != nil {
		return err
	}

	totalVotes := int64(0)
	if totalVotesResult.TotalVotes != nil {
		totalVotes = *totalVotesResult.TotalVotes
	}

	// 更新投票的统计数据
	updateStmt := voteTbl.UPDATE().SET(
		voteTbl.ParticipantsCount.SET(pg.Int64(participantsResult.ParticipantsCount)),
		voteTbl.TotalVotesCount.SET(pg.Int64(totalVotes)),
	).WHERE(
		voteTbl.ID.EQ(pg.Int64(voteID)),
	)

	_, err = updateStmt.ExecContext(ctx, db)
	return err
}

/*
RecalculateAllVoteStats 重新计算所有投票的统计数据.
包括：参与人数、总票数.
*/
func RecalculateAllVoteStats(
	ctx context.Context,
	db qrm.DB,
) error {
	voteTbl := table.Votes
	userVotesTbl := table.UserVotes
	voteOptionTbl := table.VoteOptions

	// 统计每个投票的参与人数和总票数
	stmt := pg.SELECT(
		voteOptionTbl.VoteID,
		pg.COUNT(pg.DISTINCT(userVotesTbl.UserID)).AS("participants_count"),
		pg.SUM(pg.CAST(userVotesTbl.VoteCount).AS("BIGINT")).AS("total_votes"),
	).FROM(
		userVotesTbl.INNER_JOIN(voteOptionTbl, voteOptionTbl.ID.EQ(userVotesTbl.OptionID)),
	).GROUP_BY(
		voteOptionTbl.VoteID,
	)

	var results []struct {
		VoteID            int64  `alias:"vote_options.vote_id"`
		ParticipantsCount int64  `alias:"participants_count"`
		TotalVotes        *int64 `alias:"total_votes"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return err
	}

	// 对每个投票执行更新
	for _, r := range results {
		totalVotes := int64(0)
		if r.TotalVotes != nil {
			totalVotes = *r.TotalVotes
		}

		updateStmt := voteTbl.UPDATE().SET(
			voteTbl.ParticipantsCount.SET(pg.Int64(r.ParticipantsCount)),
			voteTbl.TotalVotesCount.SET(pg.Int64(totalVotes)),
		).WHERE(
			voteTbl.ID.EQ(pg.Int64(r.VoteID)),
		)
		_, err := updateStmt.ExecContext(ctx, db)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
RecalculateVoteOptionStats 重新计算指定投票的所有选项的票数统计.
*/
func RecalculateVoteOptionStats(
	ctx context.Context,
	db qrm.DB,
	voteID int64,
) error {
	voteOptionTbl := table.VoteOptions
	userVotesTbl := table.UserVotes

	// 获取该投票的所有选项
	optionsStmt := pg.SELECT(voteOptionTbl.ID).
		FROM(voteOptionTbl).
		WHERE(voteOptionTbl.VoteID.EQ(pg.Int64(voteID)))

	var options []struct {
		ID int64 `alias:"vote_options.id"`
	}
	err := optionsStmt.QueryContext(ctx, db, &options)
	if err != nil {
		return err
	}

	// 统计每个选项的总票数
	stmt := pg.SELECT(
		userVotesTbl.OptionID,
		pg.SUM(pg.CAST(userVotesTbl.VoteCount).AS("BIGINT")).AS("total_votes"),
	).FROM(
		userVotesTbl,
	).WHERE(
		userVotesTbl.OptionID.IN(
			pg.SELECT(voteOptionTbl.ID).
				FROM(voteOptionTbl).
				WHERE(voteOptionTbl.VoteID.EQ(pg.Int64(voteID))),
		),
	).GROUP_BY(
		userVotesTbl.OptionID,
	)

	var results []struct {
		OptionID   int64  `alias:"user_votes.option_id"`
		TotalVotes *int64 `alias:"total_votes"`
	}
	err = stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return err
	}

	// 构建选项ID到票数的映射
	voteCountMap := make(map[int64]int64)
	for _, r := range results {
		voteCount := int64(0)
		if r.TotalVotes != nil {
			voteCount = *r.TotalVotes
		}
		voteCountMap[r.OptionID] = voteCount
	}

	// 对每个选项执行更新
	for _, opt := range options {
		voteCount := voteCountMap[opt.ID] // 如果没有投票记录，默认为 0

		updateStmt := voteOptionTbl.UPDATE().SET(
			voteOptionTbl.VoteCount.SET(pg.Int64(voteCount)),
		).WHERE(
			voteOptionTbl.ID.EQ(pg.Int64(opt.ID)),
		)
		_, err := updateStmt.ExecContext(ctx, db)
		if err != nil {
			return err
		}
	}

	return nil
}
