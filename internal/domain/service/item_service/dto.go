package item_service

import "github.com/PavelKhripkov/object_storage/internal/domain/model/item_model"

type CreateItemDTO struct {
	Name        string
	ContainerID string
	Size        int64
	ChunkCount  int8
}

type ChangeItemDTO struct {
	Status     *item_model.Status
	ChunkCount *uint8
}
