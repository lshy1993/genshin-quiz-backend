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
	likeStatus := oapi.LikeStatus(userLikeStatus)
	mapedStr := oapi.LocalizedText{
		"zh-CN": "res.Translation.QuestionText",
		"en-US": "Hello",
	}

	return oapi.Question{
		AnswersCount:        &answered,
		Category:            oapi.Category(res.Question.Category),
		CorrectAnswersCount: &correct,
		CreatedAt:           res.Question.CreatedAt,
		CreatedBy:           res.User.UserUUID,
		Difficulty:          oapi.Difficulty(res.Question.Difficulty),
		Explanation:         nil, // 简单模式不返回解释
		Id:                  res.Question.QuestionUUID,
		LikeStatus:          likeStatus,
		LikesCount:          likes,
		Options:             nil, // 简单模式不返回选项
		Public:              res.Question.Public,
		QuestionText:        mapedStr,
		QuestionType:        oapi.QuestionType(res.Question.QuestionType),
		Solved:              solved,
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
	likeStatus := oapi.LikeStatus(userLikeStatus)

	optionMap := make(map[int64]model.QuestionOptions)
	for _, opt := range res.Options {
		optionMap[opt.ID] = opt
	}

	options := make([]oapi.QuestionOption, 0, len(res.Options))
	for _, translation := range res.OptionTranslations {
		dto := ToQuestionOption(optionMap[translation.OptionID], translation, solved)
		options = append(options, dto)
	}

	expStr := oapi.LocalizedText{
		"zh-CN": *res.Translation.Explanation,
	}
	textStr := oapi.LocalizedText{
		"zh-CN": res.Translation.QuestionText,
	}

	return oapi.Question{
		AnswersCount:        &answer,
		Category:            oapi.Category(res.Question.Category),
		CorrectAnswersCount: &correct,
		CreatedAt:           res.Question.CreatedAt,
		CreatedBy:           res.User.UserUUID,
		Difficulty:          oapi.Difficulty(res.Question.Difficulty),
		Explanation:         &expStr,
		Id:                  res.Question.QuestionUUID,
		LikeStatus:          likeStatus,
		LikesCount:          likes,
		Options:             options,
		Public:              res.Question.Public,
		QuestionText:        textStr,
		QuestionType:        oapi.QuestionType(res.Question.QuestionType),
		Solved:              solved,
	}
}

func ToQuestionOption(
	option model.QuestionOptions,
	translation model.QuestionOptionTranslations,
	solved bool,
) oapi.QuestionOption {
	count := int(option.SelectedCount)
	text := oapi.LocalizedText{}
	text[translation.Language] = translation.OptionText

	dto := oapi.QuestionOption{
		Id:            option.OptionUUID,
		MediaUrl:      option.ImgURL,
		Text:          &text,
		SelectedCount: &count,
		OptionType:    oapi.OptionType(option.OptionType),
	}
	if solved {
		dto.IsAnswer = &option.IsAnswer
	}

	return dto
}
