package item_usecase

import (
	"github.com/PavelKhripkov/object_storage/internal/domain/service/file_server_service"
)

type StoreItemDTO struct {
	F           file_server_service.Opener
	Name        string
	ContainerID string
	Size        int64
}
