package item_usecase

import (
	"github.com/PavelKhripkov/object_storage/internal/domain/service/file_server"
)

type StoreItemDTO struct {
	F        file_server.Opener
	Name     string
	BucketID string
	Path     string
	Size     int64
}
