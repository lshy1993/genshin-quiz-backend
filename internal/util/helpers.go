package util

import (
	"net/url"
	"strings"

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

func buildAuthLink(domain, basePath, rawToken string) string {
	domain = strings.TrimSuffix(domain, "/")
	u, err := url.Parse(domain + basePath)
	if err != nil {
		// 一般静态路径解析不会报错，如果报错属于程序 Bug
		panic(err)
	}

	q := u.Query()
	q.Set("token", rawToken)
	u.RawQuery = q.Encode()

	return u.String()
}

func GenerateResetLink(domain, rawToken string) string {
	return buildAuthLink(domain, "/reset-password", rawToken)
}

func GenerateEmailVerifyLink(domain, rawToken string) string {
	return buildAuthLink(domain, "/verify-email", rawToken)
}
