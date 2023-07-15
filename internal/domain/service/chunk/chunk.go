package chunk

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/chunk"
	"github.com/gofrs/uuid/v5"
	log "github.com/sirupsen/logrus"
	"time"
)

type Service struct {
	storage chunkStorage
	l       *log.Entry
}

func NewChunkService(chunkStorage chunkStorage, l *log.Logger) *Service {
	return &Service{
		storage: chunkStorage,
		l:       l.WithField("component", "ChunkService"),
	}
}

func (s *Service) Get(ctx context.Context, id string) (chunk.Chunk, error) {
	res, err := s.storage.Get(ctx, id)
	if err != nil {
		return chunk.Chunk{}, err
	}

	return res, nil
}

func (s *Service) Create(ctx context.Context, dto CreateChunkDTO) (chunk.Chunk, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return chunk.Chunk{}, err
	}

	now := time.Now()

	newChunk := chunk.Chunk{
		ID:           id.String(),
		ItemID:       dto.ItemID,
		Position:     dto.Position,
		FileServerID: dto.FileServerID,
		FilePath:     dto.FilePath,
		Size:         dto.Size,
		Created:      now,
		Modified:     now,
	}

	err = s.storage.Create(ctx, newChunk)
	if err != nil {
		return chunk.Chunk{}, err
	}

	return newChunk, nil
}

func (s *Service) Delete(ctx context.Context, chunk chunk.Chunk) error {
	return nil
}

func (s *Service) GetItemChunks(ctx context.Context, id string) ([]chunk.Chunk, error) {
	chunks, err := s.storage.GetItemChunks(ctx, id)
	if err != nil {
		return nil, err
	}

	return chunks, nil
}
