package user_repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/db/genshinquiz/public/table"

	"genshin-quiz/internal/common"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserLeaderboardRow struct {
	User          model.Users
	LikesReceived int64
}

type UserLeaderboardResult struct {
	Users []UserLeaderboardRow
	Total int
}

func GetUserByEmail(
	ctx context.Context,
	db qrm.DB,
	email string,
) (*model.Users, error) {
	userTbl := table.Users

	stmt := pg.SELECT(userTbl.AllColumns).
		FROM(userTbl).
		WHERE(
			userTbl.Email.EQ(pg.String(email)),
		)

	var user []model.Users
	err := stmt.QueryContext(ctx, db, &user)
	if len(user) == 0 {
		return nil, common.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user[0], nil
}

func GetUserInfoByID(
	ctx context.Context,
	db qrm.DB,
	id int64,
) (*model.Users, error) {
	tbl := table.Users
	stmt := pg.SELECT(tbl.AllColumns).FROM(tbl).WHERE(
		tbl.ID.EQ(pg.Int64(id)),
	)

	var user model.Users
	err := stmt.QueryContext(ctx, db, &user)
	if err != nil {
		// 如果是记录不存在的错误，返回标准错误
		if err.Error() == "qrm: no rows in result set" {
			return nil, common.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func GetUserInfoByUUID(
	ctx context.Context,
	db qrm.DB,
	uuid uuid.UUID,
) (*model.Users, error) {
	tbl := table.Users
	stmt := pg.SELECT(tbl.AllColumns).FROM(tbl).WHERE(
		tbl.UserUUID.EQ(pg.UUID(uuid)),
	)

	var users []*model.Users
	err := stmt.QueryContext(ctx, db, &users)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, common.ErrUserNotFound
	}

	return users[0], nil
}

func GetUserInfosByUUIDs(
	ctx context.Context,
	db qrm.DB,
	uuids []uuid.UUID,
) ([]*model.Users, error) {
	tbl := table.Users

	uuidExp := make([]pg.Expression, len(uuids))
	for i, id := range uuids {
		uuidExp[i] = pg.UUID(id)
	}
	stmt := pg.SELECT(tbl.AllColumns).FROM(tbl).WHERE(
		tbl.UserUUID.IN(uuidExp...),
	)

	var users []*model.Users
	err := stmt.QueryContext(ctx, db, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUsersLeaderboard(
	ctx context.Context,
	db qrm.DB,
	ids *[]uuid.UUID,
	limit int,
	offset int,
	sortBy string,
	sortDesc bool,
) (*UserLeaderboardResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	orderExpr := buildUserLeaderboardOrderExpr(sortBy)
	orderDirection := "DESC"
	if !sortDesc {
		orderDirection = "ASC"
	}

	whereClauses := []string{"1=1"}
	countWhereClauses := []string{"1=1"}
	args := make([]interface{}, 0)
	countArgs := make([]interface{}, 0)
	argPos := 1

	if ids != nil && len(*ids) > 0 {
		uuidStrings := make([]string, 0, len(*ids))
		for _, id := range *ids {
			uuidStrings = append(uuidStrings, id.String())
		}
		whereClauses = append(whereClauses, fmt.Sprintf("u.user_uuid = ANY($%d::uuid[])", argPos))
		countWhereClauses = append(
			countWhereClauses,
			fmt.Sprintf("u.user_uuid = ANY($%d::uuid[])", argPos),
		)
		args = append(args, pq.Array(uuidStrings))
		countArgs = append(countArgs, pq.Array(uuidStrings))
		argPos++
	}

	if sortBy == "accuracy" || sortBy == "" {
		whereClauses = append(whereClauses, "u.total_submissions > 0")
		countWhereClauses = append(countWhereClauses, "u.total_submissions > 0")
	}

	countSQL := fmt.Sprintf(
		"SELECT COUNT(*) FROM users u WHERE %s",
		strings.Join(countWhereClauses, " AND "),
	)

	var total int
	countRows, err := db.QueryContext(ctx, countSQL, countArgs...)
	if err != nil {
		return nil, err
	}
	defer countRows.Close()
	if countRows.Next() {
		if err := countRows.Scan(&total); err != nil {
			return nil, err
		}
	}
	if err = countRows.Err(); err != nil {
		return nil, err
	}

	querySQL := fmt.Sprintf(`
WITH likes_agg AS (
    SELECT q.created_by AS user_id,
           COUNT(*) FILTER (WHERE ql.value = 1) AS likes_received
    FROM questions q
    LEFT JOIN question_likes ql ON ql.question_id = q.id
    GROUP BY q.created_by
)
SELECT
    u.id,
    u.user_uuid,
    u.email,
    u.display_name,
    u.avatar_url,
    u.location,
    u.timezone,
    u.language,
    u.show_email,
    u.created_at,
    u.updated_at,
    u.total_submissions,
    u.correct_submissions,
    u.questions_created,
    u.total_votes,
    u.user_role,
    COALESCE(la.likes_received, 0) AS likes_received
FROM users u
LEFT JOIN likes_agg la ON la.user_id = u.id
WHERE %s
ORDER BY %s %s, u.id DESC
LIMIT $%d OFFSET $%d`,
		strings.Join(whereClauses, " AND "),
		orderExpr,
		orderDirection,
		argPos,
		argPos+1,
	)

	args = append(args, limit, offset)
	rows, err := db.QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]UserLeaderboardRow, 0, limit)
	for rows.Next() {
		var item UserLeaderboardRow
		var displayName sql.NullString
		var avatarURL sql.NullString
		var location sql.NullString
		var timezone sql.NullString
		var language sql.NullString
		var showEmail sql.NullBool
		var userRole sql.NullInt32

		err = rows.Scan(
			&item.User.ID,
			&item.User.UserUUID,
			&item.User.Email,
			&displayName,
			&avatarURL,
			&location,
			&timezone,
			&language,
			&showEmail,
			&item.User.CreatedAt,
			&item.User.UpdatedAt,
			&item.User.TotalSubmissions,
			&item.User.CorrectSubmissions,
			&item.User.QuestionsCreated,
			&item.User.TotalVotes,
			&userRole,
			&item.LikesReceived,
		)
		if err != nil {
			return nil, err
		}

		if displayName.Valid {
			item.User.DisplayName = &displayName.String
		}
		if avatarURL.Valid {
			item.User.AvatarURL = &avatarURL.String
		}
		if location.Valid {
			item.User.Country = &location.String
		}
		if timezone.Valid {
			item.User.Timezone = &timezone.String
		}
		if language.Valid {
			item.User.Language = &language.String
		}
		if showEmail.Valid {
			item.User.ShowEmail = showEmail.Bool
		}
		if userRole.Valid {
			item.User.UserRole = &userRole.Int32
		}

		results = append(results, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &UserLeaderboardResult{
		Users: results,
		Total: total,
	}, nil
}

func buildUserLeaderboardOrderExpr(sortBy string) string {
	switch sortBy {
	case "votes":
		return "u.total_votes"
	case "questions_created":
		return "u.questions_created"
	case "likes_received":
		return "COALESCE(la.likes_received, 0)"
	case "accuracy", "":
		return `(
((u.correct_submissions::float / NULLIF(u.total_submissions, 0))
 + 3.8416 / (2 * NULLIF(u.total_submissions, 0))
 - 1.96 * sqrt((
     ((u.correct_submissions::float / NULLIF(u.total_submissions, 0))
      * (1 - (u.correct_submissions::float / NULLIF(u.total_submissions, 0))))
     + 3.8416 / (4 * NULLIF(u.total_submissions, 0))
 ) / NULLIF(u.total_submissions, 0)))
 / (1 + 3.8416 / NULLIF(u.total_submissions, 0))
)`
	default:
		return "u.total_votes"
	}
}

func CheckUserExists(
	ctx context.Context,
	db qrm.DB,
	userID int64,
) (bool, error) {
	_, err := GetUserInfoByID(ctx, db, userID)
	if err != nil {
		if errors.Is(err, common.ErrUserNotFound) {
			return false, nil
		}
		// 其他错误返回
		return false, err
	}

	return true, nil
}
