package transformer

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
)

func ConvertSimpleToQuestion(
	res dao.SimpleQuestion,
	solved bool,
	userLikeStatus int16,
) oapi.Question {
	answered := int(res.Question.SubmitCount)
	correct := int(res.Question.CorrectCount)
	likes := int(res.Question.Likes)
	likeStatus := oapi.QuestionLikeStatus(userLikeStatus)

	return oapi.Question{
		AnswerCount:  &answered,
		Category:     oapi.QuestionCategory(res.Question.Category),
		CorrectCount: &correct,
		CreatedAt:    res.Question.CreatedAt,
		CreatedBy:    res.User.UserUUID,
		Difficulty:   oapi.QuestionDifficulty(res.Question.Difficulty),
		Explanation:  nil, // 简单模式不返回解释
		Id:           res.Question.QuestionUUID,
		LikeStatus:   &likeStatus,
		Likes:        &likes,
		Options:      nil, // 简单模式不返回选项
		Public:       res.Question.Public,
		QuestionText: res.Translation.QuestionText,
		QuestionType: oapi.QuestionType(res.Question.QuestionType),
		Solved:       &solved,
	}
}

func ConvertDetailToQuestion(
	res dao.DetailedQuestion,
	solved bool,
	userLikeStatus int16,
) oapi.Question {
	answer := int(res.SubmissionCount)
	correct := int(res.Question.CorrectCount)
	likes := int(res.Question.Likes)
	likeStatus := oapi.QuestionLikeStatus(userLikeStatus)

	optionMap := make(map[int64]model.QuestionOptions)
	for _, opt := range res.Options {
		optionMap[opt.ID] = opt
	}

	options := make([]oapi.QuestionOption, 0, len(res.Options))
	languageSet := make(map[string]bool)
	for _, translation := range res.OptionTranslations {
		dto := ToQuestionOption(optionMap[translation.OptionID], translation, solved)
		options = append(options, dto)
		languageSet[translation.Language] = true
	}

	// 将 map 转换为 slice
	languages := make([]string, 0, len(languageSet))
	for lang := range languageSet {
		languages = append(languages, lang)
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
		Languages:    languages,
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
	translation model.QuestionOptionTranslations,
	solved bool,
) oapi.QuestionOption {
	count := int(option.SelectedCount)
	text := map[string]string{}
	text[translation.Language] = translation.OptionText

	dto := oapi.QuestionOption{
		Id:    &option.OptionUUID,
		Count: &count,
		Image: option.ImgURL,
		Text:  &text,
		Type:  oapi.QuestionOptionType(option.OptionType),
	}
	if solved {
		dto.IsAnswer = &option.IsAnswer
	}

	return dto
}
