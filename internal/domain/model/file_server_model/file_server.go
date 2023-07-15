package file_server_model

type FileServer interface {
	HideCredentials()
	GetID() string
}
