package sqlite

import (
	"context"
	"database/sql"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/item_model"
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

func (s *ItemStorage) Get(ctx context.Context, id string) (item_model.Item, error) {
	stmt, err := s.db.PrepareContext(
		ctx,
		"SELECT id, name, size, container_id, chunk_count, status, created, modified FROM item WHERE id = ? LIMIT 1",
	)
	if err != nil {
		return item_model.Item{}, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	entity := item_model.Item{}

	var created, modified int64

	err = stmt.QueryRowContext(ctx, id).
		Scan(&entity.ID, &entity.Name, &entity.Size, &entity.ContainerID, &entity.ChunkCount, &entity.Status, &created, &modified)
	switch {
	case err == sql.ErrNoRows:
		return item_model.Item{}, ErrNotFound
	case err != nil:
		return item_model.Item{}, err
	}

	entity.Created = time.UnixMilli(created)
	entity.Modified = time.UnixMilli(modified)

	return entity, nil
}

func (s *ItemStorage) List(ctx context.Context, containerID string) ([]item_model.Item, error) {
	stmt, err := s.db.PrepareContext(
		ctx,
		"SELECT id, name, status, size, created, modified FROM item WHERE container_id = ?",
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	rows, err := stmt.QueryContext(ctx, containerID)

	res := make([]item_model.Item, 0)

	for rows.Next() {
		entity := item_model.Item{}
		var created, modified int64
		if err = rows.Scan(&entity.ID, &entity.Name, &entity.Status, &entity.Size, &created, &modified); err != nil {
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

func (s *ItemStorage) Create(ctx context.Context, item item_model.Item) error {
	stmt, err := s.db.PrepareContext(
		ctx,
		"INSERT INTO item (id, name, container_id, size, chunk_count, status, created, modified) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
	)
	if err != nil {
		return err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	_, err = stmt.ExecContext(
		ctx, item.ID, item.Name, item.ContainerID, item.Size, item.ChunkCount, item.Status, item.Created.UnixMilli(), item.Modified.UnixMilli(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *ItemStorage) Update(ctx context.Context, item item_model.Item) error {
	stmt, err := s.db.PrepareContext(ctx, "UPDATE item SET name=?, container_id=?, chunk_count=?, status=?, modified=? WHERE id = ?")
	if err != nil {
		return err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	modified := time.Now().UnixMilli()

	_, err = stmt.ExecContext(ctx, item.Name, item.ContainerID, item.ChunkCount, item.Status, modified, item.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *ItemStorage) Delete(ctx context.Context, item *item_model.Item) error {
	return nil
}
