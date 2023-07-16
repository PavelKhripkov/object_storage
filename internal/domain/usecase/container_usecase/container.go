package container_usecase

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/container_model"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/container_service"
	log "github.com/sirupsen/logrus"
)

// Usecase represents container use cases.
type Usecase struct {
	containerService *container_service.Service
	l                *log.Entry
}

// NewContainerUsecase creates new container use cases service.
func NewContainerUsecase(containerService *container_service.Service, l *log.Logger) *Usecase {
	return &Usecase{
		containerService: containerService,
		l:                l.WithField("component", "ContainerUsecase"),
	}
}

func (s *Usecase) Get(ctx context.Context, id string) (container_model.Container, error) {
	res, err := s.containerService.Get(ctx, id)
	if err != nil {
		return container_model.Container{}, err
	}

	return res, nil
}

func (s *Usecase) Create(ctx context.Context, dto CreateContainerDTO) (container_model.Container, error) {
	params := container_service.CreateContainerDTO{
		Name:        dto.Name,
		Description: dto.Description,
		ParentID:    dto.ParentID,
	}

	entity, err := s.containerService.Create(ctx, params)
	if err != nil {
		return container_model.Container{}, err
	}

	return entity, nil
}

func (s *Usecase) List(ctx context.Context) ([]container_model.Container, error) {
	res, err := s.containerService.List(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}
