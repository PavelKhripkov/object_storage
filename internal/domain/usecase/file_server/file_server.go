package file_server_usecase

import (
	"context"
	file_server2 "github.com/PavelKhripkov/object_storage/internal/domain/model/file_server"
	"github.com/PavelKhripkov/object_storage/internal/domain/service/file_server"
	log "github.com/sirupsen/logrus"
)

type Usecase struct {
	fileServerService *file_server.Service
	l                 *log.Entry
}

func NewFileServerUsecase(fileServerService *file_server.Service, l *log.Logger) *Usecase {
	return &Usecase{
		fileServerService: fileServerService,
		l:                 l.WithField("component", "FileServerUsecase"),
	}
}

func (s Usecase) Add(ctx context.Context, dto file_server.AddFileServerDTO) (file_server2.FileServer, error) {
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	res, err := s.fileServerService.Add(ctx, dto)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s Usecase) Get(ctx context.Context, id string) (file_server2.FileServer, error) {
	res, err := s.fileServerService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return res, nil
}
