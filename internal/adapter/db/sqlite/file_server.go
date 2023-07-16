package sqlite

import (
	"context"
	"database/sql"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/file_server_model"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
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

func (s *FileServerStorage) Add(ctx context.Context, prm CommonFileServerDTO) error {
	stmt, err := s.db.PrepareContext(
		ctx,
		"INSERT INTO file_server (id, name, type, params, total_space, status, created, modified) values (?, ?, ?, ?, ?, ?, ?, ?)",
	)
	if err != nil {
		return err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	_, err = stmt.ExecContext(ctx, prm.ID, prm.Name, prm.Type, prm.Params, prm.TotalSpace, prm.Status, prm.Created.UnixMilli(), prm.Modified.UnixMilli())
	if err != nil {
		return err
	}

	return nil
}

func (s *FileServerStorage) Get(ctx context.Context, id string) (CommonFileServerDTO, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT id, name, type, params, total_space, used_space, status, created, modified FROM file_server WHERE id = ? LIMIT 1")
	if err != nil {
		return CommonFileServerDTO{}, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	entity := CommonFileServerDTO{}
	var created, modified int64

	err = stmt.QueryRowContext(ctx, id).Scan(&entity.ID, &entity.Name, &entity.Type, &entity.Params, &entity.TotalSpace, &entity.UsedSpace, &entity.Status, &created, &modified)

	switch {
	case err == sql.ErrNoRows:
		return CommonFileServerDTO{}, ErrNotFound
	case err != nil:
		return CommonFileServerDTO{}, err
	}

	entity.Created = time.UnixMilli(created)
	entity.Modified = time.UnixMilli(modified)

	return entity, nil
}

func (s *FileServerStorage) ChooseOneExcluding(ctx context.Context, exclude []string) (CommonFileServerDTO, error) {
	whereClauseExclude := ""
	params := make([]interface{}, len(exclude))

	if len(exclude) != 0 {
		whereClauseExclude = "AND id NOT IN (?" + strings.Repeat(",?", len(exclude)-1) + ")"

		for i, val := range exclude {
			params[i] = val
		}
	}

	query := "SELECT id, name, type, total_space, used_space, status, params, created, modified FROM file_server WHERE status = 'ok' " + whereClauseExclude +
		" ORDER BY total_space - used_space DESC LIMIT 1"

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
	var created, modified int64

	err = stmt.QueryRowContext(ctx, params...).
		Scan(&entity.ID, &entity.Name, &entity.Type, &entity.TotalSpace, &entity.UsedSpace, &entity.Status, &entity.Params, &created, &modified)

	switch {
	case err == sql.ErrNoRows:
		return CommonFileServerDTO{}, ErrNotFound
	case err != nil:
		return CommonFileServerDTO{}, err
	}

	entity.Created = time.UnixMilli(created)
	entity.Modified = time.UnixMilli(modified)

	return entity, nil
}

func (s *FileServerStorage) Count(ctx context.Context) (int, error) {
	stmt, err := s.db.PrepareContext(ctx, "SELECT count(*) FROM file_server WHERE status = 'ok'")
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

//func (s *FileServerStorage) Update(ctx context.Context, entity CommonFileServerDTO) error {
//	stmt, err := s.db.PrepareContext(ctx, "UPDATE file_server SET name=?, params=?, total_space=?, status=?, modified=? WHERE id = ?")
//	if err != nil {
//		return err
//	}
//	defer func() {
//		if err := stmt.Close(); err != nil {
//			s.l.Error(err)
//		}
//	}()
//
//	modified := time.Now().UnixMilli()
//
//	_, err = stmt.ExecContext(ctx, entity.Name, entity.Params, entity.TotalSpace, entity.Status, modified, entity.ID)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func (s *FileServerStorage) UpdateUsedSpace(ctx context.Context, id string, change int64) error {
	var used int64

	// TODO implement better retry for serializable tx.
	for {
		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return err
		}

		err = tx.QueryRowContext(ctx, "SELECT used_space FROM file_server WHERE id = ?", id).Scan(&used)
		if err != nil {
			return err
		}

		used += change

		_, execErr := tx.ExecContext(ctx, "UPDATE file_server SET used_space = ? WHERE id = ?", used, id)

		if execErr != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.l.Debugf("Update failed: %v, unable to rollback: %v.", execErr, rollbackErr)
				continue
			}
			s.l.Debugf("Update failed: %v.", execErr)
			continue
		}

		if err = tx.Commit(); err != nil {
			s.l.Debug(err)
			continue
		}

		return nil
	}

}

func (s *FileServerStorage) UpdateStatus(ctx context.Context, id string, status file_server_model.Status) error {
	stmt, err := s.db.PrepareContext(ctx, "UPDATE file_server SET status=?, modified=? WHERE id = ?")
	if err != nil {
		return err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	modified := time.Now().UnixMilli()

	_, err = stmt.ExecContext(ctx, status, modified, id)
	if err != nil {
		return err
	}

	return nil
}
