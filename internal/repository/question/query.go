package question_repo

import (
	"context"
	"strings"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"
	"genshin-quiz/generated/oapi"
	dao "genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"
	"genshin-quiz/internal/webserver/middleware"

	"genshin-quiz/internal/common"
	"genshin-quiz/internal/util"

	"github.com/go-errors/errors"
	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
)

func GetQuestions(
	ctx context.Context,
	db qrm.DB,
	params dao.QuestionListParams,
) (*dao.QuestionListResult, error) {
	tbl := table.Questions
	userTbl := table.Users

	offset := (params.Page - 1) * params.NumPerPage
	if offset < 0 {
		offset = 0
	}

	stmt := pg.SELECT(
		tbl.AllColumns,
		userTbl.AllColumns,
	).FROM(
		tbl.
			LEFT_JOIN(userTbl, tbl.CreatedBy.EQ(userTbl.ID)),
	).
		WHERE(
			buildQuestionCondition(params),
		).
		ORDER_BY(buildQuestionOrder(params)).
		LIMIT(int64(params.NumPerPage)).
		OFFSET(int64(offset))

	// 先获取总数
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(tbl).
		WHERE(buildQuestionCondition(params))
	var countResult struct {
		Count int64 `alias:"count"`
	}
	err := countStmt.QueryContext(ctx, db, &countResult)
	if err != nil {
		return nil, err
	}

	var questions []dao.SimpleQuestion
	err = stmt.QueryContext(ctx, db, &questions)
	if err != nil {
		return nil, err
	}

	return &dao.QuestionListResult{
		Questions: questions,
		Total:     int(countResult.Count),
	}, nil
}

func buildQuestionCondition(params dao.QuestionListParams) pg.BoolExpression {
	tbl := table.Questions
	transTbl := table.QuestionTranslations

	condition := pg.Bool(true)

	if params.IsPublic != nil {
		if *params.IsPublic {
			condition = condition.AND(tbl.Public.IS_TRUE())
		} else {
			condition = condition.AND(tbl.Public.IS_FALSE())
		}
	}

	if params.IsPublished != nil {
		if *params.IsPublished {
			condition = condition.AND(tbl.IsPublished.IS_TRUE())
		} else {
			condition = condition.AND(tbl.IsPublished.IS_FALSE())
		}
	}

	// 添加创建者过滤
	if params.Author != nil {
		userID := *params.Author
		condition = condition.AND(tbl.CreatedBy.EQ(pg.Int64(userID)))
	}

	// 添加分类过滤
	if params.Category != nil {
		cat := string(*params.Category)
		condition = condition.AND(tbl.Category.EQ(pg.NewEnumValue(cat)))
	}

	// 添加难度过滤
	if params.Difficulty != nil {
		diffExp := []pg.Expression{}
		for _, diff := range *params.Difficulty {
			diffStr := string(diff)
			diffExp = append(diffExp, pg.NewEnumValue(diffStr))
		}
		condition = condition.AND(tbl.Difficulty.IN(diffExp...))
	}

	// 添加关键字搜索（在翻译表的question_text中搜索）
	if params.Query != nil && *params.Query != "" {
		searchTerm := "%" + strings.ToLower(*params.Query) + "%"

		condition = condition.AND(
			pg.EXISTS(
				pg.SELECT(pg.Int(1)).
					FROM(transTbl).
					WHERE(
						transTbl.QuestionID.EQ(tbl.ID).
							AND(
								pg.LOWER(transTbl.QuestionText).LIKE(pg.String(searchTerm)),
							),
					),
			),
		)
	}

	return condition
}

func buildQuestionOrder(params dao.QuestionListParams) pg.OrderByClause {
	tbl := table.Questions
	var orderExpr pg.Expression

	switch params.SortBy {
	case "PublishDate": // 上线时间
		orderExpr = tbl.PublishedAt
	case "Difficulty":
		// 难度排序：easy < medium < hard
		orderExpr = pg.CASE().
			WHEN(tbl.Difficulty.EQ(pg.String("easy"))).THEN(pg.Int(1)).
			WHEN(tbl.Difficulty.EQ(pg.String("medium"))).THEN(pg.Int(2)).
			WHEN(tbl.Difficulty.EQ(pg.String("hard"))).THEN(pg.Int(3)).
			ELSE(pg.Int(0))
	case "Likes": // 点赞数
		orderExpr = tbl.Likes
	case "Submissions": // 参与人数
		orderExpr = tbl.SubmitCount
	case "CorrectRate":
		orderExpr = tbl.CorrectCount
	default:
		orderExpr = tbl.PublishedAt
	}

	if params.SortDesc {
		return orderExpr.DESC()
	}
	return orderExpr.ASC()
}

func GetQuestionByUUID(
	ctx context.Context,
	db qrm.DB,
	uuid uuid.UUID,
) (*dao.SimpleQuestion, error) {
	tbl := table.Questions
	userTbl := table.Users

	stmt := pg.SELECT(
		tbl.AllColumns,
		userTbl.UserUUID,
	).FROM(
		tbl.LEFT_JOIN(userTbl, userTbl.ID.EQ(tbl.CreatedBy)),
	).WHERE(
		tbl.QuestionUUID.EQ(pg.UUID(uuid)),
	)

	var result []dao.SimpleQuestion
	err := stmt.QueryContext(ctx, db, &result)
	if err != nil {
		return nil, errors.WrapPrefix(err, "query question by uuid failed", 0)
	}

	if len(result) == 0 {
		return nil, common.ErrQuestionNotFound
	}

	return &result[0], nil
}

// getQuestionTranslationFallback 获取题目的任意语言翻译（用于 fallback）.
// func getQuestionTranslationFallback(
// 	ctx context.Context,
// 	db qrm.DB,
// 	questionID int64,
// ) (*model.QuestionTranslations, error) {
// 	tbl := table.QuestionTranslations

// 	stmt := pg.SELECT(tbl.AllColumns).
// 		FROM(tbl).
// 		WHERE(tbl.QuestionID.EQ(pg.Int64(questionID))).
// 		LIMIT(1)

// 	var trans []model.QuestionTranslations
// 	err := stmt.QueryContext(ctx, db, &trans)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(trans) == 0 {
// 		return nil, common.ErrNotFound
// 	}

// 	return &trans[0], nil
// }

func GetQuestionIDByUUID(
	ctx context.Context,
	db qrm.DB,
	uuid uuid.UUID,
) (*int64, error) {
	tbl := table.Questions
	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(
			tbl.QuestionUUID.EQ(pg.UUID(uuid)),
		)

	var dbID []model.Questions
	err := stmt.QueryContext(ctx, db, &dbID)
	if err != nil {
		return nil, err
	}

	if len(dbID) == 0 {
		return nil, common.ErrQuestionNotFound
	}

	return &dbID[0].ID, nil
}

func GetQuestionOptions(
	ctx context.Context,
	db qrm.DB,
	questionID int64,
) ([]model.QuestionOptions, error) {
	tbl := table.QuestionOptions

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(
			tbl.QuestionID.EQ(pg.Int64(questionID)))

	var options []model.QuestionOptions
	err := stmt.QueryContext(ctx, db, &options)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get question options failed", 0)
	}

	return options, nil
}

func GetQuestionCorrectOptionUUIDs(
	ctx context.Context,
	db qrm.DB,
	questionID int64,
) (*[]uuid.UUID, error) {
	tbl := table.QuestionOptions
	questionTbl := table.Questions

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl.LEFT_JOIN(questionTbl, tbl.QuestionID.EQ(questionTbl.ID))).
		WHERE(
			tbl.IsAnswer.EQ(pg.Bool(true)).AND(questionTbl.ID.EQ(pg.Int64(questionID))),
		)

	var options []model.QuestionOptions
	err := stmt.QueryContext(ctx, db, &options)
	if err != nil {
		return nil, err
	}

	correctOptionUUIDs := make([]uuid.UUID, 0, len(options))
	for _, opt := range options {
		correctOptionUUIDs = append(correctOptionUUIDs, opt.OptionUUID)
	}

	return &correctOptionUUIDs, nil
}

func GetQuestionCorrectOptions(
	ctx context.Context,
	db qrm.DB,
	questionID int64,
) (*[]int64, error) {
	tbl := table.QuestionOptions
	questionTbl := table.Questions

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl.LEFT_JOIN(questionTbl, tbl.QuestionID.EQ(questionTbl.ID))).
		WHERE(
			tbl.IsAnswer.EQ(pg.Bool(true)).AND(questionTbl.ID.EQ(pg.Int64(questionID))),
		)

	var options []model.QuestionOptions
	err := stmt.QueryContext(ctx, db, &options)
	if err != nil {
		return nil, err
	}

	correctOptionIDs := make([]int64, 0, len(options))
	for _, opt := range options {
		correctOptionIDs = append(correctOptionIDs, opt.ID)
	}

	return &correctOptionIDs, nil
}

func GetQuestionOptionTranslations(
	ctx context.Context,
	db qrm.DB,
	optionIDs []int64,
) ([]model.QuestionOptionTranslations, error) {
	tbl := table.QuestionOptionTranslations

	optionIDExpressions := util.BuildInt64Expressions(optionIDs)

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.OptionID.IN(optionIDExpressions...))

	var dto []model.QuestionOptionTranslations
	err := stmt.QueryContext(ctx, db, &dto)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get option translations failed", 0)
	}

	return dto, nil
}

func GetQuestionSubmissions(
	ctx context.Context,
	db qrm.DB,
	questionUUID uuid.UUID,
) (*[]dao.SubmissionWithUserName, error) {
	submissionsTbl := table.QuestionSubmissions
	questionsTbl := table.Questions
	userTbl := table.Users

	stmt := pg.SELECT(
		submissionsTbl.AllColumns,
		userTbl.UserUUID.AS("user_id"),
		userTbl.Nickname.AS("user_name"),
	).FROM(
		submissionsTbl.
			INNER_JOIN(questionsTbl, submissionsTbl.QuestionID.EQ(questionsTbl.ID)).
			INNER_JOIN(userTbl, submissionsTbl.UserID.EQ(userTbl.ID)),
	).WHERE(
		questionsTbl.QuestionUUID.EQ(pg.UUID(questionUUID)),
	).ORDER_BY(
		submissionsTbl.CreatedAt.DESC(),
	)

	var results []struct {
		model.QuestionSubmissions
		UserID   string `db:"user_id"`
		UserName string `db:"user_name"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, err
	}

	daos := make([]dao.SubmissionWithUserName, 0, len(results))
	for _, submission := range results {
		dto := dao.SubmissionWithUserName{
			ID:        submission.ID,
			IsCorrect: submission.IsCorrect,
			CreatedAt: submission.CreatedAt,
			TimeTaken: submission.TimeTaken,
			UserName:  submission.UserName,
		}
		daos = append(daos, dto)
	}

	return &daos, nil
}

func GetQuestionSubmissionsWithOptions(
	ctx context.Context,
	db qrm.DB,
	submissionIDs []int64,
) (*map[int64][]uuid.UUID, error) {
	submissionOptionTbl := table.QuestionSubmissionOptions
	optionTbl := table.QuestionOptions

	submissionExps := util.BuildInt64Expressions(submissionIDs)

	stmt := pg.SELECT(
		submissionOptionTbl.OptionID.AS("option_id"),
		submissionOptionTbl.SubmissionID.AS("submission_id"),
		optionTbl.OptionUUID.AS("option_uuid"),
	).FROM(
		submissionOptionTbl.LEFT_JOIN(optionTbl, submissionOptionTbl.OptionID.EQ(optionTbl.ID)),
	).WHERE(
		submissionOptionTbl.SubmissionID.IN(submissionExps...),
	)

	var results []struct {
		SubmissionID int64     `db:"submission_id"`
		OptionID     int64     `db:"option_id"`
		OptionUUID   uuid.UUID `db:"option_uuid"`
	}
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, err
	}

	// 将选项信息映射到提交
	submissionMap := make(map[int64][]uuid.UUID)
	for _, submission := range results {
		key := submission.SubmissionID
		if _, ok := submissionMap[key]; !ok {
			// 不存在，初始化一个空的切片
			submissionMap[key] = []uuid.UUID{}
		}
		submissionMap[key] = append(submissionMap[key], submission.OptionUUID)
	}

	return &submissionMap, nil
}

func GetQuestionSubmissionCount(
	ctx context.Context,
	db qrm.DB,
	questionID int64,
) (*int64, error) {
	tbl := table.QuestionSubmissions

	stmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(tbl).
		WHERE(tbl.QuestionID.EQ(pg.Int64(questionID)).AND(tbl.IsPractice.EQ(pg.Bool(false))))

	var result struct {
		Count int64 `alias:"count"`
	}
	err := stmt.QueryContext(ctx, db, &result)
	if err != nil {
		return nil, errors.WrapPrefix(err, "get question submissions failed", 0)
	}

	return &result.Count, nil
}

func GetOptionIDsByUUIDs(
	ctx context.Context,
	db qrm.DB,
	optionUUIDs []uuid.UUID,
) ([]int64, error) {
	tbl := table.QuestionOptions

	uuidExpressions := util.BuildUUIDExpressions(optionUUIDs)

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.OptionUUID.IN(uuidExpressions...)).
		ORDER_BY(tbl.ID)

	var optionIDs []model.QuestionOptions
	err := stmt.QueryContext(ctx, db, &optionIDs)
	if err != nil {
		return nil, err
	}

	var result []int64
	for _, option := range optionIDs {
		result = append(result, option.ID)
	}

	return result, nil
}

func BuildQuestionsWithTransaction(
	ctx context.Context,
	db qrm.DB,
	result *dao.QuestionListResult,
) ([]oapi.Question, error) {
	questionIDs := make([]int64, 0)
	for _, q := range result.Questions {
		questionIDs = append(questionIDs, q.Question.ID)
	}
	// 获取翻译
	trans, err := GetQuestionTransByIDs(ctx, db, questionIDs)
	if err != nil {
		return nil, err
	}

	var solvedMap map[int64]bool
	var likedMap map[int64]int16
	// 检查用户是否已解答这些题目（如果用户已登录）
	if userClaims, ok := middleware.GetUserFromContextOnly(ctx); ok {
		solvedMap, err = CheckMultipleQuestionsSolved(
			ctx,
			db,
			userClaims.UserID,
			questionIDs,
		)
		if err != nil {
			return nil, err
		}

		likedMap, err = GetMultipleQuestionsLikeStatus(
			ctx,
			db,
			userClaims.UserID,
			questionIDs,
		)
		if err != nil {
			return nil, err
		}

		// logger.L.Debug("Solved Map:", zap.Any("solvedMap", solvedMap))
	}

	dtos := make([]oapi.Question, 0, len(result.Questions))
	for _, q := range result.Questions {
		id := q.Question.ID
		dtos = append(
			dtos,
			transformer.ConvertSimpleToQuestion(q, trans[id], solvedMap[id], likedMap[id]),
		)
	}

	return dtos, nil
}
