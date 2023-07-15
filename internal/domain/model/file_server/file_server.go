package file_server

type FileServer interface {
	HideCredentials()
	GetID() string
}
