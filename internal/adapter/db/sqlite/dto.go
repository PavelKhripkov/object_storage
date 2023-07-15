package sqlite

import "time"

type CommonFileServerDTO struct {
	ID       string    `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Type     string    `json:"type,omitempty"`
	Params   string    `json:"params,omitempty"`
	Created  time.Time `json:"created,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
}

func (s CommonFileServerDTO) Validate() error {
	return nil
}
