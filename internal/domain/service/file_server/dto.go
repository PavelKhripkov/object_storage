package file_server

import (
	"encoding/json"
	"os"
)

type AddFileServerDTO interface {
	Validate() error
	GetName() string
	GetType() string
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

func (s AddAPIFileServerDTO) MarshalParams() (string, error) {
	// Using copy of the object
	s.Name = ""
	res, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

type AddSSHFileServerDTO struct {
	Name     string `json:"name,omitempty"`
	Address  string `json:"address,omitempty"`
	Port     string `json:"port,omitempty"`
	BasePath string `json:"base_path,omitempty"`
	User     string `json:"user,omitempty"`
	KeyFile  string `json:"key_file,omitempty"`
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

func (s AddSSHFileServerDTO) MarshalParams() (string, error) {
	key, err := os.ReadFile(s.KeyFile)
	if err != nil {
		return "", err
	}

	temp := struct {
		Name     string `json:"name,omitempty"`
		Address  string `json:"address,omitempty"`
		Port     string `json:"port,omitempty"`
		BasePath string `json:"base_path,omitempty"`
		User     string `json:"user,omitempty"`
		Key      string `json:"key,omitempty"`
	}{
		s.Name,
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
