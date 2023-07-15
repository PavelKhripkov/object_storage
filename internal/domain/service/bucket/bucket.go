package bucket

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/item"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	storage bucketStorage
	l       *log.Entry
}

func NewBucketService(bucketStorage bucketStorage, l *log.Logger) *Service {
	return &Service{
		storage: bucketStorage,
		l:       l.WithField("component", "BucketService"),
	}
}

func (s Service) Create(ctx context.Context, name string) (string, error) {
	return "", nil
}

func (s Service) List(ctx context.Context) ([]item.Item, error) {
	return nil, nil
}

func (s Service) Delete(ctx context.Context, id string) error {
	return nil
}
