package sqlite

import (
	"context"
	"database/sql"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/bucket"
	log "github.com/sirupsen/logrus"
)

type BucketStorage struct {
	db *sql.DB
	l  *log.Entry
}

func NewBucketStorage(db *sql.DB, l *log.Logger) *BucketStorage {
	return &BucketStorage{
		db: db,
		l:  l.WithField("component", "BucketStorage"),
	}
}

func (s *BucketStorage) Get(ctx context.Context, id string) (bucket.Bucket, error) {
	return bucket.Bucket{}, nil
}

func (s *BucketStorage) Create(ctx context.Context, bucket bucket.Bucket) error {
	return nil
}

func (s *BucketStorage) Delete(ctx context.Context, bucket bucket.Bucket) error {
	return nil
}
