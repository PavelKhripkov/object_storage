package sqlite

import (
	"context"
	"database/sql"
	log "github.com/sirupsen/logrus"
	"strings"
)

type FileServerStorage struct {
	db *sql.DB
	l  *log.Entry
}

func NewFileServerStorage(db *sql.DB, l *log.Logger) *FileServerStorage {
	return &FileServerStorage{
		db: db,
		l:  l.WithField("component", "FileServerStorage"),
	}
}

func (s *FileServerStorage) Add(ctx context.Context, params CommonFileServerDTO) error {
	stmt, err := s.db.PrepareContext(ctx, "INSERT INTO file_server (id, name, type, params, created, modified) values (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	_, err = stmt.ExecContext(ctx, params.ID, params.Name, params.Type, params.Params, params.Created.UnixMilli(), params.Modified.UnixMilli())
	if err != nil {
		return err
	}

	return nil
}

func (s *FileServerStorage) Get(ctx context.Context, id string) (CommonFileServerDTO, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT id, name, type, params FROM file_server WHERE id = ? LIMIT 1")
	if err != nil {
		return CommonFileServerDTO{}, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	entity := CommonFileServerDTO{}
	err = stmt.QueryRowContext(ctx, id).Scan(&entity.ID, &entity.Name, &entity.Type, &entity.Params)

	switch {
	case err == sql.ErrNoRows:
		return CommonFileServerDTO{}, ErrNotFound
	case err != nil:
		return CommonFileServerDTO{}, err
	}

	return entity, nil
}

func (s *FileServerStorage) ChooseOneExcluding(ctx context.Context, exclude []string) (CommonFileServerDTO, error) {
	whereCluse := ""
	params := make([]interface{}, len(exclude))

	if len(exclude) != 0 {
		whereCluse = "WHERE id NOT IN (?" + strings.Repeat(",?", len(exclude)-1) + ")"

		for i, val := range exclude {
			params[i] = val
		}
	}

	query := "SELECT id, name, type, params FROM file_server " + whereCluse + " LIMIT 1"

	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return CommonFileServerDTO{}, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	entity := CommonFileServerDTO{}
	err = stmt.QueryRowContext(ctx, params...).Scan(&entity.ID, &entity.Name, &entity.Type, &entity.Params)

	switch {
	case err == sql.ErrNoRows:
		return CommonFileServerDTO{}, ErrNotFound
	case err != nil:
		return CommonFileServerDTO{}, err
	}

	return entity, nil
}

func (s *FileServerStorage) Count(ctx context.Context) (int, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT count(*) FROM file_server")
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	var res int
	err = stmt.QueryRowContext(ctx).Scan(&res)

	switch {
	case err == sql.ErrNoRows:
		return 0, ErrNotFound
	case err != nil:
		return 0, err
	}

	return res, nil
}
