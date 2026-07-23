package question_repo

import (
	"context"

	"genshin-quiz/generated/db/genshinquiz/public/model"

	"github.com/go-jet/jet/v2/qrm"
)

// GetQuestionComments 获取问题的评论列表（分页）。
func GetQuestionComments(
	ctx context.Context,
	db qrm.DB,
	questionID int64,
	limit int,
	offset int,
) (*[]model.QuestionComments, error) {
	// TODO: 实现评论列表获取
	panic("not implemented")
}

// GetQuestionCommentCount 获取问题的评论总数。
func GetQuestionCommentCount(
	ctx context.Context,
	db qrm.DB,
	questionID int64,
) (int64, error) {
	// TODO: 实现评论计数
	panic("not implemented")
}

// GetMultipleQuestionsCommentsCount 批量获取多个问题的评论数（避免N+1查询）。
func GetMultipleQuestionsCommentsCount(
	ctx context.Context,
	db qrm.DB,
	questionIDs []int64,
) (map[int64]int64, error) {
	// TODO: 实现批量评论计数
	panic("not implemented")
}

// InsertQuestionComment 添加问题评论。
func InsertQuestionComment(
	ctx context.Context,
	db qrm.DB,
	comment model.QuestionComments,
) (*model.QuestionComments, error) {
	// TODO: 实现评论插入
	panic("not implemented")
}

// UpdateQuestionComment 编辑问题评论。
func UpdateQuestionComment(
	ctx context.Context,
	db qrm.DB,
	commentID int64,
	commentText string,
) error {
	// TODO: 实现评论更新
	panic("not implemented")
}

// DeleteQuestionComment 删除问题评论。
func DeleteQuestionComment(
	ctx context.Context,
	db qrm.DB,
	commentID int64,
) error {
	// TODO: 实现评论删除
	panic("not implemented")
}
