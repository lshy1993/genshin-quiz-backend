package transformer

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
)

func ConvertSimpleToQuestion(
	res dao.SimpleQuestion,
	solved bool,
) oapi.Question {
	answered := 0
	correct := int(res.Question.CorrectCount)
	likes := int(res.Question.Likes)
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
		LikeStatus:   &likeStatus,
		Likes:        &likes,
		Options:      nil,
		Public:       res.Question.Public,
		QuestionText: res.Translation.QuestionText,
		QuestionType: oapi.QuestionType(res.Question.QuestionType),
		Solved:       &solved,
	}
}

func ConvertDetailToQuestion(
	res dao.DetailedQuestion,
) oapi.Question {
	answer := int(res.SubmissionCount)
	correct := int(res.Question.CorrectCount)
	likes := int(res.Question.Likes)
	likeStatus := oapi.QuestionLikeStatus(0) // TODO: 根据用户数据设置喜欢状态
	options := make([]oapi.QuestionOption, 0, len(res.Options))
	solved := false
	for i, translation := range res.OptionTranslations {
		opt := res.Options[i]
		dto := ToQuestionOption(opt, translation)
		options = append(options, dto)
	}

	return oapi.Question{
		AnswerCount:  &answer,
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
		Solved:       &solved,
	}
}

func ToQuestionOption(
	option model.QuestionOptions,
	translation model.OptionTranslations,
) oapi.QuestionOption {
	count := int(option.SelectedCount)
	text := map[string]string{}
	text[translation.Language] = translation.OptionText
	return oapi.QuestionOption{
		Id:    &option.OptionUUID,
		Count: &count,
		Image: option.ImgURL,
		Text:  &text,
		Type:  oapi.QuestionOptionType(option.OptionType),
	}
}
