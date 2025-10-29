package question_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	"github.com/go-jet/jet/v2/qrm"
)

func InsertQuestion(
	ctx context.Context,
	db qrm.DB,
	insertModel model.Questions,
) (*model.Questions, error) {
	tbl := table.Questions
	insertStmt := tbl.INSERT(
		tbl.QuestionUUID,
		tbl.Public,
		tbl.QuestionType,
		tbl.Category,
		tbl.Difficulty,
		tbl.IsPublished,
		tbl.PublishedAt,
		tbl.CreatedBy,
		tbl.CreatedAt,
	).MODEL(insertModel).
		RETURNING(tbl.AllColumns)

	var question model.Questions
	err := insertStmt.QueryContext(ctx, db, &question)
	if err != nil {
		return nil, err
	}

	return &question, nil
}

func InsertQuestionTranslations(
	ctx context.Context,
	db qrm.DB,
	translations []model.QuestionTranslations,
) error {
	if len(translations) == 0 {
		return nil
	}

	tbl := table.QuestionTranslations
	insertStmt := tbl.INSERT(
		tbl.QuestionID,
		tbl.Language,
		tbl.QuestionText,
		tbl.Explanation,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).MODELS(translations)

	_, err := insertStmt.ExecContext(ctx, db)
	return err
}

func InsertQuestionOptions(
	ctx context.Context,
	db qrm.DB,
	options []model.QuestionOptions,
) (*[]model.QuestionOptions, error) {
	if len(options) == 0 {
		return nil, nil
	}

	tbl := table.QuestionOptions
	insertStmt := tbl.INSERT(
		tbl.QuestionID,
		tbl.OptionUUID,
		tbl.OptionType,
		tbl.ImgURL,
		tbl.CreatedAt,
	).MODELS(options).RETURNING(tbl.AllColumns)

	var insertedOptions []model.QuestionOptions
	err := insertStmt.QueryContext(ctx, db, &insertedOptions)
	if err != nil {
		return nil, err
	}

	return &insertedOptions, nil
}

func InsertOptionTranslations(
	ctx context.Context,
	db qrm.DB,
	optionTranslations []model.OptionTranslations,
) error {
	if len(optionTranslations) == 0 {
		return nil
	}

	tbl := table.OptionTranslations
	insertStmt := tbl.INSERT(
		tbl.OptionID,
		tbl.Language,
		tbl.OptionText,
		tbl.CreatedAt,
		tbl.UpdatedAt,
	).MODELS(optionTranslations)

	_, err := insertStmt.ExecContext(ctx, db)
	return err
}
