package container_service

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/container_model"
	"github.com/gofrs/uuid/v5"
	log "github.com/sirupsen/logrus"
	"time"
)

// Service provides methods to manage containers.
type Service struct {
	storage containerStorage
	l       *log.Entry
}

// NewContainerService creates new container service.
func NewContainerService(containerStorage containerStorage, l *log.Logger) *Service {
	return &Service{
		storage: containerStorage,
		l:       l.WithField("component", "ContainerService"),
	}
}

// Get returns container model by ID.
func (s Service) Get(ctx context.Context, id string) (container_model.Container, error) {
	res, err := s.storage.Get(ctx, id)
	if err != nil {
		return container_model.Container{}, err
	}

	return res, nil
}

// Create creates and returns new container model.
func (s Service) Create(ctx context.Context, dto CreateContainerDTO) (container_model.Container, error) {
	newID, err := uuid.NewV7()
	if err != nil {
		return container_model.Container{}, err
	}

	now := time.Now()

	newContainer := container_model.Container{
		ID:          newID.String(),
		Name:        dto.Name,
		Description: dto.Description,
		ParentID:    dto.ParentID,
		Created:     now,
		Modified:    now,
	}

	err = s.storage.Create(ctx, newContainer)
	if err != nil {
		return container_model.Container{}, err
	}

	return newContainer, nil
}

// List returns all available containers.
func (s Service) List(ctx context.Context) ([]container_model.Container, error) {
	containers, err := s.storage.List(ctx)
	if err != nil {
		return nil, err
	}

	return containers, nil
}

// Delete removes container by ID.
func (s Service) Delete(ctx context.Context, id string) error {
	return s.storage.Delete(ctx, id)
}
