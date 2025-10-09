package quiz_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

func GetAll(
	ctx context.Context,
	db qrm.DB,
	difficulty string,
	category string,
	offset int,
) ([]model.Quizzes, error) {
	tbl := table.Quizzes
	// Build dynamic WHERE conditions using go-jet
	var conditions []pg.BoolExpression

	if category != "" {
		conditions = append(conditions, tbl.Category.EQ(pg.String(category)))
	}
	if difficulty != "" {
		conditions = append(conditions, tbl.Difficulty.EQ(pg.String(difficulty)))
	}

	// Get total count
	countStmt := pg.SELECT(pg.COUNT(pg.STAR)).FROM(tbl)
	if len(conditions) > 0 {
		countStmt = countStmt.WHERE(pg.AND(conditions...))
	}

	var total int
	err := countStmt.QueryContext(ctx, db, &total)
	if err != nil {
		return nil, err
	}

	// Get quizzes with pagination
	stmt := pg.SELECT(
		tbl.ID,
		tbl.Title,
		tbl.Description,
		tbl.Category,
		tbl.Difficulty,
		tbl.TimeLimit,
		tbl.CreatedBy,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).FROM(
		tbl,
	).ORDER_BY(
		tbl.CreatedAt.DESC(),
	).LIMIT(100).OFFSET(int64(offset))

	if len(conditions) > 0 {
		stmt = stmt.WHERE(pg.AND(conditions...))
	}

	var quizzes []model.Quizzes
	err = stmt.QueryContext(ctx, db, &quizzes)
	if err != nil {
		return nil, err
	}

	return quizzes, nil
}

// TODO: Implement methods using generated models
/*
func (r *QuizRepository) getQuestionsByQuizID(quizID int64) ([]models.Question, error) {
	// This method needs to be updated to work with generated models
	return nil, fmt.Errorf("not implemented")
}

func (r *QuizRepository) Create(req models.CreateQuizRequest) (*models.Quiz, error) {
	// This method needs to be updated to work with generated models
	return nil, fmt.Errorf("not implemented")
}

func (r *QuizRepository) Update(id int64, req models.UpdateQuizRequest) (*models.Quiz, error) {
	// This method needs to be updated to work with generated models
	return nil, fmt.Errorf("not implemented")
}

func (r *QuizRepository) Delete(id int64) error {
	// This method needs to be updated to work with generated models
	return fmt.Errorf("not implemented")
}
*/
