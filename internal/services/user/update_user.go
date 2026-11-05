package services

import (
	"genshin-quiz/generated/oapi"
	"genshin-quiz/logger"

	"go.uber.org/zap"
)

func DTOToGender(g oapi.Gender) int16 {
	switch g {
	case oapi.GenderUnknown:
		return 0
	case oapi.GenderMale:
		return 1
	case oapi.GenderFemale:
		return 2
	case oapi.GenderOther:
		return 3
	default:
		// 理论上不该出现，记录一下方便排查脏数据
		logger.L.Warn("unexpected gender value", zap.String("gender", string(g)))
		return 0
	}
}
