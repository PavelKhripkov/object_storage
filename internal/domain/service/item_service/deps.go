package item_service

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/item_model"
)

type itemStorage interface {
	Get(ctx context.Context, id string) (item_model.Item, error)
	List(ctx context.Context, containerID string) ([]item_model.Item, error)
	Create(ctx context.Context, item item_model.Item) error
	Update(ctx context.Context, item item_model.Item) error
	Delete(ctx context.Context, item *item_model.Item) error
}
