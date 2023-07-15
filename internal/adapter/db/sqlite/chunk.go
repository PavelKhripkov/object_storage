package sqlite

import (
	"context"
	"database/sql"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/chunk"
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

func (s *ChunkStorage) Get(ctx context.Context, id string) (chunk.Chunk, error) {
	return chunk.Chunk{}, nil

	//stmt, err := s.db.PrepareContext(ctx, "SELECT id, name, path, bucket_id FROM item WHERE id = ? LIMIT 1")
	//if err != nil {
	//	return nil, err
	//}
	//defer stmt.Close()
	//
	//newItem := item.Item{}
	//
	//err = stmt.QueryRowContext(ctx, id).Scan(&newItem.ID, &newItem.Name, &newItem.Path, &newItem.BucketId)
	//switch {
	//case err == sql.ErrNoRows:
	//	return nil, ErrNotFound
	//case err != nil:
	//	return nil, err
	//}
	//
	//return &newItem, nil
}

func (s *ChunkStorage) Create(ctx context.Context, chunk chunk.Chunk) error {
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO chunk (id, item_id, position, file_server_id, file_path, size, created, modified) values (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, chunk.ID, chunk.ItemID, chunk.Position, chunk.FileServerID, chunk.FilePath, chunk.Size, chunk.Created.UnixMilli(), chunk.Modified.UnixMilli())
	if err != nil {
		return err
	}

	return nil
}

func (s *ChunkStorage) Delete(ctx context.Context, chunk chunk.Chunk) error {
	return nil
}

func (s *ChunkStorage) GetItemChunks(ctx context.Context, id string) ([]chunk.Chunk, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT id, item_id, position, file_server_id, file_path, size, created, modified FROM chunk WHERE item_id = ? ORDER BY position")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, id)

	res := make([]chunk.Chunk, 0)

	for rows.Next() {
		chnk := chunk.Chunk{}
		var created, modified int64
		if err = rows.Scan(&chnk.ID, &chnk.ItemID, &chnk.Position, &chnk.FileServerID, &chnk.FilePath, &chnk.Size, &created, &modified); err != nil {
			return nil, err
		}
		chnk.Created = time.UnixMilli(created)
		chnk.Modified = time.UnixMilli(modified)
		res = append(res, chnk)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
