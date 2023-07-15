package container_service

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/container_model"
)

type containerStorage interface {
	Get(ctx context.Context, id string) (container_model.Container, error)
	List(ctx context.Context) ([]container_model.Container, error)
	Create(ctx context.Context, container container_model.Container) error
	Delete(ctx context.Context, id string) error
}
