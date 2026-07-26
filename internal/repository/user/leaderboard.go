package user_repo

import (
	"context"
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"
	"genshin-quiz/internal/dao"
	"genshin-quiz/internal/enum"

	"github.com/go-errors/errors"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

const (
	defaultLeaderboardLimit = 10
	maxLeaderboardLimit     = 100
)

func GetUsersLeaderboard(
	ctx context.Context,
	db qrm.DB,
	params dao.LeaderboardParams,
) ([]dao.LeaderboardRow, int, error) {
	if params.Limit <= 0 {
		params.Limit = defaultLeaderboardLimit
	}
	if params.Limit > maxLeaderboardLimit {
		params.Limit = maxLeaderboardLimit
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	users := table.Users
	profiles := table.UserProfiles
	privacies := table.UserPrivacies
	stats := table.UserStats

	orderCol := buildOrderByColumn(params.SortBy)
	orderClause := orderCol.ASC()
	if params.SortDesc {
		orderClause = orderCol.DESC()
	}
	// accuracy 排序时，只统计有过答题记录的用户，避免全是 0/0 的用户挤占榜单
	condition := postgres.Bool(true)
	if params.SortBy == enum.SortByAccuracy || params.SortBy == "" {
		condition = stats.TotalSubmissions.GT(postgres.Int(0))
	}

	stmt := postgres.SELECT(
		users.AllColumns,
		profiles.AllColumns,
		privacies.AllColumns,
		stats.AllColumns,
		postgres.RawInt("COUNT(*) OVER()").AS("total_count"),
	).FROM(
		users.
			INNER_JOIN(profiles, profiles.UserID.EQ(users.ID)).
			INNER_JOIN(privacies, privacies.UserID.EQ(users.ID)).
			INNER_JOIN(stats, stats.UserID.EQ(users.ID)),
	).WHERE(
		condition,
	).ORDER_BY(
		orderClause,
		users.ID.DESC(), // 排序值相同时的稳定 tie-breaker，避免分页错乱
	).LIMIT(int64(params.Limit)).
		OFFSET(int64(params.Offset))

	var rawResults []struct {
		model.Users
		model.UserProfiles
		model.UserPrivacies
		model.UserStats
		TotalCount int `alias:"total_count"`
	}

	err := stmt.QueryContext(ctx, db, &rawResults)
	if err != nil {
		return nil, 0, errors.WrapPrefix(err, "query users leaderboard failed", 0)
	}

	total := 0
	rows := make([]dao.LeaderboardRow, 0, len(rawResults))
	for _, r := range rawResults {
		total = r.TotalCount
		rows = append(rows, dao.LeaderboardRow{
			User:    r.Users,
			Profile: r.UserProfiles,
			Privacy: r.UserPrivacies,
			Stats:   r.UserStats,
		})
	}

	return rows, total, nil
}

func buildOrderByColumn(sortBy enum.LeaderboardSortBy) postgres.Expression {
	stats := table.UserStats
	switch sortBy {
	case enum.SortByVotesCast:
		return stats.VotesCast
	case enum.SortByQuestionsCreated:
		return stats.QuestionsCreated
	case enum.SortByLikesReceived:
		return stats.LikesReceived
	case enum.SortByPollsCreated:
		return stats.PollsCreated
	case enum.SortByAccuracy, "":
		// Wilson score 置信区间下界，比单纯正确率更抗"小样本刷分"
		return postgres.RawFloat(`(
((us.correct_submissions::float / NULLIF(us.total_submissions, 0))
 + 3.8416 / (2 * NULLIF(us.total_submissions, 0))
 - 1.96 * sqrt((
     ((us.correct_submissions::float / NULLIF(us.total_submissions, 0))
      * (1 - (us.correct_submissions::float / NULLIF(us.total_submissions, 0))))
     + 3.8416 / (4 * NULLIF(us.total_submissions, 0))
 ) / NULLIF(us.total_submissions, 0)))
 / (1 + 3.8416 / NULLIF(us.total_submissions, 0))
)`)
	default:
		return stats.VotesCast
	}
}
