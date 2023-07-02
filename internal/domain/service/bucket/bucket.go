package bucket

import (
	"context"
	"github.com/PavelKhripkov/best_blob_storage/internal/model/item"
)

type Service struct {
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
