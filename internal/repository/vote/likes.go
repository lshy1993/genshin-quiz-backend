package vote_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

// GetVoteLikeStatus 获取用户对指定投票的点赞状态.
func GetVoteLikeStatus(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	voteID int64,
) (*int16, error) {
	likesTbl := table.VoteLikes

	stmt := pg.SELECT(
		likesTbl.Value,
	).FROM(
		likesTbl,
	).WHERE(
		likesTbl.UserID.EQ(pg.Int64(userID)).
			AND(likesTbl.VoteID.EQ(pg.Int64(voteID))),
	)

	var results []struct {
		Value int16 `alias:"vote_likes.value"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, err
	}

	// 如果没有找到记录，返回 nil 表示未点赞
	if len(results) == 0 {
		defaultValue := int16(0)
		return &defaultValue, nil
	}

	// 返回点赞状态
	return &results[0].Value, nil
}

// GetMultipleVotesLikeStatus 批量获取用户对多个投票的点赞状态.
func GetMultipleVotesLikeStatus(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	voteUUIDs []uuid.UUID,
) (map[uuid.UUID]*int16, error) {
	if len(voteUUIDs) == 0 {
		return make(map[uuid.UUID]*int16), nil
	}

	voteTbl := table.Votes
	likesTbl := table.VoteLikes

	// 构建 UUID 列表
	uuidList := make([]pg.Expression, 0, len(voteUUIDs))
	for _, uuid := range voteUUIDs {
		uuidList = append(uuidList, pg.UUID(uuid))
	}

	stmt := pg.SELECT(
		voteTbl.VoteUUID,
		likesTbl.Value,
	).FROM(
		likesTbl.INNER_JOIN(voteTbl, voteTbl.ID.EQ(likesTbl.VoteID)),
	).WHERE(
		likesTbl.UserID.EQ(pg.Int64(userID)).
			AND(voteTbl.VoteUUID.IN(uuidList...)),
	)

	var results []struct {
		VoteUUID uuid.UUID `alias:"votes.vote_uuid"`
		Value    int16     `alias:"vote_likes.value"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, err
	}

	// 构建结果 map
	likeStatusMap := make(map[uuid.UUID]*int16)

	// 初始化所有投票为未点赞
	for _, uuid := range voteUUIDs {
		likeStatusMap[uuid] = nil
	}

	// 设置已点赞的投票状态
	for _, result := range results {
		value := result.Value
		likeStatusMap[result.VoteUUID] = &value
	}

	return likeStatusMap, nil
}

// GetVoteLikesCount 获取投票的总点赞数（实时计算）。
func GetVoteLikesCount(
	ctx context.Context,
	db qrm.DB,
	voteID int64,
) (int64, error) {
	likesTbl := table.VoteLikes

	// 统计该投票的点赞数（value = 1）
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(likesTbl).
		WHERE(
			likesTbl.VoteID.EQ(pg.Int64(voteID)).
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

// GetMultipleVotesLikesCount 批量获取多个投票的点赞数。
func GetMultipleVotesLikesCount(
	ctx context.Context,
	db qrm.DB,
	voteIDs []int64,
) (map[int64]int64, error) {
	if len(voteIDs) == 0 {
		return make(map[int64]int64), nil
	}

	likesTbl := table.VoteLikes

	// 构建投票ID列表
	idList := make([]pg.Expression, 0, len(voteIDs))
	for _, id := range voteIDs {
		idList = append(idList, pg.Int64(id))
	}

	// 统计每个投票的点赞数
	stmt := pg.SELECT(
		likesTbl.VoteID,
		pg.COUNT(pg.STAR).AS("count"),
	).FROM(likesTbl).
		WHERE(
			likesTbl.VoteID.IN(idList...).
				AND(likesTbl.Value.EQ(pg.Int16(1))),
		).
		GROUP_BY(likesTbl.VoteID)

	var results []struct {
		VoteID int64 `alias:"vote_id"`
		Count  int64 `alias:"count"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, err
	}

	// 构建结果map
	likesCountMap := make(map[int64]int64)

	// 初始化所有投票为0
	for _, id := range voteIDs {
		likesCountMap[id] = 0
	}

	// 设置有点赞的投票
	for _, result := range results {
		likesCountMap[result.VoteID] = result.Count
	}

	return likesCountMap, nil
}

// ============================================
// 投票评论功能预留
// ============================================
// 以下函数为投票评论功能预留，表结构已在迁移文件中创建（vote_comments）
// 实现时需要添加：
// - GetVoteComments: 获取投票的评论列表
// - GetVoteCommentCount: 获取投票的评论总数
// - InsertVoteComment: 添加投票评论
// - UpdateVoteComment: 编辑投票评论
// - DeleteVoteComment: 删除投票评论
// - GetMultipleVotesCommentsCount: 批量获取多个投票的评论数（避免N+1查询）

// UpdateVoteLikeCount 更新投票的总点赞数。
func UpdateVoteLikeCount(
	ctx context.Context,
	db qrm.DB,
	voteID int64,
) error {
	voteTbl := table.Votes
	likesTbl := table.VoteLikes

	// 统计该投票的点赞数（value = 1）
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(likesTbl).
		WHERE(
			likesTbl.VoteID.EQ(pg.Int64(voteID)).
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
	updateStmt := voteTbl.UPDATE().
		SET(voteTbl.LikesCount.SET(pg.Int32(int32(countResult.Count)))).
		WHERE(voteTbl.ID.EQ(pg.Int64(voteID)))

	_, err = updateStmt.ExecContext(ctx, db)
	if err != nil {
		return err
	}

	return nil
}

// UpsertVoteLike 插入或更新用户对投票的点赞状态.
func UpsertVoteLike(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	voteUUID uuid.UUID,
	value int16, // 1=点赞, -1=点踩, 0=取消
) error {
	voteTbl := table.Votes
	likesTbl := table.VoteLikes

	// 首先获取 vote_id
	voteIDStmt := pg.SELECT(voteTbl.ID).
		FROM(voteTbl).
		WHERE(voteTbl.VoteUUID.EQ(pg.UUID(voteUUID)))

	var voteIDResult struct {
		ID int64 `alias:"votes.id"`
	}
	err := voteIDStmt.QueryContext(ctx, db, &voteIDResult)
	if err != nil {
		return err
	}

	voteID := voteIDResult.ID

	if value == 0 {
		// 删除点赞记录
		deleteStmt := likesTbl.DELETE().WHERE(
			likesTbl.VoteID.EQ(pg.Int64(voteID)).
				AND(likesTbl.UserID.EQ(pg.Int64(userID))),
		)
		_, err = deleteStmt.ExecContext(ctx, db)
		return err
	}

	// 插入或更新点赞记录
	now := pg.NOW()
	upsertStmt := likesTbl.INSERT(
		likesTbl.VoteID,
		likesTbl.UserID,
		likesTbl.Value,
		likesTbl.CreatedAt,
		likesTbl.UpdatedAt,
	).VALUES(
		voteID,
		userID,
		value,
		now,
		now,
	).ON_CONFLICT(likesTbl.VoteID, likesTbl.UserID).DO_UPDATE(
		pg.SET(
			likesTbl.Value.SET(pg.Int16(value)),
			likesTbl.UpdatedAt.SET(now),
		),
	)

	_, err = upsertStmt.ExecContext(ctx, db)
	return err
}
