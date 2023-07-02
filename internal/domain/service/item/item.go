package item

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/item"
	"io"
)

type Service struct {
}

func (s Service) Get(ctx context.Context, id string) ([]byte, error) {
	return nil, nil
}

func (s Service) List(ctx context.Context) ([]item.Item, error) {
	return nil, nil
}

func (s Service) Put(ctx context.Context, reader io.Reader) error {
	return nil
}

func (s Service) Delete(ctx context.Context, id string) error {
	return nil
}
