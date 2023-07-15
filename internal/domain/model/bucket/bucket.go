package bucket

import "time"

type Bucket struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Modified    time.Time `json:"modified,omitempty"`
}
