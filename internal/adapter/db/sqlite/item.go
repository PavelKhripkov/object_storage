package sqlite

import (
	"context"
	"database/sql"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/item"
	log "github.com/sirupsen/logrus"
	"time"
)

type ItemStorage struct {
	db *sql.DB
	l  *log.Entry
}

func NewItemStorage(db *sql.DB, l *log.Logger) *ItemStorage {
	return &ItemStorage{
		db: db,
		l:  l.WithField("component", "ItemStorage"),
	}
}

func (s *ItemStorage) Get(ctx context.Context, id string) (*item.Item, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT id, name, size, path, bucket_id, chunk_count, status, created, modified FROM item WHERE id = ? LIMIT 1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	newItem := item.Item{}

	var created, modified int64

	err = stmt.QueryRowContext(ctx, id).Scan(&newItem.ID, &newItem.Name, &newItem.Size, &newItem.Path, &newItem.BucketId, &newItem.ChunkCount, &newItem.Status, &created, &modified)
	switch {
	case err == sql.ErrNoRows:
		return nil, ErrNotFound
	case err != nil:
		return nil, err
	}

	newItem.Created = time.UnixMilli(created)
	newItem.Modified = time.UnixMilli(modified)

	return &newItem, nil
}

func (s *ItemStorage) GetAll(ctx context.Context) []*item.Item {
	return nil
}

func (s *ItemStorage) Create(ctx context.Context, item item.Item) error {
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO item (id, name, path, size, bucket_id, chunk_count, status, created, modified) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, item.ID, item.Name, item.Path, item.Size, item.BucketId, item.ChunkCount, item.Status, item.Created.UnixMilli(), item.Modified.UnixMilli())
	if err != nil {
		return err
	}

	return nil
}

func (s *ItemStorage) Update(ctx context.Context, item item.Item) error {
	stmt, err := s.db.PrepareContext(ctx, "UPDATE item SET name=?, path=?, bucket_id=?, chunk_count=?, status=? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, item.Name, item.Path, item.BucketId, item.ChunkCount, item.Status, item.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *ItemStorage) Delete(ctx context.Context, item *item.Item) error {
	return nil
}
