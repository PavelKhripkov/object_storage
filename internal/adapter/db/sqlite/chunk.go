package sqlite

import (
	"context"
	"database/sql"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/chunk_model"
	log "github.com/sirupsen/logrus"
	"time"
)

type ChunkStorage struct {
	db *sql.DB
	l  *log.Entry
}

func NewChunkStorage(db *sql.DB, l *log.Logger) *ChunkStorage {
	return &ChunkStorage{
		db: db,
		l:  l.WithField("component", "ChunkStorage"),
	}
}

func (s *ChunkStorage) Get(ctx context.Context, id string) (chunk_model.Chunk, error) {
	stmt, err := s.db.PrepareContext(
		ctx,
		"SELECT id, item_id, position, file_server_id, file_path, size, created, modified FROM chunk WHERE id = ? LIMIT 1",
	)
	if err != nil {
		return chunk_model.Chunk{}, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	entity := chunk_model.Chunk{}

	var created, modified int64

	err = stmt.QueryRowContext(ctx, id).
		Scan(&entity.ID, &entity.ItemID, &entity.Position, &entity.FileServerID, &entity.FilePath, &entity.Size, &created, &modified)
	switch {
	case err == sql.ErrNoRows:
		return chunk_model.Chunk{}, ErrNotFound
	case err != nil:
		return chunk_model.Chunk{}, err
	}

	entity.Created = time.UnixMilli(created)
	entity.Modified = time.UnixMilli(modified)

	return entity, nil
}

func (s *ChunkStorage) Create(ctx context.Context, chunk chunk_model.Chunk) error {
	stmt, err := s.db.PrepareContext(
		ctx,
		"INSERT INTO chunk (id, item_id, position, file_server_id, file_path, size, created, modified) values (?, ?, ?, ?, ?, ?, ?, ?)",
	)
	if err != nil {
		return err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	_, err = stmt.ExecContext(ctx, chunk.ID, chunk.ItemID, chunk.Position, chunk.FileServerID, chunk.FilePath, chunk.Size, chunk.Created.UnixMilli(), chunk.Modified.UnixMilli())
	if err != nil {
		return err
	}

	return nil
}

func (s *ChunkStorage) Delete(ctx context.Context, chunk chunk_model.Chunk) error {
	return nil
}

func (s *ChunkStorage) GetItemChunks(ctx context.Context, id string) ([]chunk_model.Chunk, error) {
	stmt, err := s.db.PrepareContext(
		ctx,
		"SELECT id, item_id, position, file_server_id, file_path, size, created, modified FROM chunk WHERE item_id = ? ORDER BY position",
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	rows, err := stmt.QueryContext(ctx, id)

	res := make([]chunk_model.Chunk, 0)

	for rows.Next() {
		entity := chunk_model.Chunk{}
		var created, modified int64
		if err = rows.Scan(&entity.ID, &entity.ItemID, &entity.Position, &entity.FileServerID, &entity.FilePath, &entity.Size, &created, &modified); err != nil {
			return nil, err
		}
		entity.Created = time.UnixMilli(created)
		entity.Modified = time.UnixMilli(modified)
		res = append(res, entity)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
