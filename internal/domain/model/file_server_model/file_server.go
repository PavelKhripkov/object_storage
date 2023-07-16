package file_server_model

type FileServer interface {
	HideCredentials()
	GetID() string
	GetFreeSpace() int64
}

type Status string

const (
	FileServerStatusOK      Status = "ok"
	FileServerStatusFail    Status = "fail"
	FileServerStatusUnknown Status = "unknown"
)
