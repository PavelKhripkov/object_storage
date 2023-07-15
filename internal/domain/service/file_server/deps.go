package file_server

import (
	"context"
	"github.com/PavelKhripkov/object_storage/internal/adapter/db/sqlite"
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
}
