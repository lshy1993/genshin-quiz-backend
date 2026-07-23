package question_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

/*
UpdateQuestionsLikesCount 根据 question_likes 表统计每个问题的点赞数，并更新 questions.likes 字段.
只统计 value=1 的点赞数.
*/
func RecalculateAllQuestionLikeStats(
	ctx context.Context,
	db qrm.DB,
) error {
	questionTbl := table.Questions
	likesTbl := table.QuestionLikes

	// 统计每个问题的点赞数
	stmt := pg.SELECT(
		likesTbl.QuestionID,
		pg.COUNT(likesTbl.Value).AS("like_count"),
	).FROM(
		likesTbl,
	).WHERE(
		likesTbl.Value.EQ(pg.Int16(1)),
	).GROUP_BY(
		likesTbl.QuestionID,
	)

	var results []struct {
		QuestionID int64 `alias:"question_likes.question_id"`
		LikeCount  int64 `alias:"like_count"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return err
	}

	// 构建 questionID -> likeCount 映射
	likeCountMap := make(map[int64]int64, len(results))
	for _, r := range results {
		likeCountMap[r.QuestionID] = r.LikeCount
	}

	// 对每个问题执行更新
	for qid, likeCount := range likeCountMap {
		updateStmt := questionTbl.UPDATE().SET(
			questionTbl.Likes.SET(pg.Int64(likeCount)),
		).WHERE(
			questionTbl.ID.EQ(pg.Int64(qid)),
		)
		_, err := updateStmt.ExecContext(ctx, db)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
RecalculateAllQuestionSubmissionStats 根据 question_submissions 表统计每个问题的提交数据，
并更新 questions 表的 submit_count 和 correct_count 字段.
只统计非练习模式 (is_practice=false) 的提交.
*/
func RecalculateAllQuestionSubmissionStats(
	ctx context.Context,
	db qrm.DB,
) error {
	questionTbl := table.Questions
	submissionTbl := table.QuestionSubmissions

	// 统计每个问题的总提交数和正确提交数
	stmt := pg.SELECT(
		submissionTbl.QuestionID,
		pg.COUNT(pg.STAR).AS("total_submissions"),
		pg.SUM(pg.CAST(submissionTbl.IsCorrect).AS("INTEGER")).AS("correct_submissions"),
	).FROM(
		submissionTbl,
	).WHERE(
		submissionTbl.IsPractice.EQ(pg.Bool(false)), // 只统计非练习模式
	).GROUP_BY(
		submissionTbl.QuestionID,
	)

	var results []struct {
		QuestionID         int64  `alias:"question_submissions.question_id"`
		TotalSubmissions   int64  `alias:"total_submissions"`
		CorrectSubmissions *int64 `alias:"correct_submissions"` // 使用指针处理可能的 NULL
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return err
	}

	// 对每个问题执行更新
	for _, r := range results {
		correctCount := int64(0)
		if r.CorrectSubmissions != nil {
			correctCount = *r.CorrectSubmissions
		}

		updateStmt := questionTbl.UPDATE().SET(
			questionTbl.SubmitCount.SET(pg.Int64(r.TotalSubmissions)),
			questionTbl.CorrectCount.SET(pg.Int64(correctCount)),
		).WHERE(
			questionTbl.ID.EQ(pg.Int64(r.QuestionID)),
		)
		_, err := updateStmt.ExecContext(ctx, db)
		if err != nil {
			return err
		}
	}

	return nil
}
