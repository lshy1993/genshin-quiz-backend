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
