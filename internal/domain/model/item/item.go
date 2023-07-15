package item

import "time"

type Status string

func (s Status) Pointer() *Status {
	return &s
}

const (
	ItemStatusOK      Status = "ok"
	ItemStatusFail    Status = "fail"
	ItemStatusPending Status = "pending"
)

type Item struct {
	ID         string    `json:"id,omitempty"`
	Name       string    `json:"name,omitempty"`
	Size       int64     `json:"size,omitempty"`
	Path       string    `json:"path,omitempty"`
	BucketId   string    `json:"bucket_id,omitempty"`
	ChunkCount uint8     `json:"chunk_count,omitempty"`
	Status     Status    `json:"status,omitempty"`
	Created    time.Time `json:"created,omitempty"`
	Modified   time.Time `json:"modified,omitempty"`
}
