package file_server_model

import "time"

// SSHFileServer represents file server working via SSH.
type SSHFileServer struct {
	ID         string    `json:"id,omitempty"`
	Name       string    `json:"name,omitempty"`
	Host       string    `json:"host,omitempty"`
	Port       string    `json:"port,omitempty"`
	BasePath   string    `json:"base_path,omitempty"`
	User       string    `json:"user,omitempty"`
	Key        string    `json:"key,omitempty"`
	TotalSpace int64     `json:"total_space,omitempty"`
	UsedSpace  int64     `json:"used_space"`
	Status     Status    `json:"status,omitempty"`
	Created    time.Time `json:"created,omitempty"`
	Modified   time.Time `json:"modified,omitempty"`
}

func (s *SSHFileServer) HideCredentials() {
	if s.Key != "" {
		s.Key = "***"
	}
}

func (s *SSHFileServer) GetID() string {
	return s.ID
}

func (s *SSHFileServer) GetFreeSpace() int64 {
	return s.TotalSpace - s.UsedSpace
}
