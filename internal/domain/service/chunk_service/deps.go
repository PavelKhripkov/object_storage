package chunk_service

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/chunk_model"
)

type chunkStorage interface {
	Get(ctx context.Context, id string) (chunk_model.Chunk, error)
	Create(ctx context.Context, chunk chunk_model.Chunk) error
	Delete(ctx context.Context, chunk chunk_model.Chunk) error
	GetItemChunks(ctx context.Context, id string) ([]chunk_model.Chunk, error)
}
