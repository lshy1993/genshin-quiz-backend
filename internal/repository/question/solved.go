package question_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

// CheckQuestionSolved 检查用户是否已正确解决指定题目.
func CheckQuestionSolved(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	questionID int64,
) (bool, error) {
	submissionTbl := table.QuestionSubmissions

	stmt := pg.SELECT(
		pg.COUNT(pg.STAR).AS("count"),
	).FROM(
		submissionTbl,
	).WHERE(
		submissionTbl.UserID.EQ(pg.Int64(userID)).
			AND(submissionTbl.QuestionID.EQ(pg.Int64(questionID))).
			AND(submissionTbl.IsCorrect.IS_TRUE()).
			AND(submissionTbl.IsPractice.IS_FALSE()),
	)

	var result struct {
		Count int64 `alias:"count"`
	}
	err := stmt.QueryContext(ctx, db, &result)
	if err != nil {
		return false, err
	}

	return result.Count > 0, nil
}

// CheckMultipleQuestionsSolved 批量检查用户是否已解答多个题目.
func CheckMultipleQuestionsSolved(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	questionUUIDs []uuid.UUID,
) (map[uuid.UUID]bool, error) {
	if len(questionUUIDs) == 0 {
		return make(map[uuid.UUID]bool), nil
	}

	questionTbl := table.Questions
	submissionTbl := table.QuestionSubmissions

	// 构建 UUID 列表
	uuidList := make([]pg.Expression, 0, len(questionUUIDs))
	for _, uuid := range questionUUIDs {
		uuidList = append(uuidList, pg.UUID(uuid))
	}

	stmt := pg.SELECT(
		questionTbl.QuestionUUID,
	).FROM(
		submissionTbl.INNER_JOIN(questionTbl, questionTbl.ID.EQ(submissionTbl.QuestionID)),
	).WHERE(
		submissionTbl.UserID.EQ(pg.Int64(userID)).
			AND(questionTbl.QuestionUUID.IN(uuidList...)).
			AND(submissionTbl.IsPractice.EQ(pg.Bool(false))),
	).GROUP_BY(questionTbl.QuestionUUID)

	var results []struct {
		QuestionUUID uuid.UUID `alias:"questions.question_uuid"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, err
	}

	// 构建结果 map
	solvedMap := make(map[uuid.UUID]bool)

	// 初始化所有题目为未解答
	for _, uuid := range questionUUIDs {
		solvedMap[uuid] = false
	}

	// 标记已解答的题目
	for _, result := range results {
		solvedMap[result.QuestionUUID] = true
	}

	return solvedMap, nil
}
