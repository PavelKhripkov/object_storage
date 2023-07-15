package chunk

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/chunk"
)

type chunkStorage interface {
	Get(ctx context.Context, id string) (chunk.Chunk, error)
	Create(ctx context.Context, chunk chunk.Chunk) error
	Delete(ctx context.Context, chunk chunk.Chunk) error
	GetItemChunks(ctx context.Context, id string) ([]chunk.Chunk, error)
}
