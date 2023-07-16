package file_server_model

import "time"

type APIFileServer struct {
	ID         string    `json:"id,omitempty"`
	Name       string    `json:"name,omitempty"`
	Host       string    `json:"host,omitempty"`
	Port       string    `json:"port,omitempty"`
	Endpoint   string    `json:"endpoint,omitempty"`
	APIVersion string    `json:"api_ersion,omitempty"`
	User       string    `json:"user,omitempty"`
	Password   string    `json:"password,omitempty"`
	TotalSpace int64     `json:"total_space,omitempty"`
	UsedSpace  int64     `json:"used_space"`
	Status     Status    `json:"status,omitempty"`
	Created    time.Time `json:"created,omitempty"`
	Modified   time.Time `json:"modified,omitempty"`
}

func (s *APIFileServer) HideCredentials() {
	if s.Password != "" {
		s.Password = "***"
	}
}

func (s *APIFileServer) GetID() string {
	return s.ID
}

func (s *APIFileServer) GetFreeSpace() int64 {
	return s.TotalSpace - s.UsedSpace
}
