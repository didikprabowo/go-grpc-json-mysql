package repository

import (
	"context"
	"github.com/didikprabowo/go-grpc-json-mysql/model"
)

// ArticleRepository
type ArticleRepository interface {
	Create(ctx context.Context, a model.Article) (int64, error)
	List(ctx context.Context, start, end int64) ([]model.Article, error)
}
