package chunk_model

import "time"

type Chunk struct {
	ID           string    `json:"id,omitempty"`
	ItemID       string    `json:"item_id,omitempty"`
	Position     uint8     `json:"position,omitempty"`
	FileServerID string    `json:"file_server_id,omitempty"`
	FilePath     string    `json:"file_path,omitempty"`
	Size         int64     `json:"size,omitempty"`
	Created      time.Time `json:"created,omitempty"`
	Modified     time.Time `json:"modified,omitempty"`
}
