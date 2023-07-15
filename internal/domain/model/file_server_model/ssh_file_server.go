package file_server_model

import "time"

type SSHFileServer struct {
	ID       string    `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Address  string    `json:"address,omitempty"`
	Port     string    `json:"port,omitempty"`
	BasePath string    `json:"base_path,omitempty"`
	User     string    `json:"user,omitempty"`
	Key      string    `json:"key,omitempty"`
	Created  time.Time `json:"created,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
}

func (s *SSHFileServer) HideCredentials() {
	if s.Key != "" {
		s.Key = "***"
	}
}

func (s *SSHFileServer) GetID() string {
	return s.ID
}
