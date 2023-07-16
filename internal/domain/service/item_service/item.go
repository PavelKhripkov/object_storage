package item_service

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/item_model"
	"github.com/gofrs/uuid/v5"
	log "github.com/sirupsen/logrus"
	"time"
)

// Service provides methods to manage items.
type Service struct {
	storage itemStorage
	l       *log.Entry
}

// NewItemService creates new item service.
func NewItemService(itemStorage itemStorage, l *log.Logger) *Service {
	return &Service{
		storage: itemStorage,
		l:       l.WithField("component", "ItemService"),
	}
}

// Get returns item model by ID.
func (s Service) Get(ctx context.Context, id string) (item_model.Item, error) {
	res, err := s.storage.Get(ctx, id)
	if err != nil {
		return item_model.Item{}, err
	}

	return res, nil
}

// Create creates and returns new item model.
func (s Service) Create(ctx context.Context, dto CreateItemDTO) (item_model.Item, error) {
	newID, err := uuid.NewV7()
	if err != nil {
		return item_model.Item{}, err
	}

	now := time.Now()

	newItem := item_model.Item{
		ID:          newID.String(),
		Name:        dto.Name,
		Size:        dto.Size,
		ContainerID: dto.ContainerID,
		Status:      item_model.ItemStatusPending,
		Created:     now,
		Modified:    now,
	}

	err = s.storage.Create(ctx, newItem)
	if err != nil {
		return item_model.Item{}, err
	}

	return newItem, nil
}

// List return all item entities.
func (s Service) List(ctx context.Context, containerID string) ([]item_model.Item, error) {
	items, err := s.storage.List(ctx, containerID)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// Update updates specified fields of an item.
func (s Service) Update(ctx context.Context, itm item_model.Item, params UpdateItemDTO) (item_model.Item, error) {
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
		return item_model.Item{}, err
	}

	return itm, nil
}

func (s Service) Delete(ctx context.Context, id string) error {
	// TODO implement.
	return nil
}
