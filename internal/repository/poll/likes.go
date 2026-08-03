package poll_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/table"

	"github.com/go-errors/errors"
	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

// GetPollLikeStatus 获取用户对指定投票的点赞状态.
func GetPollLikeStatus(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	pollID int64,
) (*int16, error) {
	likesTbl := table.PollLikes

	stmt := pg.SELECT(
		likesTbl.Value,
	).FROM(
		likesTbl,
	).WHERE(
		likesTbl.UserID.EQ(pg.Int64(userID)).
			AND(likesTbl.PollID.EQ(pg.Int64(pollID))),
	)

	var results []struct {
		Value int16 `alias:"poll_likes.value"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get poll like status failed", 0)
	}

	// 如果没有找到记录，返回 nil 表示未点赞
	if len(results) == 0 {
		defaultValue := int16(0)
		return &defaultValue, nil
	}

	// 返回点赞状态
	return &results[0].Value, nil
}

// GetPollsLikeStatusByUser 批量获取用户对多个投票的点赞状态.
func GetPollsLikeStatusByUser(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	pollIDs []int64,
) (map[int64]int16, error) {
	if len(pollIDs) == 0 {
		return make(map[int64]int16), nil
	}

	pollTbl := table.Polls
	likesTbl := table.PollLikes

	// 构建 UUID 列表
	idList := make([]pg.Expression, 0, len(pollIDs))
	for _, id := range pollIDs {
		idList = append(idList, pg.Int64(id))
	}

	stmt := pg.SELECT(
		pollTbl.ID,
		likesTbl.Value,
	).FROM(
		likesTbl.INNER_JOIN(pollTbl, pollTbl.ID.EQ(likesTbl.PollID)),
	).WHERE(
		likesTbl.UserID.EQ(pg.Int64(userID)).
			AND(pollTbl.ID.IN(idList...)),
	)

	var results []struct {
		PollID int64 `alias:"polls.id"`
		Value  int16 `alias:"poll_likes.value"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get poll like status by user failed", 0)
	}

	// 构建结果 map
	likeStatusMap := make(map[int64]int16)

	// 设置已点赞的投票状态
	for _, result := range results {
		value := result.Value
		likeStatusMap[result.PollID] = value
	}

	return likeStatusMap, nil
}

// GetPollLikesCount 获取投票的总点赞数（实时计算）。
func GetPollLikesCount(
	ctx context.Context,
	db qrm.DB,
	pollID int64,
) (int64, error) {
	likesTbl := table.PollLikes

	// 统计该投票的点赞数（value = 1）
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(likesTbl).
		WHERE(
			likesTbl.PollID.EQ(pg.Int64(pollID)).
				AND(likesTbl.Value.EQ(pg.Int16(1))),
		)

	var countResult struct {
		Count int64 `alias:"count"`
	}
	err := countStmt.QueryContext(ctx, db, &countResult)
	if err != nil {
		return 0, err
	}

	return countResult.Count, nil
}

// GetMultiplePollsLikesCount 批量获取多个投票的点赞数。
func GetMultiplePollsLikesCount(
	ctx context.Context,
	db qrm.DB,
	pollIDs []int64,
) (map[int64]int64, error) {
	if len(pollIDs) == 0 {
		return make(map[int64]int64), nil
	}

	likesTbl := table.PollLikes

	// 构建投票ID列表
	idList := make([]pg.Expression, 0, len(pollIDs))
	for _, id := range pollIDs {
		idList = append(idList, pg.Int64(id))
	}

	// 统计每个投票的点赞数
	stmt := pg.SELECT(
		likesTbl.PollID,
		pg.COUNT(pg.STAR).AS("count"),
	).FROM(likesTbl).
		WHERE(
			likesTbl.PollID.IN(idList...).
				AND(likesTbl.Value.EQ(pg.Int16(1))),
		).
		GROUP_BY(likesTbl.PollID)

	var results []struct {
		PollID int64 `alias:"poll_likes.poll_id"`
		Count  int64 `alias:"count"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get multi poll likes count failed", 0)
	}

	// 构建结果map
	likesCountMap := make(map[int64]int64, len(pollIDs))
	// 设置有点赞的投票
	for _, result := range results {
		likesCountMap[result.PollID] = result.Count
	}

	return likesCountMap, nil
}

// ============================================
// 投票评论功能预留
// ============================================
// 以下函数为投票评论功能预留，表结构已在迁移文件中创建（poll_comments）
// 实现时需要添加：
// - GetPollComments: 获取投票的评论列表
// - GetPollCommentCount: 获取投票的评论总数
// - InsertPollComment: 添加投票评论
// - UpdatePollComment: 编辑投票评论
// - DeletePollComment: 删除投票评论
// - GetMultiplePollsCommentsCount: 批量获取多个投票的评论数（避免N+1查询）

// UpdatePollLikeCount 更新投票的总点赞数。
func UpdatePollLikeCount(
	ctx context.Context,
	db qrm.DB,
	pollID int64,
) error {
	pollTbl := table.Polls
	likesTbl := table.PollLikes

	// 统计该投票的点赞数（value = 1）
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(likesTbl).
		WHERE(
			likesTbl.PollID.EQ(pg.Int64(pollID)).
				AND(likesTbl.Value.EQ(pg.Int16(1))),
		)

	var countResult struct {
		Count int64 `alias:"count"`
	}
	err := countStmt.QueryContext(ctx, db, &countResult)
	if err != nil {
		return err
	}

	// 更新投票表中的 likes_count
	updateStmt := pollTbl.UPDATE().
		SET(pollTbl.LikesCount.SET(pg.Int32(int32(countResult.Count)))).
		WHERE(pollTbl.ID.EQ(pg.Int64(pollID)))

	_, err = updateStmt.ExecContext(ctx, db)
	if err != nil {
		return err
	}

	return nil
}

// UpsertPollLike 插入或更新用户对投票的点赞状态.
func UpsertPollLike(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	pollUUID uuid.UUID,
	value int16, // 1=点赞, -1=点踩, 0=取消
) error {
	pollTbl := table.Polls
	likesTbl := table.PollLikes

	// 首先获取 poll_id
	pollIDStmt := pg.SELECT(pollTbl.ID).
		FROM(pollTbl).
		WHERE(pollTbl.PollUUID.EQ(pg.UUID(pollUUID)))

	var pollIDResult struct {
		ID int64 `alias:"polls.id"`
	}
	err := pollIDStmt.QueryContext(ctx, db, &pollIDResult)
	if err != nil {
		return err
	}

	pollID := pollIDResult.ID

	if value == 0 {
		// 删除点赞记录
		deleteStmt := likesTbl.DELETE().WHERE(
			likesTbl.PollID.EQ(pg.Int64(pollID)).
				AND(likesTbl.UserID.EQ(pg.Int64(userID))),
		)
		_, err = deleteStmt.ExecContext(ctx, db)
		return err
	}

	// 插入或更新点赞记录
	now := pg.NOW()
	upsertStmt := likesTbl.INSERT(
		likesTbl.PollID,
		likesTbl.UserID,
		likesTbl.Value,
		likesTbl.CreatedAt,
		likesTbl.UpdatedAt,
	).VALUES(
		pollID,
		userID,
		value,
		now,
		now,
	).ON_CONFLICT(likesTbl.PollID, likesTbl.UserID).DO_UPDATE(
		pg.SET(
			likesTbl.Value.SET(pg.Int16(value)),
			likesTbl.UpdatedAt.SET(now),
		),
	)

	_, err = upsertStmt.ExecContext(ctx, db)
	return err
}
