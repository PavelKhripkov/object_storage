package container_model

import "time"

// Container represents container for items (e.g. folder).
type Container struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	ParentID    string    `json:"parent_id,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Modified    time.Time `json:"modified,omitempty"`
}
