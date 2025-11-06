package question_repo

import (
	"context"
	"strings"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"
	"genshin-quiz/generated/oapi"
	dao "genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"

	"genshin-quiz/internal/common"

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
	transTbl := table.QuestionTranslations
	userTbl := table.Users

	offset := (params.Page - 1) * params.NumPerPage
	if offset < 0 {
		offset = 0
	}

	// 基础查询 - 使用语言过滤的JOIN
	defaultLang := "zh" // 默认语言
	if params.Language != nil {
		defaultLang = (*params.Language)[0] // 使用第一个指定语言
	}

	stmt := pg.SELECT(
		tbl.AllColumns,
		transTbl.AllColumns,
		userTbl.AllColumns,
	).FROM(
		tbl.LEFT_JOIN(
			transTbl,
			tbl.ID.EQ(transTbl.QuestionID).AND(transTbl.Language.EQ(pg.String(defaultLang))),
		).
			LEFT_JOIN(userTbl, tbl.CreatedBy.EQ(userTbl.ID)),
	).
		WHERE(
			baseCondition(params),
		).
		ORDER_BY(baseOrder(params)).
		LIMIT(int64(params.NumPerPage)).
		OFFSET(int64(offset))

	// 先获取总数
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(tbl).
		WHERE(baseCondition(params))
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

	dtos := make([]oapi.Question, 0, len(questions))
	for _, q := range questions {
		dtos = append(dtos, transformer.ConvertSimpleToQuestion(q, false, 0))
	}

	return &dao.QuestionListResult{
		Questions: dtos,
		Total:     int(countResult.Count),
	}, nil
}

func baseCondition(params dao.QuestionListParams) pg.BoolExpression {
	tbl := table.Questions
	translationTbl := table.QuestionTranslations
	condition := tbl.Public.IS_TRUE().AND(tbl.IsPublished.IS_TRUE())

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
		condition = condition.AND(pg.LOWER(translationTbl.QuestionText).LIKE(pg.String(searchTerm)))
	}
	return condition
}

func baseOrder(params dao.QuestionListParams) pg.OrderByClause {
	tbl := table.Questions
	switch *params.SortBy {
	case "PublishDate": // 上线时间
		if params.SortDesc {
			return tbl.PublishedAt.DESC()
		}
		return tbl.PublishedAt.ASC()
	case "Difficulty":
		// 难度排序：easy < medium < hard
		difficultyOrder := pg.CASE().
			WHEN(tbl.Difficulty.EQ(pg.String("easy"))).THEN(pg.Int(1)).
			WHEN(tbl.Difficulty.EQ(pg.String("medium"))).THEN(pg.Int(2)).
			WHEN(tbl.Difficulty.EQ(pg.String("hard"))).THEN(pg.Int(3)).
			ELSE(pg.Int(0))
		if params.SortDesc {
			return difficultyOrder.DESC()
		}
		return difficultyOrder.ASC()
	case "Likes": // 点赞数
		if params.SortDesc {
			return tbl.Likes.DESC()
		}
		return tbl.Likes.ASC()
	case "Submissions": // 参与人数
		if params.SortDesc {
			return tbl.SubmitCount.DESC()
		}
		return tbl.SubmitCount.ASC()
	case "CorrectRate":
		if params.SortDesc {
			return tbl.CorrectCount.DESC()
		}
		return tbl.CorrectCount.ASC()
	default:
		if params.SortDesc {
			return tbl.PublishedAt.DESC()
		}
		return tbl.PublishedAt.ASC()
	}
}

func GetQuestionByUUID(
	ctx context.Context,
	db qrm.DB,
	uuid uuid.UUID,
	language *string,
) (*dao.SimpleQuestion, error) {
	tbl := table.Questions
	transTbl := table.QuestionTranslations
	userTbl := table.Users

	defaultLang := "zh" // 默认语言
	if language != nil {
		defaultLang = *language // 使用指定语言
	}

	stmt := pg.SELECT(
		tbl.AllColumns,
		transTbl.AllColumns,
		userTbl.UserUUID,
	).FROM(
		tbl.LEFT_JOIN(transTbl,
			tbl.ID.EQ(transTbl.QuestionID).AND(transTbl.Language.EQ(pg.String(defaultLang))),
		).LEFT_JOIN(userTbl, userTbl.ID.EQ(tbl.CreatedBy)),
	).WHERE(
		tbl.QuestionUUID.EQ(pg.UUID(uuid)),
	)

	var result []dao.SimpleQuestion
	err := stmt.QueryContext(ctx, db, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, common.ErrQuestionNotFound
	}

	return &result[0], nil
}

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

func GetQuestionTranslation(
	ctx context.Context,
	db qrm.DB,
	questionID int64,
	language *[]string,
) (*[]model.QuestionTranslations, error) {
	tbl := table.QuestionTranslations

	// langExp := []pg.Expression{}
	// for _, lang := range *language {
	// 	langExp = append(langExp, pg.String(string(lang)))
	// }

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.QuestionID.EQ(pg.Int64(questionID)))

	var dto []model.QuestionTranslations
	err := stmt.QueryContext(ctx, db, &dto)
	if err != nil {
		return nil, err
	}

	return &dto, nil
}

func GetQuestionOptions(
	ctx context.Context,
	db qrm.DB,
	questionID int64,
) (*[]model.QuestionOptions, error) {
	tbl := table.QuestionOptions

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(
			tbl.QuestionID.EQ(pg.Int64(questionID)))

	var options []model.QuestionOptions
	err := stmt.QueryContext(ctx, db, &options)
	if err != nil {
		return nil, err
	}

	return &options, nil
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
	language *[]string,
) (*[]model.OptionTranslations, error) {
	tbl := table.OptionTranslations

	// 直接构建 IN 表达式
	optionIDExpressions := make([]pg.Expression, 0, len(optionIDs))
	for _, id := range optionIDs {
		optionIDExpressions = append(optionIDExpressions, pg.Int64(id))
	}

	stmt := pg.SELECT(tbl.AllColumns).
		FROM(tbl).
		WHERE(tbl.OptionID.IN(optionIDExpressions...))

	var dto []model.OptionTranslations
	err := stmt.QueryContext(ctx, db, &dto)
	if err != nil {
		return nil, err
	}

	return &dto, nil
}

func GetQuestionSubmissions(
	ctx context.Context,
	db qrm.DB,
	questionUUID uuid.UUID,
) (*[]model.QuestionSubmissions, error) {
	submissionsTbl := table.QuestionSubmissions
	questionsTbl := table.Questions

	stmt := pg.SELECT(
		submissionsTbl.AllColumns,
	).FROM(
		submissionsTbl.
			INNER_JOIN(questionsTbl, submissionsTbl.QuestionID.EQ(questionsTbl.ID)),
	).WHERE(
		questionsTbl.QuestionUUID.EQ(pg.UUID(questionUUID)),
	).ORDER_BY(
		submissionsTbl.CreatedAt.DESC(),
	)

	var results []model.QuestionSubmissions
	err := stmt.QueryContext(ctx, db, &results)
	if err != nil {
		return nil, err
	}
	return &results, nil
}

func GetQuestionSubmissionsWithOptions(
	ctx context.Context,
	db qrm.DB,
	submissionIDs []int64,
) (*map[int64][]uuid.UUID, error) {
	submissionOptionTbl := table.QuestionSubmissionOptions
	optionTbl := table.QuestionOptions

	submissionExps := make([]pg.Expression, 0, len(submissionIDs))
	for _, id := range submissionIDs {
		submissionExps = append(submissionExps, pg.Int64(id))
	}

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
		return nil, err
	}

	return &result.Count, nil
}

func GetOptionIDsByUUIDs(
	ctx context.Context,
	db qrm.DB,
	optionUUIDs []uuid.UUID,
) ([]int64, error) {
	tbl := table.QuestionOptions

	uuidExpressions := make([]pg.Expression, 0, len(optionUUIDs))
	for _, optionUUID := range optionUUIDs {
		uuidExpressions = append(uuidExpressions, pg.UUID(optionUUID))
	}

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
