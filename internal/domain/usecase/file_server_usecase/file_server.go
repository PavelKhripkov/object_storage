package file_server_usecase

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/file_server_model"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/file_server_service"
	log "github.com/sirupsen/logrus"
)

type Usecase struct {
	fileServerService *file_server_service.Service
	l                 *log.Entry
}

func NewFileServerUsecase(fileServerService *file_server_service.Service, l *log.Logger) *Usecase {
	return &Usecase{
		fileServerService: fileServerService,
		l:                 l.WithField("component", "FileServerUsecase"),
	}
}

func (s Usecase) Add(ctx context.Context, dto file_server_service.AddFileServerDTO) (file_server_model.FileServer, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	res, err := s.fileServerService.Add(ctx, dto)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s Usecase) Get(ctx context.Context, id string) (file_server_model.FileServer, error) {
	res, err := s.fileServerService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return res, nil
}
