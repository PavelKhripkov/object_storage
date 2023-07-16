package file_server_service

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/adapter/db/sqlite"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/file_server_model"
	"mime/multipart"
)

//type File interface {
//	io.Reader
//	io.ReaderAt
//	io.Seeker
//	io.Closer
//}

type Opener interface {
	Open() (multipart.File, error)
}

type fileServerStorage interface {
	Add(ctx context.Context, params sqlite.CommonFileServerDTO) error
	Get(ctx context.Context, id string) (sqlite.CommonFileServerDTO, error)
	ChooseOneExcluding(ctx context.Context, exclude []string) (sqlite.CommonFileServerDTO, error)
	Count(ctx context.Context) (int, error)
	UpdateUsedSpace(ctx context.Context, id string, change int64) error
	UpdateStatus(ctx context.Context, id string, status file_server_model.Status) error
}
