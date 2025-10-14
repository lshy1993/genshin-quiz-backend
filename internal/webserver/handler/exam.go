package handler

import (
	"context"
	"genshin-quiz/generated/oapi"
)

func GetExams(ctx context.Context, req oapi.GetExamsRequestObject) (oapi.GetExamsResponseObject, error) {
	return (oapi.GetExams200JSONResponse{}), nil
}
