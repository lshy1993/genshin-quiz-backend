package poll_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"

	"github.com/go-jet/jet/v2/qrm"
)

// GetPollComments 获取投票的评论列表（分页）。
func GetPollComments(
	ctx context.Context,
	db qrm.DB,
	pollID int64,
	limit int,
	offset int,
) (*[]model.PollComments, error) {
	// TODO: 实现评论列表获取
	panic("not implemented")
}

// GetPollCommentCount 获取投票的评论总数。
func GetPollCommentCount(
	ctx context.Context,
	db qrm.DB,
	pollID int64,
) (int64, error) {
	// TODO: 实现评论计数
	panic("not implemented")
}

// GetMultiplePollsCommentsCount 批量获取多个投票的评论数（避免N+1查询）。
func GetMultiplePollsCommentsCount(
	ctx context.Context,
	db qrm.DB,
	pollIDs []int64,
) (map[int64]int64, error) {
	// TODO: 实现批量评论计数
	panic("not implemented")
}

// InsertPollComment 添加投票评论。
func InsertPollComment(
	ctx context.Context,
	db qrm.DB,
	comment model.PollComments,
) (*model.PollComments, error) {
	// TODO: 实现评论插入
	panic("not implemented")
}

// UpdatePollComment 编辑投票评论。
func UpdatePollComment(
	ctx context.Context,
	db qrm.DB,
	commentID int64,
	commentText string,
) error {
	// TODO: 实现评论更新
	panic("not implemented")
}

// DeletePollComment 删除投票评论。
func DeletePollComment(
	ctx context.Context,
	db qrm.DB,
	commentID int64,
) error {
	// TODO: 实现评论删除
	panic("not implemented")
}
