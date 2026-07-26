package enum

type TokenType string

//nolint:gosec // 常量名包含 Token 触发了 gosec G101 误报
const (
	TokenTypePasswordReset TokenType = "password_reset"
	TokenTypeEmailVerify   TokenType = "email_verification"
	TokenTypeMagicLink     TokenType = "magic_link"
)

func (t TokenType) String() string {
	return string(t)
}

type Environment string

const (
	DEV  Environment = "develop"
	TEST Environment = "testing"
	// STG  Environment = "staging".
	PROD Environment = "production"
)

func (t Environment) String() string {
	return string(t)
}

type LoginStatus int16

const (
	LoginStatusSuccess LoginStatus = 0
	LoginStatusFailed  LoginStatus = 1
	LoginStatusBlocked LoginStatus = 2
)

type LoginProvider string

const (
	LoginPassword LoginProvider = "password"
)

type LeaderboardSortBy string

const (
	SortByAccuracy         LeaderboardSortBy = "accuracy"
	SortByVotesCast        LeaderboardSortBy = "votes_cast"
	SortByQuestionsCreated LeaderboardSortBy = "questions_created"
	SortByLikesReceived    LeaderboardSortBy = "likes_received"
	SortByPollsCreated     LeaderboardSortBy = "polls_created"
)
