package user_repo

import (
	"context"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

func UpdateUserSubmissionStats(
	ctx context.Context,
	db qrm.DB,
	userID int64,
	isCorrect bool,
) error {
	tbl := table.Users

	// 构建更新语句
	if isCorrect {
		// 如果答案正确，同时更新总提交数和正确答案数
		updateStmt := tbl.UPDATE().SET(
			tbl.TotalSubmissions.SET(tbl.TotalSubmissions.ADD(pg.Int(1))),
			tbl.CorrectSubmissions.SET(tbl.CorrectSubmissions.ADD(pg.Int(1))),
		).WHERE(tbl.ID.EQ(pg.Int64(userID)))

		_, err := updateStmt.ExecContext(ctx, db)
		return err
	} else {
		// 如果答案错误，只更新总提交数
		updateStmt := tbl.UPDATE().SET(
			tbl.TotalSubmissions.SET(tbl.TotalSubmissions.ADD(pg.Int(1))),
		).WHERE(tbl.ID.EQ(pg.Int64(userID)))

		_, err := updateStmt.ExecContext(ctx, db)
		return err
	}
}

func RecalculateUserStats(
	ctx context.Context,
	db qrm.DB,
	userID int64,
) error {
	userTbl := table.Users
	submissionTbl := table.QuestionSubmissions
	questionTbl := table.Questions

	// 计算用户的提交统计
	submissionStats := pg.SELECT(
		pg.COUNT(pg.STAR).AS("total_submissions"),
		pg.SUM(
			pg.CAST(submissionTbl.IsCorrect).AS("INTEGER"),
		).AS("correct_submissions"),
	).FROM(submissionTbl).
		WHERE(
			submissionTbl.UserID.EQ(pg.Int64(userID)).
				AND(submissionTbl.IsPractice.EQ(pg.Bool(false))),
		)

	// 计算用户创建的题目数
	questionStats := pg.SELECT(
		pg.COUNT(pg.STAR).AS("questions_created"),
	).FROM(questionTbl).
		WHERE(questionTbl.CreatedBy.EQ(pg.Int64(userID)))

	// 执行统计查询
	var submissionResult struct {
		TotalSubmissions   int64 `alias:"total_submissions"`
		CorrectSubmissions int64 `alias:"correct_submissions"`
	}
	err := submissionStats.QueryContext(ctx, db, &submissionResult)
	if err != nil {
		return err
	}

	var questionResult struct {
		QuestionsCreated int64 `alias:"questions_created"`
	}
	err = questionStats.QueryContext(ctx, db, &questionResult)
	if err != nil {
		return err
	}

	// 更新用户统计信息
	updateStmt := userTbl.UPDATE().SET(
		userTbl.TotalSubmissions.SET(pg.Int64(submissionResult.TotalSubmissions)),
		userTbl.CorrectSubmissions.SET(pg.Int64(submissionResult.CorrectSubmissions)),
		userTbl.QuestionsCreated.SET(pg.Int64(questionResult.QuestionsCreated)),
	).WHERE(userTbl.ID.EQ(pg.Int64(userID)))

	_, err = updateStmt.ExecContext(ctx, db)
	return err
}

// RecalculateAllUserStats 重新计算所有用户的统计信息.
func RecalculateAllUserStats(
	ctx context.Context,
	db qrm.DB,
) error {
	userTbl := table.Users

	// 获取所有用户ID
	stmt := pg.SELECT(userTbl.ID).FROM(userTbl)

	var userIDs []struct {
		ID int64 `alias:"users.id"`
	}
	err := stmt.QueryContext(ctx, db, &userIDs)
	if err != nil {
		return err
	}

	// 逐个重新计算用户统计
	for _, user := range userIDs {
		err = RecalculateUserStats(ctx, db, user.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
