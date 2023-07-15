package item

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/item"
	"github.com/gofrs/uuid/v5"
	log "github.com/sirupsen/logrus"
	"time"
)

type Service struct {
	storage itemStorage
	l       *log.Entry
}

func NewItemService(itemStorage itemStorage, l *log.Logger) *Service {
	return &Service{
		storage: itemStorage,
		l:       l.WithField("component", "ItemService"),
	}
}

func (s Service) Get(ctx context.Context, id string) (item.Item, error) {
	res, err := s.storage.Get(ctx, id)
	if err != nil {
		return item.Item{}, err
	}

	return *res, nil
}

func (s Service) Create(ctx context.Context, dto CreateItemDTO) (item.Item, error) {
	newID, err := uuid.NewV7()
	if err != nil {
		return item.Item{}, err
	}

	now := time.Now()

	newItem := item.Item{
		ID:       newID.String(),
		Name:     dto.Name,
		Size:     dto.Size,
		Path:     dto.Path,
		BucketId: dto.BucketId,
		Status:   item.ItemStatusPending,
		Created:  now,
		Modified: now,
	}

	err = s.storage.Create(ctx, newItem)
	if err != nil {
		return item.Item{}, err
	}

	return newItem, nil
}

func (s Service) List(ctx context.Context) ([]item.Item, error) {
	return nil, nil
}

func (s Service) Update(ctx context.Context, itm item.Item, params ChangeItemDTO) (item.Item, error) {
	var isChanged bool

	if params.Status != nil {
		isChanged = true
		itm.Status = *params.Status
	}

	if params.ChunkCount != nil {
		isChanged = true
		itm.ChunkCount = *params.ChunkCount
	}

	if !isChanged {
		return itm, nil
	}

	err := s.storage.Update(ctx, itm)
	if err != nil {
		return item.Item{}, err
	}

	return itm, nil
}

func (s Service) Delete(ctx context.Context, id string) error {
	return nil
}
