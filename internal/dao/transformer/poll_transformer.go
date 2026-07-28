package transformer

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
)

// 将SimplePoll（不包含选项详情）模型转换为 OAPI.
func ConvertSimplePollToDTO(
	poll dao.SimplePoll,
	trans []model.PollTranslations,
	voted bool,
	likeStatus int16,
) oapi.Poll {
	// 构建多语言标题和描述（简单模式只有一种语言）
	title := make(oapi.LocalizedText)
	description := make(oapi.LocalizedText)

	for _, q := range trans {
		description[q.Language] = *q.Description
		title[q.Language] = q.Title
	}

	participants := int(poll.Poll.ParticipantsCount)
	totalVotes := int(poll.Poll.TotalVotesCount)
	likes := int(poll.Poll.LikesCount)
	votesPerUser := int(poll.Poll.VotesPerUser)

	votesPerOption := int(poll.Poll.VotesPerOption)

	expireAt := poll.Poll.ExpiresAt

	likeStatusValue := oapi.LikeStatus(likeStatus)

	return oapi.Poll{
		Id:                poll.Poll.PollUUID,
		Public:            poll.Poll.Public,
		Title:             title,
		Description:       &description,
		Category:          oapi.Category(poll.Poll.Category),
		StartAt:           poll.Poll.StartAt,
		ExpireAt:          expireAt,
		MyVotes:           nil, // 简单模式不返回
		VotesPerUser:      votesPerUser,
		VotesPerOption:    votesPerOption,
		Options:           nil, // 简单模式不返回选项
		ParticipantsCount: participants,
		TotalVotesCount:   totalVotes,
		CreatedBy:         poll.User.UserUUID,
		CreatedAt:         poll.Poll.CreatedAt,
		LikesCount:        likes,
		LikeStatus:        likeStatusValue,
	}
}

func ConvertDetailedVoteToDTO(
	poll dao.DetailedPoll,
	myVotes []oapi.PollVote,
	likeStatus int16,
) oapi.Poll {
	title := make(oapi.LocalizedText)
	description := make(oapi.LocalizedText)

	for _, trans := range poll.Translation {
		title[trans.Language] = trans.Title

		if trans.Description != nil {
			description[trans.Language] = *trans.Description
		}
	}

	options := make([]oapi.PollOption, 0, len(poll.Options))
	for _, opt := range poll.Options {
		// 获取翻译
		translation := poll.OptionTranslations[opt.ID]
		dto := ToPollOption(opt, translation)
		options = append(options, dto)
	}

	participants := int(poll.Poll.ParticipantsCount)
	totalVotes := int(poll.Poll.TotalVotesCount)
	likes := int(poll.Poll.LikesCount)
	votesPerUser := int(poll.Poll.VotesPerUser)

	votesPerOption := int(poll.Poll.VotesPerOption)
	expireAt := poll.Poll.ExpiresAt

	likeStatusValue := oapi.LikeStatus(likeStatus)

	return oapi.Poll{
		Id:                poll.Poll.PollUUID,
		Public:            poll.Poll.Public,
		Title:             title,
		Description:       &description,
		Category:          oapi.Category(poll.Poll.Category),
		StartAt:           poll.Poll.StartAt,
		ExpireAt:          expireAt,
		MyVotes:           myVotes,
		VotesPerUser:      votesPerUser,
		VotesPerOption:    votesPerOption,
		Options:           options,
		ParticipantsCount: participants,
		TotalVotesCount:   totalVotes,
		CreatedBy:         poll.User.UserUUID,
		CreatedAt:         poll.Poll.CreatedAt,
		LikesCount:        likes,
		LikeStatus:        likeStatusValue,
	}
}

func ToPollOption(
	option model.PollOptions,
	translations oapi.LocalizedText,
) oapi.PollOption {
	optionUUID := option.OptionUUID
	votes := int(option.VoteCount)

	dto := oapi.PollOption{
		Id:         optionUUID,
		Text:       translations,
		OptionType: oapi.OptionType("text"), // 默认文本类型，根据实际情况调整
		VotesCount: votes,
	}

	return dto
}
