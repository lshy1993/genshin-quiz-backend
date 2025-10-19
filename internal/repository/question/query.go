package question_repo

import (
	"context"
	"strings"

	"genshin-quiz/generated/db/genshinquiz/public/table"
	"genshin-quiz/generated/oapi"
	dao "genshin-quiz/internal/dao"
	"genshin-quiz/internal/dao/transformer"

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
	subTbl := table.QuestionSubmissions
	userTbl := table.Users

	offset := (params.Page - 1) * params.NumPerPage
	if offset < 0 {
		offset = 0
	}

	// 基础查询
	stmt := pg.SELECT(
		tbl.AllColumns,
	).FROM(
		tbl.LEFT_JOIN(
			transTbl,
			tbl.ID.EQ(transTbl.QuestionID),
		).
			LEFT_JOIN(subTbl, tbl.ID.EQ(subTbl.QuestionID)).
			LEFT_JOIN(userTbl, tbl.CreatedBy.EQ(userTbl.ID)),
	).
		WHERE(
			baseCondition(params),
		).ORDER_BY(baseOrder(params)).
		LIMIT(int64(params.NumPerPage)).
		OFFSET(int64(offset))

	// 先获取总数
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).
		FROM(tbl).
		WHERE(baseCondition(params))
	var countResult struct {
		Count int64 `alias:"count"`
	}
	// fmt.Print(countStmt.DebugSql())
	err := countStmt.QueryContext(ctx, db, &countResult)
	if err != nil {
		return nil, err
	}

	var questions []dao.SimpleQuestion
	err = stmt.QueryContext(ctx, db, &questions)
	if err != nil {
		return nil, err
	}

	dtos := make([]oapi.Question, len(questions), 0)
	for _, q := range questions {
		dtos = append(dtos, transformer.ToSimpleQuestion(q))
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
	// 添加语言过滤
	if params.Language != nil && len(*params.Language) > 1 {
		firstLang := []pg.Expression{}
		for _, v := range *params.Language {
			firstLang = append(firstLang, pg.String(v))
		}
		condition = condition.AND(translationTbl.Language.IN(firstLang...))
	}

	// 添加分类过滤
	if params.Category != nil {
		cat := string(*params.Category)
		condition = condition.AND(tbl.Category.EQ(pg.String(cat)))
	}

	// 添加难度过滤
	if params.Difficulty != nil {
		diffExp := []pg.Expression{}
		for _, diff := range *params.Difficulty {
			diffStr := string(diff)
			diffExp = append(diffExp, pg.String(diffStr))
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
) (*oapi.Question, error) {
	tbl := table.Questions
	stmt := pg.SELECT(tbl.AllColumns).FROM(
		tbl,
	).WHERE(
		tbl.QuestionUUID.EQ(pg.UUID(uuid)),
	)

	var result dao.DetailedQuestion
	err := stmt.QueryContext(ctx, db, &result)
	if err != nil {
		return nil, err
	}

	dto := transformer.ToDetailQuestion(result)

	return &dto, nil
}
