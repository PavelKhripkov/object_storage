package item

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/item"
)

type itemStorage interface {
	Get(ctx context.Context, id string) (*item.Item, error)
	GetAll(ctx context.Context) []*item.Item
	Create(ctx context.Context, item item.Item) error
	Update(ctx context.Context, item item.Item) error
	Delete(ctx context.Context, item *item.Item) error
}
