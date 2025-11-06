package question_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
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
	questionIDs []int64,
) (map[int64]bool, error) {
	if len(questionIDs) == 0 {
		return make(map[int64]bool), nil
	}

	questionTbl := table.Questions
	submissionTbl := table.QuestionSubmissions

	// 构建 ID 列表
	idList := make([]pg.Expression, 0, len(questionIDs))
	for _, id := range questionIDs {
		idList = append(idList, pg.Int64(id))
	}

	stmt := pg.SELECT(
		questionTbl.ID,
	).FROM(
		submissionTbl.INNER_JOIN(questionTbl, questionTbl.ID.EQ(submissionTbl.QuestionID)),
	).WHERE(
		submissionTbl.UserID.EQ(pg.Int64(userID)).
			AND(questionTbl.ID.IN(idList...)).
			AND(submissionTbl.IsPractice.EQ(pg.Bool(false))),
	).GROUP_BY(questionTbl.ID)

	var results []struct {
		QuestionID int64 `alias:"questions.id"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, err
	}

	// 构建结果 map
	solvedMap := make(map[int64]bool)

	// 初始化所有题目为未解答
	for _, id := range questionIDs {
		solvedMap[id] = false
	}

	// 标记已解答的题目
	for _, result := range results {
		solvedMap[result.QuestionID] = true
	}

	return solvedMap, nil
}
