package sqlite

import (
	"context"
	"database/sql"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/container_model"
	log "github.com/sirupsen/logrus"
	"time"
)

type ContainerStorage struct {
	db *sql.DB
	l  *log.Entry
}

func NewContainerStorage(db *sql.DB, l *log.Logger) *ContainerStorage {
	return &ContainerStorage{
		db: db,
		l:  l.WithField("component", "ContainerStorage"),
	}
}

func (s ContainerStorage) Get(ctx context.Context, id string) (container_model.Container, error) {
	stmt, err := s.db.PrepareContext(
		ctx,
		"SELECT id, name, description, parent_id, created, modified FROM container WHERE id = ? LIMIT 1",
	)
	if err != nil {
		return container_model.Container{}, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	entity := container_model.Container{}

	var created, modified int64

	err = stmt.QueryRowContext(ctx, id).
		Scan(&entity.ID, &entity.Name, &entity.Description, &entity.ParentID, &created, &modified)
	switch {
	case err == sql.ErrNoRows:
		return container_model.Container{}, ErrNotFound
	case err != nil:
		return container_model.Container{}, err
	}

	entity.Created = time.UnixMilli(created)
	entity.Modified = time.UnixMilli(modified)

	return entity, nil
}

func (s ContainerStorage) List(ctx context.Context) ([]container_model.Container, error) {
	stmt, err := s.db.PrepareContext(
		ctx,
		"SELECT id, name, description, parent_id, created, modified FROM container",
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	rows, err := stmt.QueryContext(ctx)

	res := make([]container_model.Container, 0)

	for rows.Next() {
		entity := container_model.Container{}
		var created, modified int64
		if err = rows.Scan(&entity.ID, &entity.Name, &entity.Description, &entity.ParentID, &created, &modified); err != nil {
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

func (s ContainerStorage) Create(ctx context.Context, container container_model.Container) error {
	stmt, err := s.db.PrepareContext(
		ctx,
		"INSERT INTO container (id, name, description, parent_id, created, modified) VALUES (?, ?, ?, ?, ?, ?)",
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
		ctx, container.ID, container.Name, container.Description, container.ParentID, container.Created.UnixMilli(), container.Modified.UnixMilli(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s ContainerStorage) Delete(ctx context.Context, id string) error {
	// TODO implement.
	return nil
}
