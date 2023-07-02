package item_usecase

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/item_data"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/file_server"
)

type fileService interface {
	ID() string
	Status(ctx context.Context, id string) (bool, error)
	List(ctx context.Context) ([]fileService, error)
	ChooseOne(ctx context.Context, exclude map[string]fileService) (fileService, error)
	Store(ctx context.Context, file file_server.File, start, limit int64) error
	GetPart(ctx context.Context, part item_data.Part) ([]byte, error)
}
