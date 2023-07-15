package item

import "github.com/PavelKhripkov/object_storage/internal/domain/model/item"

type CreateItemDTO struct {
	Name       string
	Path       string
	Size       int64
	BucketId   string
	ChunkCount int8
}

type ChangeItemDTO struct {
	Status     *item.Status
	ChunkCount *uint8
}
