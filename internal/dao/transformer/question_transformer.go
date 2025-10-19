package transformer

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
)

func ToSimpleQuestion(
	res dao.SimpleQuestion,
) oapi.Question {
	answered := 0
	correct := int(res.Question.CorrectCount)
	likes := int(res.Question.Likes)
	solved := res.Solved
	likeStatus := oapi.QuestionLikeStatus(0)
	return oapi.Question{
		AnswerCount:  &answered,
		Category:     oapi.QuestionCategory(res.Question.Category),
		CorrectCount: &correct,
		CreatedAt:    res.Question.CreatedAt,
		CreatedBy:    res.User.UserUUID,
		Difficulty:   oapi.QuestionDifficulty(res.Question.Difficulty),
		Explanation:  nil,
		Id:           res.Question.QuestionUUID,
		Languages:    []string{},
		LikeStatus:   &likeStatus,
		Likes:        &likes,
		Options:      nil,
		Public:       res.Question.Public,
		QuestionText: res.Translation.QuestionText,
		QuestionType: oapi.QuestionType(res.Question.QuestionType),
		Solved:       &solved,
	}
}

func ToDetailQuestion(
	res dao.DetailedQuestion,
) oapi.Question {
	answered := 0
	// var answers []uuid.UUID
	// if res.Solved {
	// 	for _, opt := range res.Options {
	// 		if opt.IsAnswered {
	// 			answers = append(answers, opt.OptionUUID)
	// 		}
	// 	}
	// }
	correct := int(res.Question.CorrectCount)
	likes := int(res.Question.Likes)
	likeStatus := oapi.QuestionLikeStatus(0)
	options := make([]oapi.QuestionOption, 0, len(res.Options))
	for i, translation := range res.OptionTranslations {
		opt := res.Options[i]
		dto := ToQuestionOption(opt, translation)
		options = append(options, dto)
	}
	return oapi.Question{
		AnswerCount:  &answered,
		Category:     oapi.QuestionCategory(res.Question.Category),
		CorrectCount: &correct,
		CreatedAt:    res.Question.CreatedAt,
		CreatedBy:    res.User.UserUUID,
		Difficulty:   oapi.QuestionDifficulty(res.Question.Difficulty),
		Explanation:  res.Translation.Explanation,
		Id:           res.Question.QuestionUUID,
		Languages:    []string{},
		LikeStatus:   &likeStatus,
		Likes:        &likes,
		Options:      options,
		Public:       res.Question.Public,
		QuestionText: res.Translation.QuestionText,
		QuestionType: oapi.QuestionType(res.Question.QuestionType),
		Solved:       &res.Solved,
	}
}

func ToQuestionOption(
	option model.QuestionOptions,
	translation model.OptionTranslations,
) oapi.QuestionOption {
	count := 1
	return oapi.QuestionOption{
		Id:    option.OptionUUID,
		Count: &count,
		Image: option.ImgURL,
		Text:  &translation.OptionText,
		Type:  oapi.QuestionOptionType(option.OptionType),
	}
}
