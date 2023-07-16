package sqlite

import (
	"github.com/PavelKhripkov/object_storage/internal/domain/model/file_server_model"
	"time"
)

type CommonFileServerDTO struct {
	ID         string                   `json:"id,omitempty"`
	Name       string                   `json:"name,omitempty"`
	Type       string                   `json:"type,omitempty"`
	TotalSpace int64                    `json:"total_space,omitempty"`
	UsedSpace  int64                    `json:"used_space,omitempty"`
	Status     file_server_model.Status `json:"status,omitempty"`
	Params     string                   `json:"params,omitempty"`
	Created    time.Time                `json:"created,omitempty"`
	Modified   time.Time                `json:"modified,omitempty"`
}

func (s CommonFileServerDTO) Validate() error {
	// TODO validate
	return nil
}
