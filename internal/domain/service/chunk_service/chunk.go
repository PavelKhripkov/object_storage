package chunk_service

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/chunk_model"
	"github.com/gofrs/uuid/v5"
	log "github.com/sirupsen/logrus"
	"time"
)

// Service provides methods to manage chunks.
type Service struct {
	storage chunkStorage
	l       *log.Entry
}

// NewChunkService creates new chunk service.
func NewChunkService(chunkStorage chunkStorage, l *log.Logger) *Service {
	return &Service{
		storage: chunkStorage,
		l:       l.WithField("component", "ChunkService"),
	}
}

// Get returns chunk model by ID.
func (s *Service) Get(ctx context.Context, id string) (chunk_model.Chunk, error) {
	res, err := s.storage.Get(ctx, id)
	if err != nil {
		return chunk_model.Chunk{}, err
	}

	return res, nil
}

// Create creates and returns new chunk model.
func (s *Service) Create(ctx context.Context, dto CreateChunkDTO) (chunk_model.Chunk, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return chunk_model.Chunk{}, err
	}

	now := time.Now()

	newChunk := chunk_model.Chunk{
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
		return chunk_model.Chunk{}, err
	}

	return newChunk, nil
}

func (s *Service) Delete(ctx context.Context, chunk chunk_model.Chunk) error {
	// TODO implement
	return nil
}

// GetItemChunks returns chunks of specified Item
func (s *Service) GetItemChunks(ctx context.Context, id string) ([]chunk_model.Chunk, error) {
	chunks, err := s.storage.GetItemChunks(ctx, id)
	if err != nil {
		return nil, err
	}

	return chunks, nil
}
