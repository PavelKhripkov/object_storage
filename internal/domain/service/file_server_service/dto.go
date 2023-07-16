package file_server_service

import (
	"encoding/json"
	"github.com/PavelKhripkov/object_storage/internal/domain/model/file_server_model"
	"os"
)

type AddFileServerDTO interface {
	Validate() error
	GetName() string
	GetType() string
	GetTotalSpace() int64
	MarshalParams() (string, error)
}

type AddAPIFileServerDTO struct {
	Name       string `json:"name,omitempty"`
	Address    string `json:"address,omitempty"`
	Port       string `json:"port,omitempty"`
	Endpoint   string `json:"endpoint,omitempty"`
	APIVersion string `json:"api_version,omitempty"`
	User       string `json:"user,omitempty"`
	Password   string `json:"password,omitempty"`
	TotalSpace int64  `json:"total_space,omitempty"`
}

func (s AddAPIFileServerDTO) Validate() error {
	// TODO validation
	return nil
}

func (s AddAPIFileServerDTO) GetName() string {
	return s.Name
}

func (s AddAPIFileServerDTO) GetType() string {
	return "api"
}

func (s AddAPIFileServerDTO) GetTotalSpace() int64 {
	return s.TotalSpace
}

func (s AddAPIFileServerDTO) MarshalParams() (string, error) {
	// Using copy of the object, so we can change field.
	s.Name = ""
	res, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

type AddSSHFileServerDTO struct {
	Name       string `json:"name,omitempty"`
	Address    string `json:"address,omitempty"`
	Port       string `json:"port,omitempty"`
	BasePath   string `json:"base_path,omitempty"`
	User       string `json:"user,omitempty"`
	KeyFile    string `json:"key_file,omitempty"`
	TotalSpace int64  `json:"total_space,omitempty"`
}

func (s AddSSHFileServerDTO) Validate() error {
	// TODO validation
	return nil
}

func (s AddSSHFileServerDTO) GetName() string {
	return s.Name
}

func (s AddSSHFileServerDTO) GetType() string {
	return "ssh"
}

func (s AddSSHFileServerDTO) GetTotalSpace() int64 {
	return s.TotalSpace
}

func (s AddSSHFileServerDTO) MarshalParams() (string, error) {
	key, err := os.ReadFile(s.KeyFile)
	if err != nil {
		return "", err
	}

	temp := struct {
		Address  string `json:"address,omitempty"`
		Port     string `json:"port,omitempty"`
		BasePath string `json:"base_path,omitempty"`
		User     string `json:"user,omitempty"`
		Key      string `json:"key,omitempty"`
	}{
		s.Address,
		s.Port,
		s.BasePath,
		s.User,
		string(key),
	}

	res, err := json.Marshal(temp)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

type UpdateFileServerDTO struct {
	Status *file_server_model.Status
}
