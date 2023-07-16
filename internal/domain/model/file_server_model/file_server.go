package file_server_model

// FileServer specifies methods of file servers.
type FileServer interface {
	// HideCredentials replaces secretes with asterisks.
	HideCredentials()
	// GetID returns file server ID.
	GetID() string
	// GetFreeSpace returns file server space available for storing files.
	GetFreeSpace() int64
}

type Status string

const (
	FileServerStatusOK      Status = "ok"
	FileServerStatusFail    Status = "fail"
	FileServerStatusUnknown Status = "unknown"
)
