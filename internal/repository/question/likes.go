package question_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

// GetQuestionLikeStatus 获取用户对指定问题的点赞状态.
func GetQuestionLikeStatus(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	questionID int64,
) (*int16, error) {
	likesTbl := table.QuestionLikes

	stmt := pg.SELECT(
		likesTbl.Value,
	).FROM(
		likesTbl,
	).WHERE(
		likesTbl.UserID.EQ(pg.Int64(userID)).
			AND(likesTbl.QuestionID.EQ(pg.Int64(questionID))),
	)

	var results []struct {
		Value int16 `alias:"question_likes.value"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, err
	}

	// 如果没有找到记录，返回默认值 0 表示未点赞
	if len(results) == 0 {
		defaultValue := int16(0)
		return &defaultValue, nil
	}

	// 返回点赞状态
	return &results[0].Value, nil
}

// GetMultipleQuestionsLikeStatus 批量获取用户对多个问题的点赞状态.
func GetMultipleQuestionsLikeStatus(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	questionUUIDs []uuid.UUID,
) (map[uuid.UUID]*int16, error) {
	if len(questionUUIDs) == 0 {
		return make(map[uuid.UUID]*int16), nil
	}

	questionTbl := table.Questions
	likesTbl := table.QuestionLikes

	// 构建 UUID 列表
	uuidList := make([]pg.Expression, 0, len(questionUUIDs))
	for _, uuid := range questionUUIDs {
		uuidList = append(uuidList, pg.UUID(uuid))
	}

	stmt := pg.SELECT(
		questionTbl.QuestionUUID,
		likesTbl.Value,
	).FROM(
		likesTbl.INNER_JOIN(questionTbl, questionTbl.ID.EQ(likesTbl.QuestionID)),
	).WHERE(
		likesTbl.UserID.EQ(pg.Int64(userID)).
			AND(questionTbl.QuestionUUID.IN(uuidList...)),
	)

	var results []struct {
		QuestionUUID uuid.UUID `alias:"questions.question_uuid"`
		Value        int16     `alias:"question_likes.value"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, err
	}

	// 构建结果 map
	likeStatusMap := make(map[uuid.UUID]*int16)

	// 初始化所有题目为未点赞
	for _, uuid := range questionUUIDs {
		likeStatusMap[uuid] = nil
	}

	// 设置已点赞的题目状态
	for _, result := range results {
		value := result.Value
		likeStatusMap[result.QuestionUUID] = &value
	}

	return likeStatusMap, nil
}

// UpsertQuestionLike 插入或更新用户对问题的点赞状态.
func UpsertQuestionLike(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	questionUUID uuid.UUID,
	value int16, // 1=点赞, -1=点踩, 0=取消
) error {
	questionTbl := table.Questions
	likesTbl := table.QuestionLikes

	// 首先获取 question_id
	questionIDStmt := pg.SELECT(questionTbl.ID).
		FROM(questionTbl).
		WHERE(questionTbl.QuestionUUID.EQ(pg.UUID(questionUUID)))

	var questionIDResult struct {
		ID int64 `alias:"questions.id"`
	}
	err := questionIDStmt.QueryContext(ctx, db, &questionIDResult)
	if err != nil {
		return err
	}

	questionID := questionIDResult.ID

	if value == 0 {
		// 删除点赞记录
		deleteStmt := likesTbl.DELETE().WHERE(
			likesTbl.QuestionID.EQ(pg.Int64(questionID)).
				AND(likesTbl.UserID.EQ(pg.Int64(userID))),
		)
		_, err = deleteStmt.ExecContext(ctx, db)
		return err
	}

	// 插入或更新点赞记录
	now := pg.NOW()
	upsertStmt := likesTbl.INSERT(
		likesTbl.QuestionID,
		likesTbl.UserID,
		likesTbl.Value,
		likesTbl.CreatedAt,
		likesTbl.UpdatedAt,
	).VALUES(
		questionID,
		userID,
		value,
		now,
		now,
	).ON_CONFLICT(likesTbl.QuestionID, likesTbl.UserID).DO_UPDATE(
		pg.SET(
			likesTbl.Value.SET(pg.Int16(value)),
			likesTbl.UpdatedAt.SET(now),
		),
	)

	_, err = upsertStmt.ExecContext(ctx, db)
	return err
}

// GetQuestionLikesCount 获取问题的总点赞数（实时计算）。
func GetQuestionLikesCount(
	ctx context.Context,
	db qrm.DB,
	questionID int64,
) (int64, error) {
	likesTbl := table.QuestionLikes

	// 统计该问题的点赞数（value = 1）
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(likesTbl).
		WHERE(
			likesTbl.QuestionID.EQ(pg.Int64(questionID)).
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

// GetMultipleQuestionsLikesCount 批量获取多个问题的点赞数。
func GetMultipleQuestionsLikesCount(
	ctx context.Context,
	db qrm.DB,
	questionIDs []int64,
) (map[int64]int64, error) {
	if len(questionIDs) == 0 {
		return make(map[int64]int64), nil
	}

	likesTbl := table.QuestionLikes

	// 构建问题ID列表
	idList := make([]pg.Expression, 0, len(questionIDs))
	for _, id := range questionIDs {
		idList = append(idList, pg.Int64(id))
	}

	// 统计每个问题的点赞数
	stmt := pg.SELECT(
		likesTbl.QuestionID,
		pg.COUNT(pg.STAR).AS("count"),
	).FROM(likesTbl).
		WHERE(
			likesTbl.QuestionID.IN(idList...).
				AND(likesTbl.Value.EQ(pg.Int16(1))),
		).
		GROUP_BY(likesTbl.QuestionID)

	var results []struct {
		QuestionID int64 `alias:"question_id"`
		Count      int64 `alias:"count"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, err
	}

	// 构建结果map
	likesCountMap := make(map[int64]int64)

	// 初始化所有问题为0
	for _, id := range questionIDs {
		likesCountMap[id] = 0
	}

	// 设置有点赞的问题
	for _, result := range results {
		likesCountMap[result.QuestionID] = result.Count
	}

	return likesCountMap, nil
}
