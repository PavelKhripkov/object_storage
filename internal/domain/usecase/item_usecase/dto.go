package item_usecase

import (
	"mime/multipart"
)

type StoreItemDTO struct {
	F           *multipart.FileHeader
	Name        string
	ContainerID string
	Size        int64
	Close       func() error
}
