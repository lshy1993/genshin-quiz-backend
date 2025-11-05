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
