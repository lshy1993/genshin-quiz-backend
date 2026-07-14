package vote_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"

	"github.com/go-jet/jet/v2/qrm"
)

// GetVoteComments 获取投票的评论列表（分页）。
func GetVoteComments(
	ctx context.Context,
	db qrm.DB,
	voteID int64,
	limit int,
	offset int,
) (*[]model.VoteComments, error) {
	// TODO: 实现评论列表获取
	panic("not implemented")
}

// GetVoteCommentCount 获取投票的评论总数。
func GetVoteCommentCount(
	ctx context.Context,
	db qrm.DB,
	voteID int64,
) (int64, error) {
	// TODO: 实现评论计数
	panic("not implemented")
}

// GetMultipleVotesCommentsCount 批量获取多个投票的评论数（避免N+1查询）。
func GetMultipleVotesCommentsCount(
	ctx context.Context,
	db qrm.DB,
	voteIDs []int64,
) (map[int64]int64, error) {
	// TODO: 实现批量评论计数
	panic("not implemented")
}

// InsertVoteComment 添加投票评论。
func InsertVoteComment(
	ctx context.Context,
	db qrm.DB,
	comment model.VoteComments,
) (*model.VoteComments, error) {
	// TODO: 实现评论插入
	panic("not implemented")
}

// UpdateVoteComment 编辑投票评论。
func UpdateVoteComment(
	ctx context.Context,
	db qrm.DB,
	commentID int64,
	commentText string,
) error {
	// TODO: 实现评论更新
	panic("not implemented")
}

// DeleteVoteComment 删除投票评论。
func DeleteVoteComment(
	ctx context.Context,
	db qrm.DB,
	commentID int64,
) error {
	// TODO: 实现评论删除
	panic("not implemented")
}
