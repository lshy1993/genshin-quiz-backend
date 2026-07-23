package util

import (
	"github.com/google/uuid"

	pg "github.com/go-jet/jet/v2/postgres"
)

func BuildInt64Expressions(ids []int64) []pg.Expression {
	expressions := make([]pg.Expression, 0, len(ids))
	for _, id := range ids {
		expressions = append(expressions, pg.Int64(id))
	}
	return expressions
}

func BuildStringExpressions(values []string) []pg.Expression {
	expressions := make([]pg.Expression, 0, len(values))
	for _, value := range values {
		expressions = append(expressions, pg.String(value))
	}
	return expressions
}

func BuildUUIDExpressions(uuids []uuid.UUID) []pg.Expression {
	expressions := make([]pg.Expression, 0, len(uuids))
	for _, u := range uuids {
		expressions = append(expressions, pg.UUID(u))
	}
	return expressions
}

func GetDefaultLanguage(language *[]string) string {
	if language != nil && len(*language) > 0 {
		return (*language)[0]
	}
	return "zh" // 默认中文
}

func GetDefaultLanguageFromString(language *string) string {
	if language != nil {
		return *language
	}
	return "zh" // 默认中文
}
