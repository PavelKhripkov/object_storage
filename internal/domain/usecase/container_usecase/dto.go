package container_usecase

type CreateContainerDTO struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
}
